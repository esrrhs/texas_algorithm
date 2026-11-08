package main

import (
	"fmt"
	"os"
	"path/filepath"

	ta "github.com/esrrhs/texas_algorithm/go"
)

// countKeys parses every packed table up to maxN and reports how many keys each
// level holds. The scan streams the file and keeps nothing, so unlike a full
// load it counts level 5 (123M rows) without the ~1.9 GiB of tables.
//
// The tables are read from dir, the same directory genrate writes its output
// to, and are named like the committed data files (texas_data_winrate_N.zst).
func countKeys(dir string, maxN int) {
	var sum uint64
	for n := 0; n <= maxN; n++ {
		c, err := ta.CountWinRateFile(n, filepath.Join(dir, fmt.Sprintf("texas_data_winrate_%d.zst", n)))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("winrate_%d: rows %d, total %d\n", n, c.Rows, c.Total)
		sum += c.Rows
	}
	fmt.Printf("winrate_0..%d: rows %d\n", maxN, sum)
}
