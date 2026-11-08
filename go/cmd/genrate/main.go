package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	ta "github.com/esrrhs/texas_algorithm/go"
)

// genrate computes exact "hand vs random opponent" equity tables for 2 hole
// cards plus 0, 1, 2 or 3 known community cards.
//
// The trick that keeps this cheap: enumerate every 5-card board once, rank the
// C(47,2) hole-card pairs of the remaining 47 cards against that board, and
// then every pair's equity against a random opponent is a prefix count in that
// sorted list. One board therefore serves all 1081 hole hands at once, and
// every board containing a given known-card subset contributes to the
// corresponding entry, so a single pass fills N=0, N=1 and N=2 tables.
//
// Output files (one row per canonical key):
//
//	<key> <win> <tie> <total>
//
// The canonical key labels suits by ascending per-suit signature (which ranks
// the suit holds, and which of them are hole cards). That is the invariant of
// the 24-element suit permutation group, which leaves equity unchanged, and it
// keeps the hole/public roles distinct so different roles never share a key.

// ckptVer identifies the canonical-key scheme. Bump it whenever canonKey or the
// accumulation logic changes: a checkpoint written by an older version holds
// incompatible keys and must never be resumed.
const ckptVer = 4

// cnt aggregates one canonical key: wins, ties and the number of evaluated
// (board, opponent) pairs. N is kept so that every row can be normalised back
// to a single physical situation and asserted against its orbit size.
type cnt struct{ W, T, N int64 }

// cnt32 is the compact form used by levels 4 and 5. A per-key total there is at
// most 45,540*24 and 990*24 respectively, so 32 bits are enough and the 123M
// rows of level 5 stay within memory.
type cnt32 struct{ W, T, N uint32 }

const shardBits = 8

func hashKey(k int64) uint64 {
	x := uint64(k)
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	return x
}

type shard64 struct {
	mu sync.Mutex
	m  map[int64]cnt
}

type shard32 struct {
	mu sync.Mutex
	m  map[int64]cnt32
}

// acc is a sharded set of accumulators, one slice per level. Sharding lets all
// workers write into the same maps, so a level is materialised exactly once
// instead of once per worker.
type acc struct {
	l0, l1, l2, l3 []shard64
	l4, l5         []shard32
	maxN           int
}

func newShards64(shards int, hint int) []shard64 {
	out := make([]shard64, shards)
	for i := range out {
		out[i].m = make(map[int64]cnt, hint)
	}
	return out
}

func newShards32(shards int, hint int) []shard32 {
	out := make([]shard32, shards)
	for i := range out {
		out[i].m = make(map[int64]cnt32, hint)
	}
	return out
}

// newAcc sizes each level from its known row count so the maps never rehash
// while the 123M-row level is being filled.
func newAcc(maxN int) *acc {
	a := &acc{maxN: maxN}
	if maxN >= 0 {
		a.l0 = newShards64(1<<shardBits, 1)
	}
	if maxN >= 1 {
		a.l1 = newShards64(1<<shardBits, 20)
	}
	if maxN >= 2 {
		a.l2 = newShards64(1<<shardBits, 370)
	}
	if maxN >= 3 {
		a.l3 = newShards64(1<<shardBits, 5030)
	}
	if maxN >= 4 {
		a.l4 = newShards32(1<<shardBits, 54530)
	}
	if maxN >= 5 {
		a.l5 = newShards32(1<<shardBits, 481000)
	}
	return a
}

func add64(sh []shard64, k int64, w, t, l int64) {
	s := &sh[hashKey(k)>>(64-shardBits)]
	s.mu.Lock()
	if v, ok := s.m[k]; ok {
		v.W += w
		v.T += t
		v.N += w + t + l
		s.m[k] = v
	} else {
		s.m[k] = cnt{W: w, T: t, N: w + t + l}
	}
	s.mu.Unlock()
}

func add32(sh []shard32, k int64, w, t, l int64) {
	s := &sh[hashKey(k)>>(64-shardBits)]
	s.mu.Lock()
	if v, ok := s.m[k]; ok {
		v.W += uint32(w)
		v.T += uint32(t)
		v.N += uint32(w + t + l)
		s.m[k] = v
	} else {
		s.m[k] = cnt32{W: uint32(w), T: uint32(t), N: uint32(w + t + l)}
	}
	s.mu.Unlock()
}

func (a *acc) level64(n int) []shard64 {
	switch n {
	case 0:
		return a.l0
	case 1:
		return a.l1
	case 2:
		return a.l2
	default:
		return a.l3
	}
}

func (a *acc) shards64(n int) []shard64 { return a.level64(n) }

func (a *acc) shards32(n int) []shard32 {
	if n == 4 {
		return a.l4
	}
	return a.l5
}

func (a *acc) setShards64(n int, sh []shard64) {
	switch n {
	case 0:
		a.l0 = sh
	case 1:
		a.l1 = sh
	case 2:
		a.l2 = sh
	default:
		a.l3 = sh
	}
}

func (a *acc) setShards32(n int, sh []shard32) {
	if n == 4 {
		a.l4 = sh
	} else {
		a.l5 = sh
	}
}

func (a *acc) len64(n int) int {
	var total int
	for i := range a.shards64(n) {
		total += len(a.shards64(n)[i].m)
	}
	return total
}

func (a *acc) len32(n int) int {
	var total int
	for i := range a.shards32(n) {
		total += len(a.shards32(n)[i].m)
	}
	return total
}

// freeLevel drops one level once it has been written, so the 123M-row level 5
// map does not stay resident while the earlier levels are being compressed.
func (a *acc) freeLevel(n int) {
	if n <= 3 {
		a.setShards64(n, nil)
	} else {
		a.setShards32(n, nil)
	}
}

// reset returns the accumulator to the state of a fresh run, used when a
// checkpoint turns out to be unusable half-way through.
func (a *acc) reset() { *a = *newAcc(a.maxN) }

var pool [52]byte
var pow10 [7]int64

func initTables() {
	n := 0
	for c := 0; c < 4; c++ {
		for v := 2; v <= 14; v++ {
			pool[n] = byte(c<<4 | v)
			n++
		}
	}
	pow10[0] = 1
	for i := 1; i < 7; i++ {
		pow10[i] = pow10[i-1] * 100
	}
}

// keySorted builds the binding key of already sorted cards.
func keySorted(bs []byte) int64 {
	var k int64
	for i, b := range bs {
		k += int64(b) * pow10[len(bs)-1-i]
	}
	return k
}

// key7 inserts x,y into the sorted 5-card board and returns the 7-card key.
func key7(b [5]byte, x, y byte) int64 {
	var buf [7]byte
	i, j := 0, 0
	for j < 5 {
		if i == 0 && x < b[j] {
			buf[i+j] = x
			i++
			continue
		}
		if i == 1 && y < b[j] {
			buf[i+j] = y
			i++
			continue
		}
		buf[i+j] = b[j]
		j++
	}
	if i == 0 {
		buf[5] = x
		buf[6] = y
	} else if i == 1 {
		buf[6] = y
	}
	var k int64
	for n := 0; n < 7; n++ {
		k += int64(buf[n]) * pow10[6-n]
	}
	return k
}

// keyN builds the key of up to 5 arbitrary cards.
func keyN(bs []byte) int64 {
	tmp := make([]byte, len(bs))
	copy(tmp, bs)
	sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
	return keySorted(tmp)
}

// canonKey builds the suit-canonical key of a situation made of the first
// nHole cards plus the remaining public cards. Each suit is labelled 0..3 by
// sorting the suits on its signature (which ranks it holds, and which of those
// are hole cards), which is invariant under the 24-element suit permutation
// group; suits sharing a signature hold the same (rank, role) pairs, so any
// tie-break yields the same labelled cards. Hole cards and public cards occupy
// separate positions of the key because the two roles are not interchangeable.
func canonKey(cs []byte, nHole int) int64 {
	var mask, hole [4]int
	for i, c := range cs {
		mask[c>>4] |= 1 << (c & 0x0f)
		if i < nHole {
			hole[c>>4] |= 1 << (c & 0x0f)
		}
	}
	var label [4]int8
	used := [4]bool{}
	for k := int8(0); k < 4; k++ {
		best := -1
		var bestSig int
		for s := 0; s < 4; s++ {
			if used[s] {
				continue
			}
			sig := mask[s]<<16 | hole[s]
			if best < 0 || sig < bestSig {
				best, bestSig = s, sig
			}
		}
		used[best] = true
		label[best] = k
	}

	var holeCards, pubCards [5]byte
	nh, np := 0, 0
	for i, c := range cs {
		b := byte(label[c>>4])<<4 | (c & 0x0f)
		if i < nHole {
			holeCards[nh] = b
			nh++
		} else {
			pubCards[np] = b
			np++
		}
	}
	sortBytes(holeCards[:nh])
	sortBytes(pubCards[:np])

	var k int64
	for i := 0; i < nh; i++ {
		k = k*100 + int64(holeCards[i])
	}
	for i := 0; i < np; i++ {
		k = k*100 + int64(pubCards[i])
	}
	return k
}

func sortBytes(bs []byte) {
	for i := 1; i < len(bs); i++ {
		for j := i; j > 0 && bs[j] < bs[j-1]; j-- {
			bs[j], bs[j-1] = bs[j-1], bs[j]
		}
	}
}

// key4 sorts four card bytes and returns their binding key.
func key4(a, b, c, d byte) int64 {
	if b < a {
		a, b = b, a
	}
	if d < c {
		c, d = d, c
	}
	if c < a {
		c, a = a, c
	}
	if d < b {
		d, b = b, d
	}
	if b < c {
		b, c = c, b
	}
	if d < c {
		c, d = d, c
	}
	return int64(a)*1000000 + int64(b)*10000 + int64(c)*100 + int64(d)
}

type pairRank struct {
	idx  int32
	code int16 // i*47+j into the 47-card rest list
}

// processBoard computes the equity of every hole pair on this board and
// accumulates it into the three tables.
func processBoard(b [5]byte, a *acc) {
	var rest [47]byte
	n := 0
	for _, c := range pool {
		if c == b[0] || c == b[1] || c == b[2] || c == b[3] || c == b[4] {
			continue
		}
		rest[n] = c
		n++
	}
	if n != 47 {
		panic("rest != 47")
	}

	var arr [1081]pairRank
	k := 0
	for i := 0; i < 47; i++ {
		for j := i + 1; j < 47; j++ {
			idx := ta.LookupKeyData(key7(b, rest[i], rest[j])).Index
			arr[k] = pairRank{idx: int32(idx), code: int16(i*47 + j)}
			k++
		}
	}
	sort.Slice(arr[:], func(x, y int) bool { return arr[x].idx < arr[y].idx })

	// position of each pair inside the sorted array
	var pos [2209]int32
	for p := 0; p < 1081; p++ {
		pos[arr[p].code] = int32(p)
	}
	// per-card lists of sorted positions (sorted by rank because arr is sorted)
	var cardPos [47][46]int32
	var cardCnt [47]int
	for p := 0; p < 1081; p++ {
		j := int(arr[p].code)
		i := j / 47
		j %= 47
		cardPos[i][cardCnt[i]] = int32(p)
		cardCnt[i]++
		cardPos[j][cardCnt[j]] = int32(p)
		cardCnt[j]++
	}

	countBelow := func(card int, idx int32) (less, eq int) {
		c := cardCnt[card]
		ls := sort.Search(c, func(x int) bool { return arr[cardPos[card][x]].idx >= idx })
		hi := sort.Search(c, func(x int) bool { return arr[cardPos[card][x]].idx > idx })
		return ls, hi - ls
	}

	for i := 0; i < 47; i++ {
		for j := i + 1; j < 47; j++ {
			code := int16(i*47 + j)
			p := int(pos[code])
			heroIdx := arr[p].idx
			lo := sort.Search(1081, func(x int) bool { return arr[x].idx >= heroIdx })
			hi := sort.Search(1081, func(x int) bool { return arr[x].idx > heroIdx })

			la, ea := countBelow(i, heroIdx)
			lb, eb := countBelow(j, heroIdx)
			w := int64(lo - la - lb)
			eq := int64(hi - lo - (ea + eb - 1))
			l := 990 - w - eq

			// The same (w, eq, l) is added to every level-n key derived from
			// this (hole pair, board): summing over all completions of a level-n
			// situation reproduces exactly the per-hand evaluation count.
			var cs [7]byte
			cs[0], cs[1] = rest[i], rest[j]
			if a.maxN >= 0 {
				add64(a.l0, canonKey(cs[:2], 2), w, eq, l)
			}
			if a.maxN >= 1 {
				for _, bc := range b {
					cs[2] = bc
					add64(a.l1, canonKey(cs[:3], 2), w, eq, l)
				}
			}
			if a.maxN >= 2 {
				for x := 0; x < 5; x++ {
					for y := x + 1; y < 5; y++ {
						cs[2], cs[3] = b[x], b[y]
						add64(a.l2, canonKey(cs[:4], 2), w, eq, l)
					}
				}
			}
			if a.maxN >= 3 {
				for x := 0; x < 5; x++ {
					for y := x + 1; y < 5; y++ {
						for z := y + 1; z < 5; z++ {
							cs[2], cs[3], cs[4] = b[x], b[y], b[z]
							add64(a.l3, canonKey(cs[:5], 2), w, eq, l)
						}
					}
				}
			}
			if a.maxN >= 4 {
				for x := 0; x < 5; x++ {
					for y := x + 1; y < 5; y++ {
						for z := y + 1; z < 5; z++ {
							for u := z + 1; u < 5; u++ {
								cs[2], cs[3], cs[4], cs[5] = b[x], b[y], b[z], b[u]
								add32(a.l4, canonKey(cs[:6], 2), w, eq, l)
							}
						}
					}
				}
			}
			if a.maxN >= 5 {
				cs[2], cs[3], cs[4], cs[5], cs[6] = b[0], b[1], b[2], b[3], b[4]
				add32(a.l5, canonKey(cs[:7], 2), w, eq, l)
			}
		}
	}
}

type job struct {
	e, d int
}

func main() {
	initTables()
	workers := flag.Int("workers", runtime.NumCPU(), "worker count")
	out := flag.String("out", ".", "output directory")
	ckpt := flag.String("ckpt", "./genrate.ckpt", "checkpoint file")
	resume := flag.Bool("resume", true, "resume from checkpoint")
	maxE := flag.Int("maxE", 51, "stop after this largest-card index (smoke runs)")
	maxN := flag.Int("maxN", 3, "highest known-community-card level to accumulate (0..5)")
	mode := flag.String("mode", "gen", "gen|smoke|verify|count|expect")
	bruteN := flag.Int("bruteN", 4, "expect: highest level verified by brute-force card enumeration (the rest use the closed form)")
	flag.Parse()

	if *mode == "count" {
		countKeys(*out, *maxN)
		return
	}
	if *mode == "expect" {
		expectCounts(*out, *maxN, *bruteN)
		return
	}

	ta.Load()
	fmt.Printf("tables loaded, workers=%d maxN=%d\n", *workers, *maxN)

	a := newAcc(*maxN)
	doneE := map[int]bool{}
	if *resume {
		if loadCkpt(*ckpt, a, doneE) {
			fmt.Printf("resumed: %d boards-largest done, rows=%d/%d/%d/%d\n", len(doneE),
				a.len64(0), a.len64(1), a.len64(2), a.len64(3))
		}
	}

	t0 := time.Now()
	boardsDone := int64(0)
	var boardsDoneMu sync.Mutex
	for e := 4; e <= *maxE; e++ {
		if doneE[e] {
			continue
		}
		jobs := make(chan job, 1024)
		var wg sync.WaitGroup
		for w := 0; w < *workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for jb := range jobs {
					var b [5]byte
					b[4] = pool[jb.e]
					b[3] = pool[jb.d]
					cntLocal := 0
					// Enumerate every 5-subset 0 <= aa < bb < c < d < e < 52.
					// Lower bounds must be 2/1/0: stopping at 3/2/1 silently
					// dropped every board whose two smallest cards are below
					// those limits (10% of the board space).
					for c := jb.d - 1; c >= 2; c-- {
						b[2] = pool[c]
						for bb := c - 1; bb >= 1; bb-- {
							b[1] = pool[bb]
							for aa := bb - 1; aa >= 0; aa-- {
								b[0] = pool[aa]
								processBoard(b, a)
								cntLocal++
							}
						}
					}
					boardsDoneMu.Lock()
					boardsDone += int64(cntLocal)
					boardsDoneMu.Unlock()
				}
			}()
		}
		for d := 3; d < e; d++ {
			jobs <- job{e: e, d: d}
		}
		close(jobs)
		wg.Wait()

		doneE[e] = true
		el := time.Since(t0)
		total := 2598960.0
		prog := float64(boardsDone) / total
		eta := time.Duration(0)
		if prog > 0 {
			eta = time.Duration(float64(el) * (1 - prog) / prog)
		}
		fmt.Printf("[%s] e=%d done, boards=%d (%.2f%%), elapsed=%s, eta=%s, m0=%d m1=%d m2=%d m3=%d m4=%d m5=%d\n",
			time.Now().Format("15:04:05"), e, boardsDone, prog*100,
			el.Round(time.Second), eta.Round(time.Second),
			a.len64(0), a.len64(1), a.len64(2), a.len64(3), a.len32(4), a.len32(5))

		if e%8 == 0 {
			saveCkpt(*ckpt, a, doneE)
		}
	}
	saveCkpt(*ckpt, a, doneE)

	if *mode == "smoke" {
		dump(a, *out)
		return
	}
	writeTables(a, *out, *maxN)
	fmt.Printf("done in %s\n", time.Since(t0).Round(time.Second))
}

const ckptMagic = "TXWK"

func putRec(rec *[20]byte, k, w, t, n int64) {
	binary.LittleEndian.PutUint64(rec[0:8], uint64(k))
	binary.LittleEndian.PutUint32(rec[8:12], uint32(w))
	binary.LittleEndian.PutUint32(rec[12:16], uint32(t))
	binary.LittleEndian.PutUint32(rec[16:20], uint32(n))
}

func readU32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func readU64r(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

// saveCkpt writes the accumulators in a compact binary form. gob cannot be used
// here because the shards contain mutexes; records are key u64 plus w/t/n u32,
// i.e. 20 bytes per accumulated key.
func saveCkpt(path string, a *acc, done map[int]bool) {
	t0 := time.Now()
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		panic(err)
	}
	bw := bufio.NewWriterSize(f, 1<<20)
	bw.WriteString(ckptMagic)
	var b4 [4]byte
	binary.LittleEndian.PutUint32(b4[:], ckptVer)
	bw.Write(b4[:])
	binary.LittleEndian.PutUint32(b4[:], uint32(a.maxN))
	bw.Write(b4[:])
	var dl []int
	for e := range done {
		dl = append(dl, e)
	}
	sort.Ints(dl)
	binary.LittleEndian.PutUint32(b4[:], uint32(len(dl)))
	bw.Write(b4[:])
	for _, e := range dl {
		binary.LittleEndian.PutUint32(b4[:], uint32(e))
		bw.Write(b4[:])
	}
	var rec [20]byte
	for n := 0; n <= a.maxN; n++ {
		var count uint64
		if n <= 3 {
			for i := range a.shards64(n) {
				count += uint64(len(a.shards64(n)[i].m))
			}
		} else {
			for i := range a.shards32(n) {
				count += uint64(len(a.shards32(n)[i].m))
			}
		}
		binary.LittleEndian.PutUint64(rec[0:8], count)
		bw.Write(rec[0:8])
		if n <= 3 {
			sh := a.shards64(n)
			for i := range sh {
				for k, v := range sh[i].m {
					putRec(&rec, k, v.W, v.T, v.N)
					bw.Write(rec[:])
				}
			}
		} else {
			sh := a.shards32(n)
			for i := range sh {
				for k, v := range sh[i].m {
					putRec(&rec, k, int64(v.W), int64(v.T), int64(v.N))
					bw.Write(rec[:])
				}
			}
		}
	}
	bw.Flush()
	f.Close()
	if err := os.Rename(tmp, path); err != nil {
		panic(err)
	}
	if fi, err := os.Stat(path); err == nil {
		fmt.Printf("ckpt %s bytes=%d elapsed=%s\n", path, fi.Size(), time.Since(t0).Round(time.Second))
	}
}

// loadCkpt restores a previous run. It returns false when the file is absent,
// unreadable or written by an incompatible version; the caller then starts from
// scratch. A failed read is rolled back so a corrupt checkpoint never leaves
// half-loaded state behind.
func loadCkpt(path string, a *acc, done map[int]bool) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	br := bufio.NewReaderSize(f, 1<<20)
	if err := readCkpt(br, a, done); err != nil {
		fmt.Printf("ignoring checkpoint %s: %v\n", path, err)
		clear(done)
		a.reset()
		return false
	}
	return true
}

func readCkpt(br *bufio.Reader, a *acc, done map[int]bool) error {
	var magic [8]byte
	if _, err := io.ReadFull(br, magic[:]); err != nil {
		return err
	}
	if string(magic[:]) != ckptMagic {
		return fmt.Errorf("bad magic")
	}
	ver, err := readU32(br)
	if err != nil {
		return err
	}
	if int(ver) != ckptVer {
		return fmt.Errorf("version %d != %d", ver, ckptVer)
	}
	storedMaxN, err := readU32(br)
	if err != nil {
		return err
	}
	if int(storedMaxN) != a.maxN {
		return fmt.Errorf("maxN %d != %d", storedMaxN, a.maxN)
	}
	nd, err := readU32(br)
	if err != nil {
		return err
	}
	for i := uint32(0); i < nd; i++ {
		e, err := readU32(br)
		if err != nil {
			return err
		}
		done[int(e)] = true
	}
	var rec [20]byte
	for n := 0; n <= a.maxN; n++ {
		count, err := readU64r(br)
		if err != nil {
			return err
		}
		if n <= 3 {
			sh := newShards64(1<<shardBits, int(count>>shardBits)+1)
			for i := uint64(0); i < count; i++ {
				if _, err := io.ReadFull(br, rec[:]); err != nil {
					return err
				}
				k := int64(binary.LittleEndian.Uint64(rec[0:8]))
				sh[hashKey(k)>>(64-shardBits)].m[k] = cnt{
					W: int64(binary.LittleEndian.Uint32(rec[8:12])),
					T: int64(binary.LittleEndian.Uint32(rec[12:16])),
					N: int64(binary.LittleEndian.Uint32(rec[16:20])),
				}
			}
			a.setShards64(n, sh)
		} else {
			sh := newShards32(1<<shardBits, int(count>>shardBits)+1)
			for i := uint64(0); i < count; i++ {
				if _, err := io.ReadFull(br, rec[:]); err != nil {
					return err
				}
				k := int64(binary.LittleEndian.Uint64(rec[0:8]))
				sh[hashKey(k)>>(64-shardBits)].m[k] = cnt32{
					W: binary.LittleEndian.Uint32(rec[8:12]),
					T: binary.LittleEndian.Uint32(rec[12:16]),
					N: binary.LittleEndian.Uint32(rec[16:20]),
				}
			}
			a.setShards32(n, sh)
		}
	}
	return nil
}

// baseTotals[n] is the number of (board completion, opponent pair) evaluations
// of one physical situation with n known community cards:
//
//	n=0: C(50,5)=2118760 x 990      n=3: C(47,2)=1081 x 990
//	n=1: C(49,4)=211876  x 990      n=4: C(46,1)=46   x 990
//	n=2: C(48,3)=17296   x 990      n=5: C(45,0)=1    x 990
var baseTotals = []int64{2097572400, 209757240, 17123040, 1070190, 45540, 990}

// orbitSize returns the number of distinct card assignments that the
// suit-canonical key stands for, i.e. 24 divided by the size of the stabiliser
// of the situation inside the 24-element suit permutation group. Two suits
// stabilise each other exactly when they carry the same (rank, role) signature,
// so the stabiliser is the product of the factorials of the signature
// multiplicities.
func orbitSize(key int64, nCards int) int {
	var cs [7]byte
	k := key
	for i := nCards - 1; i >= 0; i-- {
		cs[i] = byte(k % 100)
		k /= 100
	}
	var mask, hole [4]int
	for i := 0; i < nCards; i++ {
		s := cs[i] >> 4
		mask[s] |= 1 << (cs[i] & 0x0f)
		if i < 2 {
			hole[s] |= 1 << (cs[i] & 0x0f)
		}
	}
	used := [4]bool{}
	stab := 1
	for s := 0; s < 4; s++ {
		if used[s] {
			continue
		}
		sig := mask[s]<<16 | hole[s]
		m := 0
		for t := s; t < 4; t++ {
			if !used[t] && (mask[t]<<16|hole[t]) == sig {
				used[t] = true
				m++
			}
		}
		for i := 2; i <= m; i++ {
			stab *= i
		}
	}
	return 24 / stab
}

// row is one normalised table entry: counts of a single physical situation.
type row struct {
	key  int64
	w, t uint32
}

// normalise divides every orbit-aggregated entry back to one physical
// situation and asserts that the per-row total is exactly orbit x per-hand
// evaluations. Both checks are independent: the first comes from the board
// sweep, the second is recomputed from the key alone.
func normaliseRow(n int, k, w, t, tot int64) row {
	base := baseTotals[n]
	if tot == 0 || tot%base != 0 {
		panic(fmt.Sprintf("n=%d key %d total=%d is not a multiple of %d (missing boards or mixed orbit)",
			n, k, tot, base))
	}
	mult := tot / base
	if want := orbitSize(k, 2+n); int64(want) != mult {
		panic(fmt.Sprintf("n=%d key %d: total=%d implies orbit %d, key says orbit %d",
			n, k, tot, mult, want))
	}
	if w%mult != 0 || t%mult != 0 {
		panic(fmt.Sprintf("n=%d key %d counts are not orbit-uniform (mult=%d, w=%d, t=%d)",
			n, k, mult, w, t))
	}
	return row{k, uint32(w / mult), uint32(t / mult)}
}

func normalise64(n int, sh []shard64) []row {
	var entries int
	for i := range sh {
		entries += len(sh[i].m)
	}
	out := make([]row, 0, entries)
	for i := range sh {
		for k, v := range sh[i].m {
			out = append(out, normaliseRow(n, k, v.W, v.T, v.N))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out
}

func normalise32(n int, sh []shard32) []row {
	var entries int
	for i := range sh {
		entries += len(sh[i].m)
	}
	out := make([]row, 0, entries)
	for i := range sh {
		for k, v := range sh[i].m {
			out = append(out, normaliseRow(n, k, int64(v.W), int64(v.T), int64(v.N)))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out
}

// writeTables emits the master tables. The runtime TXV1 + zstd files are
// derived from them by cmd/packcolv, so a change of runtime encoding never
// requires re-running the sweep.
//
// Master layout (master_winrate_N.bin, fixed width, 16 bytes per row):
//
//	magic "TXM1"  n u8  reserved[3]  rows u64  total u64
//	rows: key u64  win u32  tie u32
func writeTables(a *acc, dir string, maxN int) {
	for n := 0; n <= maxN; n++ {
		var rows []row
		if n <= 3 {
			rows = normalise64(n, a.shards64(n))
		} else {
			rows = normalise32(n, a.shards32(n))
		}
		writeMaster(rows, baseTotals[n], n, filepath.Join(dir, fmt.Sprintf("master_winrate_%d.bin", n)))
		rows = nil
		a.freeLevel(n)
		runtime.GC()
	}
}

func writeMaster(rows []row, total int64, n int, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	bw := bufio.NewWriterSize(f, 1<<20)
	var hdr [24]byte
	copy(hdr[0:4], "TXM1")
	hdr[4] = byte(n)
	binary.LittleEndian.PutUint64(hdr[8:16], uint64(len(rows)))
	binary.LittleEndian.PutUint64(hdr[16:24], uint64(total))
	bw.Write(hdr[:])
	var rec [16]byte
	for _, r := range rows {
		binary.LittleEndian.PutUint64(rec[0:8], uint64(r.key))
		binary.LittleEndian.PutUint32(rec[8:12], r.w)
		binary.LittleEndian.PutUint32(rec[12:16], r.t)
		bw.Write(rec[:])
	}
	bw.Flush()
	f.Close()
	if fi, err := os.Stat(path); err == nil {
		fmt.Printf("wrote %s rows=%d bytes=%d\n", path, len(rows), fi.Size())
	}
}

func dump(a *acc, dir string) {
	for n := 0; n <= a.maxN; n++ {
		var rows, bad int
		sample := func(k, w, t, tot int64) {
			rows++
			// Every (hole pair, board) contributes exactly 990 opponent
			// evaluations. A partial sweep therefore leaves totals divisible by
			// 990 and consistent with wins+ties; the exact orbit assertion is
			// only possible once the whole board space has been covered, and is
			// done by writeTables instead.
			if tot%990 != 0 || w+t > tot {
				bad++
			}
		}
		if n <= 3 {
			for i := range a.shards64(n) {
				for k, v := range a.shards64(n)[i].m {
					sample(k, v.W, v.T, v.N)
				}
			}
		} else {
			for i := range a.shards32(n) {
				for k, v := range a.shards32(n)[i].m {
					sample(k, int64(v.W), int64(v.T), int64(v.N))
				}
			}
		}
		fmt.Printf("smoke level %d rows=%d badTotal=%d\n", n, rows, bad)
	}
}
