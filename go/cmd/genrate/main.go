package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	ta "github.com/esrrhs/texas_algorithm/go"
)

// genrate computes exact "hand vs random opponent" equity tables for 2 hole
// cards plus 0, 1 or 2 known community cards.
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

type cnt struct {
	W, T, N int64
}

// ckptVer identifies the canonical-key scheme. Bump it whenever canonKey or the
// accumulation logic changes: a checkpoint written by an older version holds
// incompatible keys and must never be resumed.
const ckptVer = 3

type acc struct {
	M0, M1, M2 map[int64]*cnt
}

func newAcc() *acc {
	return &acc{
		M0: make(map[int64]*cnt, 4096),
		M1: make(map[int64]*cnt, 1<<17),
		M2: make(map[int64]*cnt, 1<<21),
	}
}

func (a *acc) add(m map[int64]*cnt, k, w, t, l int64) {
	p := m[k]
	if p == nil {
		m[k] = &cnt{W: w, T: t, N: w + t + l}
		return
	}
	p.W += w
	p.T += t
	p.N += w + t + l
}

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

			a.add(a.M0, canonKey([]byte{rest[i], rest[j]}, 2), w, eq, l)
			for _, bc := range b {
				a.add(a.M1, canonKey([]byte{rest[i], rest[j], bc}, 2), w, eq, l)
			}
			for x := 0; x < 5; x++ {
				for y := x + 1; y < 5; y++ {
					a.add(a.M2, canonKey([]byte{rest[i], rest[j], b[x], b[y]}, 2), w, eq, l)
				}
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
	mode := flag.String("mode", "gen", "gen|smoke")
	flag.Parse()

	ta.Load()
	fmt.Printf("tables loaded, workers=%d\n", *workers)

	a := newAcc()
	doneE := map[int]bool{}
	if *resume {
		if f, err := os.Open(*ckpt); err == nil {
			var saved struct {
				Ver  int
				Done []int
				A    *acc
			}
			if err := gob.NewDecoder(f).Decode(&saved); err == nil && saved.Ver == ckptVer {
				a = saved.A
				for _, e := range saved.Done {
					doneE[e] = true
				}
				fmt.Printf("resumed: %d boards-largest done, m2=%d entries\n", len(saved.Done), len(a.M2))
			} else {
				fmt.Printf("ignoring checkpoint %s: version mismatch or unreadable\n", *ckpt)
			}
			f.Close()
		}
	}

	var mu sync.Mutex
	merge := func(local *acc) {
		mu.Lock()
		defer mu.Unlock()
		for k, v := range local.M0 {
			a.add(a.M0, k, v.W, v.T, v.N-v.W-v.T)
		}
		for k, v := range local.M1 {
			a.add(a.M1, k, v.W, v.T, v.N-v.W-v.T)
		}
		for k, v := range local.M2 {
			a.add(a.M2, k, v.W, v.T, v.N-v.W-v.T)
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
				local := newAcc()
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
								processBoard(b, local)
								cntLocal++
							}
						}
					}
					boardsDoneMu.Lock()
					boardsDone += int64(cntLocal)
					boardsDoneMu.Unlock()
				}
				merge(local)
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
		fmt.Printf("[%s] e=%d done, boards=%d (%.2f%%), elapsed=%s, eta=%s, m0=%d m1=%d m2=%d\n",
			time.Now().Format("15:04:05"), e, boardsDone, prog*100,
			el.Round(time.Second), eta.Round(time.Second), len(a.M0), len(a.M1), len(a.M2))

		if e%4 == 0 || e == *maxE {
			saveCkpt(*ckpt, a, doneE)
		}
	}
	saveCkpt(*ckpt, a, doneE)

	if *mode == "smoke" {
		dump(a, *out)
		return
	}
	writeTables(a, *out)
	fmt.Printf("done in %s\n", time.Since(t0).Round(time.Second))
}

func saveCkpt(path string, a *acc, done map[int]bool) {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		panic(err)
	}
	var d []int
	for e := range done {
		d = append(d, e)
	}
	sort.Ints(d)
	enc := gob.NewEncoder(f)
	enc.Encode(struct {
		Ver  int
		Done []int
		A    *acc
	}{ckptVer, d, a})
	f.Close()
	os.Rename(tmp, path)
}

// writeTables writes the three files after normalising every row back to a
// single physical situation: a canonical key aggregates a whole suit orbit, so
// its aggregated counts are divided by the orbit size. The resulting total is
// exactly the per-situation enumeration count, which doubles as a self-check:
// any missing board or wrong orbit size makes the division fail.
func writeTables(a *acc, dir string) {
	base := []int64{2097572400, 209757240, 17123040}
	norm := func(n int, m map[int64]*cnt) map[int64]*cnt {
		out := make(map[int64]*cnt, len(m))
		for k, v := range m {
			if v.N == 0 || v.N%base[n] != 0 {
				panic(fmt.Sprintf("n=%d key %d total=%d is not a multiple of %d (missing boards or mixed orbit)",
					n, k, v.N, base[n]))
			}
			mult := v.N / base[n]
			if v.W%mult != 0 || v.T%mult != 0 {
				panic(fmt.Sprintf("n=%d key %d counts are not orbit-uniform (mult=%d, w=%d, t=%d)",
					n, k, mult, v.W, v.T))
			}
			out[k] = &cnt{W: v.W / mult, T: v.T / mult, N: base[n]}
		}
		return out
	}
	writeOne(norm(0, a.M0), filepath.Join(dir, "texas_data_winrate_0.txt"))
	writeOne(norm(1, a.M1), filepath.Join(dir, "texas_data_winrate_1.txt"))
	writeOne(norm(2, a.M2), filepath.Join(dir, "texas_data_winrate_2.txt"))
}

func writeOne(m map[int64]*cnt, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	w := bufio.NewWriterSize(f, 1<<20)
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		v := m[k]
		fmt.Fprintf(w, "%d %d %d %d\n", k, v.W, v.T, v.N)
	}
	w.Flush()
	f.Close()
	fmt.Printf("wrote %s rows=%d\n", path, len(keys))
}

func dump(a *acc, dir string) {
	fmt.Printf("smoke: m0=%d m1=%d m2=%d\n", len(a.M0), len(a.M1), len(a.M2))
	show := func(m map[int64]*cnt, name string) {
		n := 0
		for k, v := range m {
			if n >= 5 {
				break
			}
			fmt.Printf("  %s key=%d p=%.6f (w=%d t=%d n=%d)\n", name, k,
				(float64(v.W)+0.5*float64(v.T))/float64(v.N), v.W, v.T, v.N)
			n++
		}
	}
	show(a.M0, "N0")
	show(a.M1, "N1")
	show(a.M2, "N2")
}
