// Command packsqlite packs the zstd win-rate files into one SQLite database
// for the on-demand "SQLite" loading mode:
//
//	go run ./cmd/packsqlite -dir ../.. -out ../../texas_data_winrate.db
//
// The database is not committed: it is roughly 50 times the size of the zstd
// data, far beyond what belongs in git. Generate it locally, or attach it to a
// release. The command is resumable at level granularity; see PackSqlite.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/esrrhs/texas_algorithm/go"
)

func main() {
	dir := flag.String("dir", ".", "directory holding texas_data_winrate_N.zst")
	out := flag.String("out", "texas_data_winrate.db", "output SQLite database")
	from := flag.Int("from", 0, "first level to pack")
	to := flag.Int("to", 5, "last level to pack")
	flag.Parse()

	if err := texas_algorithm.PackSqlite(*dir, *out, *from, *to); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
