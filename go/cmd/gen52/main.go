// Command gen52 regenerates the lookup tables for a 52-card deck.
//
// It is resumable at every expensive stage:
//   - 7-card keys are enumerated once and split into fixed chunks; each chunk is
//     sorted and written independently (interrupted chunks are redone, finished
//     chunks are skipped).
//   - the k-way merge into texas_data.txt checkpoints every few million rows.
//   - the compact color/normal tables checkpoint their input/output offsets and
//     dedup key sets.
//   - the N-card win-rate tables are written in independent parts.
//   - every final file is written to a .tmp path and renamed only when complete.
//
// The generator reuses the old 54-card compact tables (kept in -refdir) as a
// comparator. Regular 7-card keys are isolated from joker rows in those tables
// because RemoveColor/ChangeColor keep the joker byte intact, so the old
// Postion/Max/Type values are valid for every 52-card input.
package main

import (
	"bufio"
	"container/heap"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	ta "github.com/esrrhs/texas_algorithm/go"
)

const (
	chunkSize7     = 16_000_000
	mergeCkptRows  = 5_000_000
	opt7CkptRows   = 20_000_000
	mergeReadBlock = 1 << 20 // int64 per read block
)

var (
	workDir = flag.String("workdir", "/tmp/ta_gen", "work directory")
	refDir  = flag.String("refdir", "/tmp/ta_gen/ref", "directory with the old 54-card tables")
	mode    = flag.String("mode", "all", "all|chunks7|merge7|opt7|extra|extraopt|trans|transopt|status|smoke")
	samples = flag.Int("samples", 20000, "smoke test sample count")
)

func main() {
	flag.Parse()
	if err := os.MkdirAll(*workDir, 0o755); err != nil {
		fatal(err)
	}

	switch *mode {
	case "status":
		printStatus()
		return
	case "smoke":
		must(phaseSmoke())
		return
	case "chunks7":
		loadRefTables()
		must(phaseChunks7())
	case "merge7":
		loadRefTables()
		must(phaseMerge7())
	case "opt7":
		must(phaseOpt7())
	case "extra":
		must(phaseExtra(6))
		must(phaseExtra(5))
	case "extraopt":
		must(phaseExtraOpt(6))
		must(phaseExtraOpt(5))
	case "fixextra":
		must(phaseFixExtra())
	case "trans":
		loadNewTables()
		for n := 6; n >= 2; n-- {
			must(phaseTrans(n))
		}
	case "transopt":
		for n := 6; n >= 2; n-- {
			must(phaseTransOpt(n))
		}
	case "all":
		loadRefTables()
		must(phaseChunks7())
		must(phaseMerge7())
		must(phaseOpt7())
		must(phaseExtra(6))
		must(phaseExtra(5))
		must(phaseExtraOpt(6))
		must(phaseExtraOpt(5))
		must(phaseFixExtra())
		loadNewTables()
		for n := 6; n >= 2; n-- {
			must(phaseTrans(n))
		}
		for n := 6; n >= 2; n-- {
			must(phaseTransOpt(n))
		}
		printStatus()
	default:
		fatal(fmt.Errorf("unknown mode %q", *mode))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
	os.Exit(1)
}

func must(err error) {
	if err != nil {
		fatal(err)
	}
}

func logf(format string, args ...interface{}) {
	fmt.Printf("[%s] ", time.Now().Format("15:04:05"))
	fmt.Printf(format, args...)
	fmt.Println()
}

func path(name string) string { return filepath.Join(*workDir, name) }

func exists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func atomicJSON(p string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func readJSON(p string, v interface{}) bool {
	data, err := os.ReadFile(p)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, v) == nil
}

func comb(n, k int) int64 {
	if k > n {
		return 0
	}
	if k > n-k {
		k = n - k
	}
	r := int64(1)
	for i := 1; i <= k; i++ {
		r = r * int64(n-k+i) / int64(i)
	}
	return r
}

func enumKeys(n int) []int64 {
	deck := ta.GenAllCards()
	keys := make([]int64, 0, comb(len(deck), n))
	tmp := make([]int, n)
	ta.Permutation(deck, 0, 0, n, tmp, func(t []int) {
		keys = append(keys, ta.GenCardBindInts(t))
	})
	return keys
}

func loadRefTables() {
	logf("load reference tables from %s", *refDir)
	ta.ColorMap = map[int64]*ta.KeyData{}
	ta.NormalMap = map[int64]*ta.KeyData{}
	ta.LoadDir(*refDir)
	ta.UseOpt = true
	if !ta.IsLoaded() {
		fatal(fmt.Errorf("reference tables not loaded from %s", *refDir))
	}
}

func loadNewTables() {
	logf("load new tables from %s", *workDir)
	ta.ColorMap = map[int64]*ta.KeyData{}
	ta.NormalMap = map[int64]*ta.KeyData{}
	ta.LoadDir(*workDir)
	if !ta.IsLoaded() {
		fatal(fmt.Errorf("new tables not loaded from %s", *workDir))
	}
}

// ---------------------------------------------------------------------------
// Smoke test: verify that the old 54-card tables give correct answers for
// 52-card keys (the joker rows must not leak into regular lookups).
// ---------------------------------------------------------------------------

func phaseSmoke() error {
	loadRefTables()
	deck := ta.GenAllCards()
	n := 7
	rng := uint64(0x9e3779b97f4a7c15)
	randIndex := func(m int) int {
		rng ^= rng << 13
		rng ^= rng >> 7
		rng ^= rng << 17
		return int(rng % uint64(m))
	}
	cards := make([]int, n)
	for s := 0; s < *samples; s++ {
		seen := map[int]bool{}
		for i := 0; i < n; i++ {
			for {
				j := randIndex(len(deck))
				if !seen[j] {
					seen[j] = true
					cards[i] = deck[j]
					break
				}
			}
		}
		sort.Ints(cards)
		k := ta.GenCardBindInts(cards)
		kd := ta.LookupKeyData(k)
		if kd == nil {
			return fmt.Errorf("smoke: nil key data for %d", k)
		}
		if got, want := kd.Max, ta.Max(k); !ta.Equal(got, want) {
			return fmt.Errorf("smoke: key %d Max=%d strength-differs from brute force=%d", k, got, want)
		}
		if got, want := kd.Type, ta.MaxType(k); got != want {
			return fmt.Errorf("smoke: key %d Type=%d, brute force=%d", k, got, want)
		}
	}
	logf("smoke ok: %d random 7-card keys match brute force", *samples)
	return nil
}

// ---------------------------------------------------------------------------
// Phase 1: enumerate and sort 7-card keys in chunks.
// ---------------------------------------------------------------------------

func chunk7Path(i int) string {
	return path(filepath.Join("chunks7", fmt.Sprintf("chunk_%03d.bin", i)))
}

func phaseChunks7() error {
	if exists(path("texas_data.txt")) {
		logf("texas_data.txt exists, skip chunks7")
		return nil
	}
	dir := path("chunks7")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	nChunks := (int(ta.Total) + chunkSize7 - 1) / chunkSize7
	logf("chunks7: %d chunks of %d", nChunks, chunkSize7)

	logf("enumerating 7-card keys ...")
	keys := enumKeys(7)
	if len(keys) != int(ta.Total) {
		return fmt.Errorf("enumerated %d keys, want %d", len(keys), ta.Total)
	}
	logf("enumerated %d keys", len(keys))

	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())
	var finished atomic.Int64
	var failure atomic.Value

	startAll := time.Now()
	for ci := 0; ci < nChunks; ci++ {
		start := ci * chunkSize7
		end := start + chunkSize7
		if end > len(keys) {
			end = len(keys)
		}
		p := chunk7Path(ci)
		if exists(p) {
			finished.Add(1)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(ci, start, end int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					failure.Store(fmt.Errorf("chunk %d panic: %v", ci, r))
				}
			}()

			t0 := time.Now()
			sort.Slice(keys[start:end], func(i, j int) bool {
				return ta.GenCompare(keys[start+i], keys[start+j])
			})
			if err := writeChunk(p, keys[start:end]); err != nil {
				failure.Store(err)
				return
			}
			n := finished.Add(1)
			done := n
			logf("chunk %d/%d done, %d keys, %s, elapsed %s, eta %s",
				ci+1, nChunks, end-start, time.Since(t0).Round(time.Second),
				time.Since(startAll).Round(time.Second),
				eta(startAll, done, int64(nChunks)))
		}(ci, start, end, p)
	}
	wg.Wait()
	if v := failure.Load(); v != nil {
		return v.(error)
	}
	if finished.Load() != int64(nChunks) {
		return fmt.Errorf("only %d/%d chunks finished", finished.Load(), nChunks)
	}
	logf("chunks7 done in %s", time.Since(startAll).Round(time.Second))
	return nil
}

func writeChunk(p string, keys []int64) error {
	tmp := p + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriterSize(f, 4<<20)
	const block = 1 << 20
	buf := make([]byte, 8*block)
	for off := 0; off < len(keys); off += block {
		n := block
		if off+n > len(keys) {
			n = len(keys) - off
		}
		for i := 0; i < n; i++ {
			binary.LittleEndian.PutUint64(buf[i*8:], uint64(keys[off+i]))
		}
		if _, err := w.Write(buf[:n*8]); err != nil {
			f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func eta(start time.Time, done, total int64) string {
	if done == 0 {
		return "?"
	}
	per := time.Since(start) / time.Duration(done)
	return (per * time.Duration(total-done)).Round(time.Second).String()
}

// ---------------------------------------------------------------------------
// Phase 2: k-way merge of sorted chunks into texas_data.txt.
// ---------------------------------------------------------------------------

type mergeState struct {
	Consumed    []int64 `json:"consumed"`
	Index       int64   `json:"index"`
	IIndex      int64   `json:"iindex"`
	LastPostion int     `json:"last_postion"`
	OutOffset   int64   `json:"out_offset"`
}

type chunkReader struct {
	f        *os.File
	r        *bufio.Reader
	buf      []int64
	pos      int
	consumed int64
}

func newChunkReader(p string, skip int64) (*chunkReader, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	if skip > 0 {
		if _, err := f.Seek(skip*8, io.SeekStart); err != nil {
			f.Close()
			return nil, err
		}
	}
	return &chunkReader{f: f, r: bufio.NewReaderSize(f, 1<<20), consumed: skip}, nil
}

func (cr *chunkReader) next() (int64, bool, error) {
	if cr.pos >= len(cr.buf) {
		raw := make([]byte, 8*mergeReadBlock)
		n, err := io.ReadFull(cr.r, raw)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return 0, false, err
		}
		cnt := n / 8
		if cnt == 0 {
			return 0, false, nil
		}
		cr.buf = make([]int64, cnt)
		for i := 0; i < cnt; i++ {
			cr.buf[i] = int64(binary.LittleEndian.Uint64(raw[i*8:]))
		}
		cr.pos = 0
	}
	k := cr.buf[cr.pos]
	cr.pos++
	cr.consumed++
	return k, true, nil
}

func (cr *chunkReader) close() { cr.f.Close() }

type mergeItem struct {
	key   int64
	chunk int
}

type mergeHeap []mergeItem

func (h mergeHeap) Len() int { return len(h) }
func (h mergeHeap) Less(i, j int) bool {
	if h[i].key == h[j].key {
		return h[i].chunk < h[j].chunk
	}
	if ta.GenCompare(h[i].key, h[j].key) {
		return true
	}
	if ta.GenCompare(h[j].key, h[i].key) {
		return false
	}
	return h[i].chunk < h[j].chunk
}
func (h mergeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *mergeHeap) Push(x interface{}) {
	*h = append(*h, x.(mergeItem))
}
func (h *mergeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func phaseMerge7() error {
	final := path("texas_data.txt")
	if exists(final) {
		logf("texas_data.txt exists, skip merge7")
		return nil
	}
	tmp := final + ".tmp"
	statePath := final + ".state.json"

	nChunks := (int(ta.Total) + chunkSize7 - 1) / chunkSize7
	for i := 0; i < nChunks; i++ {
		if !exists(chunk7Path(i)) {
			return fmt.Errorf("missing chunk %d", i)
		}
	}

	var st mergeState
	if !readJSON(statePath, &st) {
		st = mergeState{
			Consumed:    make([]int64, nChunks),
			LastPostion: -1,
		}
	}
	if len(st.Consumed) != nChunks {
		return fmt.Errorf("bad checkpoint: %d consumed entries", len(st.Consumed))
	}
	logf("merge7 resume: index=%d offset=%d", st.Index, st.OutOffset)

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Truncate(st.OutOffset); err != nil {
		return err
	}
	if _, err := f.Seek(st.OutOffset, io.SeekStart); err != nil {
		return err
	}
	w := bufio.NewWriterSize(f, 4<<20)

	readers := make([]*chunkReader, nChunks)
	defer func() {
		for _, r := range readers {
			if r != nil {
				r.close()
			}
		}
	}()
	h := &mergeHeap{}
	for i := 0; i < nChunks; i++ {
		cr, err := newChunkReader(chunk7Path(i), st.Consumed[i])
		if err != nil {
			return err
		}
		readers[i] = cr
		k, ok, err := cr.next()
		if err != nil {
			return err
		}
		if ok {
			heap.Push(h, mergeItem{key: k, chunk: i})
		}
	}
	heap.Init(h)

	index := st.Index
	iindex := st.IIndex
	lastPostion := st.LastPostion
	start := time.Now()
	lastCkpt := index

	flushState := func() error {
		if err := w.Flush(); err != nil {
			return err
		}
		off, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		for i, cr := range readers {
			st.Consumed[i] = cr.consumed
		}
		st.Index = index
		st.IIndex = iindex
		st.LastPostion = lastPostion
		st.OutOffset = off
		return atomicJSON(statePath, &st)
	}

	for h.Len() > 0 {
		it := heap.Pop(h).(mergeItem)
		kd := ta.LookupKeyData(it.key)
		if kd == nil {
			return fmt.Errorf("no key data for %d", it.key)
		}
		if kd.Postion < lastPostion {
			return fmt.Errorf("position decreased: key=%d postion=%d last=%d", it.key, kd.Postion, lastPostion)
		}
		if kd.Postion != lastPostion {
			iindex = index
			lastPostion = kd.Postion
		}
		line := strconv.FormatInt(it.key, 10) + " " + strconv.Itoa(kd.Postion) + " " +
			strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(ta.Total, 10) + " " +
			ta.ToString(it.key) + " " + strconv.FormatInt(kd.Max, 10) + " " +
			ta.ToString(kd.Max) + " " + strconv.Itoa(kd.Type) + "\n"
		if _, err := w.WriteString(line); err != nil {
			return err
		}
		index++

		k, ok, err := readers[it.chunk].next()
		if err != nil {
			return err
		}
		if ok {
			heap.Push(h, mergeItem{key: k, chunk: it.chunk})
		}

		if index-lastCkpt >= mergeCkptRows {
			if err := flushState(); err != nil {
				return err
			}
			lastCkpt = index
			logf("merge7 %d/%d (%.2f%%), elapsed %s, eta %s",
				index, ta.Total, float64(index)*100/float64(ta.Total),
				time.Since(start).Round(time.Second), eta(start, index, ta.Total))
		}
	}

	if err := flushState(); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		return err
	}
	os.Remove(statePath)
	logf("merge7 done: %d rows in %s", index, time.Since(start).Round(time.Second))
	if index != ta.Total {
		return fmt.Errorf("wrote %d rows, want %d", index, ta.Total)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Phase 3: compact color/normal tables from texas_data.txt (single pass).
// ---------------------------------------------------------------------------

type opt7State struct {
	Rows         int64   `json:"rows"`
	InputOffset  int64   `json:"input_offset"`
	ColorOffset  int64   `json:"color_offset"`
	NormalOffset int64   `json:"normal_offset"`
	ColorKeys    []int64 `json:"color_keys"`
	NormalKeys   []int64 `json:"normal_keys"`
}

func phaseOpt7() error {
	colorFinal := path("texas_data_color.txt")
	normalFinal := path("texas_data_normal.txt")
	if exists(colorFinal) && exists(normalFinal) {
		logf("opt7 files exist, skip")
		return nil
	}
	statePath := path("opt7.state.json")
	var st opt7State
	resume := readJSON(statePath, &st)
	colorTmp := colorFinal + ".tmp"
	normalTmp := normalFinal + ".tmp"

	in, err := os.Open(path("texas_data.txt"))
	if err != nil {
		return err
	}
	defer in.Close()
	if resume {
		if _, err := in.Seek(st.InputOffset, io.SeekStart); err != nil {
			return err
		}
	}

	colorKeys := map[int64]bool{}
	normalKeys := map[int64]bool{}
	for _, k := range st.ColorKeys {
		colorKeys[k] = true
	}
	for _, k := range st.NormalKeys {
		normalKeys[k] = true
	}
	cf, err := os.OpenFile(colorTmp, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer cf.Close()
	if err := cf.Truncate(st.ColorOffset); err != nil {
		return err
	}
	if _, err := cf.Seek(st.ColorOffset, io.SeekStart); err != nil {
		return err
	}
	cw := bufio.NewWriterSize(cf, 4<<20)

	nf, err := os.OpenFile(normalTmp, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer nf.Close()
	if err := nf.Truncate(st.NormalOffset); err != nil {
		return err
	}
	if _, err := nf.Seek(st.NormalOffset, io.SeekStart); err != nil {
		return err
	}
	nw := bufio.NewWriterSize(nf, 4<<20)

	lr := &lineReader{r: bufio.NewReaderSize(in, 4<<20), off: st.InputOffset}
	rows := st.Rows
	start := time.Now()
	lastCkpt := rows

	flushState := func() error {
		if err := cw.Flush(); err != nil {
			return err
		}
		if err := nw.Flush(); err != nil {
			return err
		}
		co, err := cf.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		no, err := nf.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		st.Rows = rows
		st.InputOffset = lr.off
		st.ColorOffset = co
		st.NormalOffset = no
		st.ColorKeys = st.ColorKeys[:0]
		st.NormalKeys = st.NormalKeys[:0]
		for k := range colorKeys {
			st.ColorKeys = append(st.ColorKeys, k)
		}
		for k := range normalKeys {
			st.NormalKeys = append(st.NormalKeys, k)
		}
		return atomicJSON(statePath, &st)
	}

	for {
		line, ok, err := lr.next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		f := strings.Split(line, " ")
		if len(f) != 8 {
			return fmt.Errorf("bad row %q", line)
		}
		key, err := strconv.ParseInt(f[0], 10, 64)
		if err != nil {
			return err
		}
		i, err := strconv.Atoi(f[1])
		if err != nil {
			return err
		}
		iindex, err := strconv.ParseInt(f[2], 10, 64)
		if err != nil {
			return err
		}
		total, err := strconv.ParseInt(f[3], 10, 64)
		if err != nil {
			return err
		}
		max, err := strconv.ParseInt(f[5], 10, 64)
		if err != nil {
			return err
		}
		typ, err := strconv.Atoi(f[7])
		if err != nil {
			return err
		}

		if isFlushType(typ) {
			ck := ta.ChangeColor(key)
			if !colorKeys[ck] {
				colorKeys[ck] = true
				cm := ta.Max(ck)
				line := strconv.FormatInt(ck, 10) + " " + strconv.Itoa(i) + " " +
					strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(total, 10) + " " +
					ta.ToString(ck) + " " + strconv.FormatInt(cm, 10) + " " +
					ta.ToString(cm) + " " + strconv.Itoa(typ) + "\n"
				if _, err := cw.WriteString(line); err != nil {
					return err
				}
			}
		} else {
			nk := ta.RemoveColor(key)
			if !normalKeys[nk] {
				normalKeys[nk] = true
				line := strconv.FormatInt(nk, 10) + " " + strconv.Itoa(i) + " " +
					strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(total, 10) + " " +
					f[4] + " " + strconv.FormatInt(max, 10) + " " + f[6] + " " +
					strconv.Itoa(typ) + "\n"
				if _, err := nw.WriteString(line); err != nil {
					return err
				}
			}
		}

		rows++
		if rows-lastCkpt >= opt7CkptRows {
			if err := flushState(); err != nil {
				return err
			}
			lastCkpt = rows
			logf("opt7 %d rows (%.2f%%), elapsed %s",
				rows, float64(rows)*100/float64(ta.Total), time.Since(start).Round(time.Second))
		}
	}

	if err := flushState(); err != nil {
		return err
	}
	if err := cw.Flush(); err != nil {
		return err
	}
	if err := nw.Flush(); err != nil {
		return err
	}
	if err := cf.Sync(); err != nil {
		return err
	}
	if err := nf.Sync(); err != nil {
		return err
	}
	cf.Close()
	nf.Close()
	if err := os.Rename(colorTmp, colorFinal); err != nil {
		return err
	}
	if err := os.Rename(normalTmp, normalFinal); err != nil {
		return err
	}
	os.Remove(statePath)

	nColor := int64(len(colorKeys))
	nNormal := int64(len(normalKeys))
	logf("opt7 done: rows=%d color=%d normal=%d in %s", rows, nColor, nNormal, time.Since(start).Round(time.Second))

	// The 52-card tables cannot be derived from the 54-card ones by simply
	// subtracting the joker rows: suit normalization merges different hands.
	// Check invariants instead of hard-coding a row count.
	if nColor == 0 || nNormal == 0 {
		return fmt.Errorf("opt7 produced an empty table: color=%d normal=%d", nColor, nNormal)
	}
	return nil
}

type lineReader struct {
	r   *bufio.Reader
	off int64
}

func (lr *lineReader) next() (string, bool, error) {
	line, err := lr.r.ReadString('\n')
	if len(line) == 0 && err != nil {
		if err == io.EOF {
			return "", false, nil
		}
		return "", false, err
	}
	lr.off += int64(len(line))
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")
	return line, true, nil
}

func isFlushType(t int) bool {
	return t == ta.TexasCardTypeTongHua || t == ta.TexasCardTypeTongHuaShun || t == ta.TexasCardTypeKingTongHuaShun
}

// ---------------------------------------------------------------------------
// Phase 4: 6-card and 5-card tables.
// ---------------------------------------------------------------------------

func extraFile(n int) string {
	return path(fmt.Sprintf("texas_data_extra_%d.txt", n))
}

func phaseExtra(n int) error {
	final := extraFile(n)
	if exists(final) {
		logf("extra %d exists, skip", n)
		return nil
	}
	logf("extra %d: enumerating", n)
	keys := enumKeys(n)
	if int64(len(keys)) != comb(52, n) {
		return fmt.Errorf("extra %d: enumerated %d keys", n, len(keys))
	}
	logf("extra %d: sorting %d keys", n, len(keys))
	t0 := time.Now()
	sort.Slice(keys, func(i, j int) bool { return ta.GenCompare(keys[i], keys[j]) })
	logf("extra %d: sorted in %s", n, time.Since(t0).Round(time.Second))

	tmp := final + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriterSize(f, 4<<20)
	i := 0
	var iindex int64
	var index int64
	var lastMax int64
	size := int64(len(keys))
	keepMax := int64(0)
	for _, k := range keys {
		curMax := ta.Max(k)
		if index == 0 {
			iindex = index
			lastMax = curMax
		} else if ta.Equal(lastMax, curMax) {
			lastMax = curMax
		} else {
			i++
			iindex = index
			lastMax = curMax
		}
		line := strconv.FormatInt(k, 10) + " " + strconv.Itoa(i) + " " +
			strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(size, 10) + " " +
			ta.ToString(k) + " " + strconv.FormatInt(curMax, 10) + " " +
			ta.ToString(curMax) + " " + strconv.Itoa(ta.MaxType(k)) + "\n"
		if _, err := w.WriteString(line); err != nil {
			f.Close()
			return err
		}
		index++
		keepMax = curMax
		_ = keepMax
		if index%5_000_000 == 0 {
			logf("extra %d: %d/%d, elapsed %s", n, index, size, time.Since(t0).Round(time.Second))
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		return err
	}
	logf("extra %d done: %d rows in %s", n, index, time.Since(t0).Round(time.Second))
	if index != size {
		return fmt.Errorf("extra %d wrote %d rows", n, index)
	}
	return nil
}

func phaseExtraOpt(n int) error {
	colorFinal := path(fmt.Sprintf("texas_data_extra_color_%d.txt", n))
	normalFinal := path(fmt.Sprintf("texas_data_extra_normal_%d.txt", n))
	if exists(colorFinal) && exists(normalFinal) {
		logf("extraopt %d exists, skip", n)
		return nil
	}
	in, err := os.Open(extraFile(n))
	if err != nil {
		return err
	}
	defer in.Close()

	colorTmp := colorFinal + ".tmp"
	normalTmp := normalFinal + ".tmp"
	cf, err := os.Create(colorTmp)
	if err != nil {
		return err
	}
	defer cf.Close()
	cw := bufio.NewWriterSize(cf, 4<<20)
	nf, err := os.Create(normalTmp)
	if err != nil {
		return err
	}
	defer nf.Close()
	nw := bufio.NewWriterSize(nf, 4<<20)

	colorKeys := map[int64]bool{}
	normalKeys := map[int64]bool{}
	lr := &lineReader{r: bufio.NewReaderSize(in, 4<<20)}
	for {
		line, ok, err := lr.next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		f := strings.Split(line, " ")
		if len(f) != 8 {
			return fmt.Errorf("bad extra row %q", line)
		}
		key, err := strconv.ParseInt(f[0], 10, 64)
		if err != nil {
			return err
		}
		i, err := strconv.Atoi(f[1])
		if err != nil {
			return err
		}
		iindex, err := strconv.ParseInt(f[2], 10, 64)
		if err != nil {
			return err
		}
		total, err := strconv.ParseInt(f[3], 10, 64)
		if err != nil {
			return err
		}
		max, err := strconv.ParseInt(f[5], 10, 64)
		if err != nil {
			return err
		}
		typ, err := strconv.Atoi(f[7])
		if err != nil {
			return err
		}

		if isFlushType(typ) {
			ck := ta.ChangeColor(key)
			if !colorKeys[ck] {
				colorKeys[ck] = true
				cm := ta.Max(ck)
				line := strconv.FormatInt(ck, 10) + " " + strconv.Itoa(i) + " " +
					strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(total, 10) + " " +
					ta.ToString(ck) + " " + strconv.FormatInt(cm, 10) + " " +
					ta.ToString(cm) + " " + strconv.Itoa(typ) + "\n"
				if _, err := cw.WriteString(line); err != nil {
					return err
				}
			}
		} else {
			nk := ta.RemoveColor(key)
			if !normalKeys[nk] {
				normalKeys[nk] = true
				line := strconv.FormatInt(nk, 10) + " " + strconv.Itoa(i) + " " +
					strconv.FormatInt(iindex, 10) + " " + strconv.FormatInt(total, 10) + " " +
					f[4] + " " + strconv.FormatInt(max, 10) + " " + f[6] + " " +
					strconv.Itoa(typ) + "\n"
				if _, err := nw.WriteString(line); err != nil {
					return err
				}
			}
		}
	}
	if err := cw.Flush(); err != nil {
		return err
	}
	if err := nw.Flush(); err != nil {
		return err
	}
	if err := cf.Sync(); err != nil {
		return err
	}
	if err := nf.Sync(); err != nil {
		return err
	}
	cf.Close()
	nf.Close()
	if err := os.Rename(colorTmp, colorFinal); err != nil {
		return err
	}
	if err := os.Rename(normalTmp, normalFinal); err != nil {
		return err
	}
	logf("extraopt %d done: color=%d normal=%d", n, len(colorKeys), len(normalKeys))
	return nil
}

// ---------------------------------------------------------------------------
// Phase 4b: keep the stored "max" representation identical to the old tables.
//
// The brute-force Max() may pick a different (but strength-equivalent) suit
// assignment than the 54-card generator did, and GetMaxHandPub maps Max back
// onto the dealt cards, so the output cards would change. Every 52-card key
// also exists in the old 54-card compact tables (new keys are a strict subset),
// so copy the old Max/MaxStr for those rows.
// ---------------------------------------------------------------------------

func phaseFixExtra() error {
	names := []string{
		"texas_data_extra_color_6.txt", "texas_data_extra_normal_6.txt",
		"texas_data_extra_color_5.txt", "texas_data_extra_normal_5.txt",
	}
	for _, name := range names {
		if err := fixExtraFile(name); err != nil {
			return err
		}
	}
	return nil
}

func fixExtraFile(name string) error {
	refPath := filepath.Join(*refDir, name)
	rf, err := os.Open(refPath)
	if err != nil {
		return err
	}
	defer rf.Close()

	type maxPair struct{ max, maxStr string }
	old := map[int64]maxPair{}
	sc := bufio.NewScanner(rf)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		f := strings.Split(sc.Text(), " ")
		if len(f) != 8 {
			continue
		}
		key, err := strconv.ParseInt(f[0], 10, 64)
		if err != nil {
			return err
		}
		old[key] = maxPair{max: f[5], maxStr: f[6]}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	cur := path(name)
	nf, err := os.Open(cur)
	if err != nil {
		return err
	}
	defer nf.Close()
	tmp := cur + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer out.Close()
	w := bufio.NewWriterSize(out, 4<<20)

	rows, patched, missing := 0, 0, 0
	lr := &lineReader{r: bufio.NewReaderSize(nf, 4<<20)}
	for {
		line, ok, err := lr.next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		f := strings.Split(line, " ")
		if len(f) != 8 {
			return fmt.Errorf("fixextra %s: bad row %q", name, line)
		}
		key, err := strconv.ParseInt(f[0], 10, 64)
		if err != nil {
			return err
		}
		rows++
		if p, ok := old[key]; ok {
			f[5], f[6] = p.max, p.maxStr
			patched++
		} else {
			missing++
		}
		if _, err := w.WriteString(strings.Join(f, " ") + "\n"); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	out.Close()
	nf.Close()
	if err := os.Rename(tmp, cur); err != nil {
		return err
	}
	logf("fixextra %s: rows=%d patched=%d no-old-row=%d", name, rows, patched, missing)
	if missing > 0 {
		return fmt.Errorf("fixextra %s: %d rows have no old max", name, missing)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Phase 5: N-card win-rate tables.
// ---------------------------------------------------------------------------

func transFile(n int) string {
	return path(fmt.Sprintf("texas_data_%d.txt", n))
}

func transPartPath(n, part int) string {
	return path(fmt.Sprintf("trans_%d_part_%03d.txt", n, part))
}

func phaseTrans(n int) error {
	final := transFile(n)
	if exists(final) {
		logf("trans %d exists, skip", n)
		return nil
	}
	logf("trans %d: enumerating", n)
	keys := enumKeys(n)
	want := comb(52, n)
	if int64(len(keys)) != want {
		return fmt.Errorf("trans %d: enumerated %d keys", n, len(keys))
	}
	nParts := runtime.NumCPU() * 2
	if nParts > len(keys) {
		nParts = len(keys)
	}
	partSize := (len(keys) + nParts - 1) / nParts
	logf("trans %d: %d keys, %d parts", n, len(keys), nParts)

	deck := ta.GenAllCards()
	startAll := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU())
	var failure atomic.Value
	var done atomic.Int64

	for part := 0; part < nParts; part++ {
		s := part * partSize
		e := s + partSize
		if e > len(keys) {
			e = len(keys)
		}
		if s >= e {
			continue
		}
		p := transPartPath(n, part)
		if exists(p) {
			done.Add(1)
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(part, s, e int, p string) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					failure.Store(fmt.Errorf("trans %d part %d panic: %v", n, part, r))
				}
			}()
			t0 := time.Now()
			if err := processTransPart(n, keys[s:e], deck, p); err != nil {
				failure.Store(err)
				return
			}
			d := done.Add(1)
			logf("trans %d part %d/%d done (%d keys), %s, elapsed %s, eta %s",
				n, part+1, nParts, e-s, time.Since(t0).Round(time.Second),
				time.Since(startAll).Round(time.Second),
				eta(startAll, d, int64(nParts)))
		}(part, s, e, p)
	}
	wg.Wait()
	if v := failure.Load(); v != nil {
		return v.(error)
	}

	// Concatenate parts into the final file.
	tmp := final + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	ow := bufio.NewWriterSize(out, 4<<20)
	rows := int64(0)
	for part := 0; part < nParts; part++ {
		p := transPartPath(n, part)
		if !exists(p) {
			continue
		}
		in, err := os.Open(p)
		if err != nil {
			out.Close()
			return err
		}
		sc := bufio.NewScanner(in)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			if _, err := ow.WriteString(sc.Text() + "\n"); err != nil {
				in.Close()
				out.Close()
				return err
			}
			rows++
		}
		if err := sc.Err(); err != nil {
			in.Close()
			out.Close()
			return err
		}
		in.Close()
	}
	if err := ow.Flush(); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		return err
	}
	// Parts can be removed after a successful concatenation.
	for part := 0; part < nParts; part++ {
		os.Remove(transPartPath(n, part))
	}
	logf("trans %d done: %d rows in %s", n, rows, time.Since(startAll).Round(time.Second))
	if rows != int64(len(keys)) {
		return fmt.Errorf("trans %d wrote %d rows, want %d", n, rows, len(keys))
	}
	return nil
}

func processTransPart(n int, keys []int64, deck []int, outPath string) error {
	tmp := outPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 4<<20)

	keyCards := make([]int, n)
	cards := make([]int, 7)
	remaining := make([]int, 0, 52-n)
	addition := make([]int, 7-n)

	writeOne := func(key int64, nk []int) error {
		remaining = remaining[:0]
		for _, c := range deck {
			inKey := false
			for _, kc := range nk {
				if c == kc {
					inKey = true
					break
				}
			}
			if !inKey {
				remaining = append(remaining, c)
			}
		}

		var win int64
		var num int64
		min := 1.0
		max := 0.0
		var enumErr error
		ta.Permutation(remaining, 0, 0, 7-n, addition, func(add []int) {
			copy(cards, nk)
			copy(cards[n:], add)
			sort.Ints(cards)
			k7 := ta.GenCardBindInts(cards)
			kd := ta.LookupKeyData(k7)
			if kd == nil {
				enumErr = fmt.Errorf("trans %d: no key data for %d", n, k7)
				return
			}
			win += int64(kd.Index)
			num++
			p := float64(kd.Index) / float64(ta.Total)
			if p < min {
				min = p
			}
			if p > max {
				max = p
			}
		})
		if enumErr != nil {
			return enumErr
		}
		if num == 0 {
			return fmt.Errorf("trans %d: no supersets for %d", n, key)
		}
		avg := float64(win) / float64(num) / float64(ta.Total)
		line := strconv.FormatInt(key, 10) + " " +
			strconv.FormatFloat(avg, 'g', -1, 64) + " " +
			strconv.FormatFloat(min, 'g', -1, 64) + " " +
			strconv.FormatFloat(max, 'g', -1, 64) + " " +
			ta.ToString(key) + "\n"
		_, err := w.WriteString(line)
		return err
	}

	for _, key := range keys {
		// Decode the key into its card bytes. n <= 6 and every byte is two
		// decimal digits, so ToArray gives the cards from high to low.
		cs := ta.ToArray(key)
		if len(cs) != n {
			return fmt.Errorf("trans %d: key %d decoded to %d cards", n, key, len(cs))
		}
		for i, c := range cs {
			keyCards[i] = int(c.ToByte())
		}
		sort.Ints(keyCards)
		if err := writeOne(key, keyCards); err != nil {
			return err
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, outPath)
}

func phaseTransOpt(n int) error {
	final := path(fmt.Sprintf("texas_data_opt_%d.txt", n))
	if exists(final) {
		logf("transopt %d exists, skip", n)
		return nil
	}
	logf("transopt %d: pass 1", n)
	type group struct {
		counts map[string]int
		best   string
	}
	groups := map[int64]*group{}

	pass := func(fn func(key int64, value string, text string) error) error {
		in, err := os.Open(transFile(n))
		if err != nil {
			return err
		}
		defer in.Close()
		sc := bufio.NewScanner(in)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			line := sc.Text()
			f := strings.Split(line, " ")
			if len(f) != 5 {
				return fmt.Errorf("bad trans row %q", line)
			}
			key, err := strconv.ParseInt(f[0], 10, 64)
			if err != nil {
				return err
			}
			value := f[1] + " " + f[2] + " " + f[3]
			if err := fn(key, value, line); err != nil {
				return err
			}
		}
		return sc.Err()
	}

	rows := int64(0)
	start := time.Now()
	if err := pass(func(key int64, value string, text string) error {
		rk := ta.RemoveColor(key)
		g := groups[rk]
		if g == nil {
			g = &group{counts: map[string]int{}}
			groups[rk] = g
		}
		g.counts[value]++
		rows++
		if rows%10_000_000 == 0 {
			logf("transopt %d pass1: %d rows, %s", n, rows, time.Since(start).Round(time.Second))
		}
		return nil
	}); err != nil {
		return err
	}
	// Deterministic winner: highest count, then lexicographically larger value.
	for _, g := range groups {
		bestN := -1
		for v, c := range g.counts {
			if c > bestN || (c == bestN && v > g.best) {
				bestN = c
				g.best = v
			}
		}
	}
	logf("transopt %d: %d rows, %d groups", n, rows, len(groups))

	tmp := final + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriterSize(out, 4<<20)
	rows2 := int64(0)
	if err := pass(func(key int64, value string, text string) error {
		rk := ta.RemoveColor(key)
		g := groups[rk]
		if g.best != value {
			if _, err := w.WriteString(strconv.FormatInt(key, 10) + " 0 " + value + " " + ta.ToString(key) + "\n"); err != nil {
				return err
			}
		}
		rows2++
		if rows2%10_000_000 == 0 {
			logf("transopt %d pass2: %d rows, %s", n, rows2, time.Since(start).Round(time.Second))
		}
		return nil
	}); err != nil {
		out.Close()
		return err
	}
	for rk, g := range groups {
		if _, err := w.WriteString(strconv.FormatInt(rk, 10) + " 1 " + g.best + " " + ta.ToString(rk) + "\n"); err != nil {
			out.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, final); err != nil {
		return err
	}
	logf("transopt %d done in %s", n, time.Since(start).Round(time.Second))
	return nil
}

// ---------------------------------------------------------------------------
// Status
// ---------------------------------------------------------------------------

func printStatus() {
	files := []string{
		"texas_data_color.txt",
		"texas_data_normal.txt",
		"texas_data_extra_color_6.txt",
		"texas_data_extra_normal_6.txt",
		"texas_data_extra_color_5.txt",
		"texas_data_extra_normal_5.txt",
		"texas_data_opt_2.txt",
		"texas_data_opt_3.txt",
		"texas_data_opt_4.txt",
		"texas_data_opt_5.txt",
		"texas_data_opt_6.txt",
	}
	for _, name := range files {
		p := path(name)
		st, err := os.Stat(p)
		if err != nil {
			logf("%-34s missing", name)
			continue
		}
		logf("%-34s %10d bytes", name, st.Size())
	}
	for _, name := range []string{"texas_data.txt", "texas_data_extra_6.txt", "texas_data_extra_5.txt", "texas_data_2.txt", "texas_data_3.txt", "texas_data_4.txt", "texas_data_5.txt", "texas_data_6.txt"} {
		p := path(name)
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		logf("%-34s %10d bytes", name, st.Size())
	}
}
