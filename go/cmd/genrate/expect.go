package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	ta "github.com/esrrhs/texas_algorithm/go"
)

// expectCounts independently derives how many canonical keys each level must
// hold and compares that with the number of rows decoded from the packed table,
// so a table that silently dropped keys cannot pass.
//
// Two derivations are used, because they trade rigour for cost:
//
//   - brute: enumerate every real 2-hole + n-public situation, canonicalise it
//     with the public CanonKey and count the distinct keys. It shares no code
//     with the generator's scan-and-accumulate pass, but it needs one CanonKey
//     call per situation: 1.6M for level 2, 26M for level 3, 305M for level 4
//     and 2.8B for level 5, so only the cheap levels use it.
//
//   - formula: a key is exactly the multiset of the four per-suit
//     (rank mask, hole subset) signatures, so the number of keys is the number
//     of such multisets with 2+n ranks and 2 hole cards. Burnside's lemma over
//     the four signature slots gives that in closed form; no card is ever
//     enumerated. This is what makes levels 4 and 5 checkable at all.
//
// The report names the source of every level, because "brute" and "formula"
// do not have the same strength: brute would catch a wrong key definition too,
// while formula only proves the table holds every key the definition admits.
func expectCounts(dir string, maxN, bruteMax int) {
	fmt.Printf("%-5s %-9s %-13s %-13s %-13s %-10s %s\n",
		"level", "source", "derived", "formula", "table", "situations", "result")
	ok := true
	for n := 0; n <= maxN; n++ {
		path := filepath.Join(dir, fmt.Sprintf("texas_data_winrate_%d.zst", n))

		t0 := time.Now()
		table, err := ta.CountWinRateFile(n, path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		decodeTime := time.Since(t0)

		t1 := time.Now()
		formula := formulaKeys(n)
		formulaTime := time.Since(t1)

		derived, source, situations := formula, "formula", int64(-1)
		if n <= bruteMax {
			t2 := time.Now()
			derived, situations = bruteKeys(n)
			bruteTime := time.Since(t2)
			source = "brute"
			if derived != formula {
				fmt.Printf("level %d: brute %d and formula %d disagree\n", n, derived, formula)
				ok = false
			}
			fmt.Printf("%-5d %-9s %-13d %-13d %-13d %-10d %-10s (brute %.1fs, decode %.1fs, formula %s)\n",
				n, source, derived, formula, table.Rows, situations, verdict(derived == int(table.Rows)),
				bruteTime.Seconds(), decodeTime.Seconds(), formulaTime)
			if derived != int(table.Rows) {
				fmt.Fprintf(os.Stderr, "level %d: expected %d keys, the table has %d\n",
					n, derived, table.Rows)
				ok = false
			}
			continue
		}
		fmt.Printf("%-5d %-9s %-13d %-13d %-13d %-10s %-10s (decode %.1fs, formula %s)\n",
			n, source, derived, formula, table.Rows, "not run", verdict(derived == int(table.Rows)),
			decodeTime.Seconds(), formulaTime)
		if derived != int(table.Rows) {
			fmt.Fprintf(os.Stderr, "level %d: expected %d keys, the table has %d\n",
				n, derived, table.Rows)
			ok = false
		}
	}

	if !ok {
		fmt.Fprintf(os.Stderr, "expect: row counts differ (derived = independent derivation, table = decoded from .zst)\n")
		os.Exit(1)
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	bruteHi := minInt(bruteMax, maxN)
	formulaLo := bruteHi + 1
	scope := fmt.Sprintf("brute 0..%d", bruteHi)
	if bruteHi < 0 {
		scope = "no brute level"
	}
	if formulaLo <= maxN {
		scope += fmt.Sprintf(", formula %d..%d", formulaLo, maxN)
	}
	fmt.Printf("expect: %d levels match (%s), heap %d MB\n", maxN+1, scope, ms.HeapAlloc>>20)
	fmt.Printf("expect: brute levels enumerate real cards and dedupe; formula levels count signature multisets in closed form, independent of the scan code but not of the key definition\n")
}

func verdict(same bool) string {
	if same {
		return "ok"
	}
	return "MISMATCH"
}

// bruteKeys enumerates every real situation of level n and returns the number
// of distinct canonical keys plus the number of situations visited.
func bruteKeys(n int) (int, int64) {
	deck := make([]byte, 0, 52)
	for s := byte(0); s < 4; s++ {
		for r := byte(2); r <= 14; r++ {
			deck = append(deck, s<<4|r)
		}
	}

	cards := make([]byte, 2+n)
	seen := make(map[int64]struct{})
	var situations int64
	var walk func(pos, lo, i, j int)
	walk = func(pos, lo, i, j int) {
		if pos == len(cards) {
			seen[ta.CanonKey(cards, 2)] = struct{}{}
			situations++
			return
		}
		for k := lo; k < 52; k++ {
			if k == i || k == j {
				continue
			}
			cards[pos] = deck[k]
			walk(pos+1, k+1, i, j)
		}
	}
	for i := 0; i < 52; i++ {
		for j := i + 1; j < 52; j++ {
			cards[0], cards[1] = deck[i], deck[j]
			walk(2, 0, i, j)
		}
	}
	return len(seen), situations
}

// formulaKeys counts the distinct canonical keys of level n without enumerating
// cards. A key is the multiset of signatures of the four suits, where a suit
// signature is the set of ranks it holds plus the subset of those that are hole
// cards. Hole and public roles stay separate, so a key is determined by such a
// multiset with 2+n mask bits and 2 hole bits in total.
//
// The count of multisets of four signatures is the Burnside average over the
// 24 permutations of the four slots: a permutation fixes only tuples whose
// signatures are equal inside each cycle, so the identity contributes A^4, a
// transposition A^2*B, two transpositions B^2, a 3-cycle A*C and a 4-cycle D,
// with B, C and D made of signatures repeated 2, 3 and 4 times.
func formulaKeys(n int) int {
	const maxMask, maxHole = 52, 8
	type poly [maxMask + 1][maxHole + 1]int64

	mul := func(p, q *poly) (r poly) {
		for a1 := 0; a1 <= maxMask; a1++ {
			for h1 := 0; h1 <= maxHole; h1++ {
				c1 := p[a1][h1]
				if c1 == 0 {
					continue
				}
				for a2 := 0; a2 <= maxMask-a1; a2++ {
					for h2 := 0; h2 <= maxHole-h1; h2++ {
						if c2 := q[a2][h2]; c2 != 0 {
							r[a1+a2][h1+h2] += c1 * c2
						}
					}
				}
			}
		}
		return r
	}

	// A: one suit picking a ranks and h of them as hole cards. Hole counts above
	// 2 can never reach the target total of 2, so they are dropped here.
	var a poly
	for m := 0; m <= 13; m++ {
		for h := 0; h <= m && h <= 2; h++ {
			a[m][h] = int64(binom(13, m) * binom(m, h))
		}
	}
	// repeat builds the polynomial of one signature repeated t times.
	repeat := func(t int) (r poly) {
		for m := 0; m <= 13; m++ {
			for h := 0; h <= m && h <= 2; h++ {
				if a[m][h] != 0 {
					r[t*m][t*h] = a[m][h]
				}
			}
		}
		return r
	}

	a2 := mul(&a, &a)
	a4 := mul(&a2, &a2)
	b := repeat(2)
	c := repeat(3)
	d := repeat(4)
	p22 := mul(&a2, &b)
	pbb := mul(&b, &b)
	pac := mul(&a, &c)

	m := 2 + n
	burnside := a4[m][2] + 6*p22[m][2] + 3*pbb[m][2] + 8*pac[m][2] + 6*d[m][2]
	return int(burnside / 24)
}

func binom(n, k int) int {
	if k > n {
		return 0
	}
	r := 1
	for i := 1; i <= k; i++ {
		r = r * (n - k + i) / i
	}
	return r
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
