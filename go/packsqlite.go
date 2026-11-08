package texas_algorithm

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/klauspost/compress/zstd"
)

// PackSqlite writes the exact win-rate tables of levels from..to (inclusive)
// into a packed SQLite database at out. The file is meant for the "SQLite"
// mode: it is queried on demand, so only the page cache is resident, at the
// cost of a file about 50 times the size of the zstd data.
//
// The schema is clustered and carries no secondary index:
//
//	CREATE TABLE w{n}(k INTEGER PRIMARY KEY, w INTEGER, t INTEGER) WITHOUT ROWID;
//
// Rows are inserted in ascending k order, which is the primary-key order of a
// WITHOUT ROWID table, so the b-tree stays compact and every lookup is a
// single primary-key descent. total is not stored: it is the constant
// winRateTotal(n) derived from the orbit size, so each row only pays for win
// and tie.
//
// The packer is resumable at level granularity: it works on out+".tmp", skips
// a level whose table already holds the expected number of rows, and only
// renames the temporary file onto out after every level is complete. A crash
// therefore never leaves a half-table that is mistaken for the real one.
// Output is deterministic: the schema, the insert order and the page size are
// fixed, so the same inputs produce the same database content.
//
// Level 5 is decoded fully in memory (~2 GB) while it is inserted. The packer
// is a local one-off tool; the library itself never does this.
func PackSqlite(dir, out string, from, to int) error {
	if from < 0 {
		from = 0
	}
	if to >= 6 {
		to = 5
	}
	if from > to {
		return fmt.Errorf("empty level range %d..%d", from, to)
	}

	// out itself is never touched until the temporary file is complete, so an
	// interrupted run cannot corrupt a previously packed database.
	tmp := out + ".tmp"
	db, err := sql.Open("sqlite", tmp)
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	pragmas := []string{
		"PRAGMA page_size=4096",
		"PRAGMA auto_vacuum=NONE",
		"PRAGMA journal_mode=OFF",
		"PRAGMA synchronous=OFF",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA cache_size=-65536",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
	}

	for n := from; n <= to; n++ {
		path := filepath.Join(dir, winRateFileName(n))
		rows, total, err := winRateHeader(n, path)
		if err != nil {
			return err
		}
		if total != uint64(winRateTotal(n)) {
			return fmt.Errorf("%s: total %d does not match the derived %d", path, total, winRateTotal(n))
		}

		table := "w" + strconv.Itoa(n)
		var have int64
		_ = db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&have)
		if uint64(have) == rows {
			fmt.Printf("level %d: already packed (%d rows)\n", n, rows)
			continue
		}
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
		if _, err := db.Exec("CREATE TABLE " + table + "(k INTEGER PRIMARY KEY, w INTEGER NOT NULL, t INTEGER NOT NULL) WITHOUT ROWID"); err != nil {
			return fmt.Errorf("create %s: %w", table, err)
		}

		fmt.Printf("level %d: %d rows, decoding %s\n", n, rows, path)
		if err := LoadWinRateFile(n, path); err != nil {
			return err
		}
		t := WinRateMap[n]
		if t == nil || uint64(t.Len()) != rows {
			return fmt.Errorf("level %d: decoded %d rows, expected %d", n, t.Len(), rows)
		}
		if err := insertWinRateRows(db, table, t); err != nil {
			return err
		}

		// Free the decoded table before the next level is decoded: holding two
		// multi-hundred-megabyte tables at once is pointless in a one-off packer.
		WinRateMap[n] = nil
		t = nil
		runtime.GC()
		fmt.Printf("level %d: packed\n", n)
	}

	if err := db.Close(); err != nil {
		return err
	}
	// os.Rename is atomic on every supported platform, but Windows refuses to
	// replace an existing target, so remove it first there.
	if err := os.Rename(tmp, out); err != nil {
		if rmErr := os.Remove(out); rmErr != nil && !os.IsNotExist(rmErr) {
			return fmt.Errorf("replace %s: %w", out, err)
		}
		if err := os.Rename(tmp, out); err != nil {
			return err
		}
	}
	return nil
}

// insertWinRateRows inserts a decoded table in one transaction per batch. A
// single transaction for 123M rows would build an enormous rollback journal in
// memory; small batches keep the packer flat while still avoiding a commit per
// row.
func insertWinRateRows(db *sql.DB, table string, t *WinRateTable) error {
	const batch = 20000
	for start := 0; start < t.Len(); start += batch {
		end := start + batch
		if end > t.Len() {
			end = t.Len()
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare("INSERT INTO " + table + "(k, w, t) VALUES (?, ?, ?)")
		if err != nil {
			tx.Rollback()
			return err
		}
		for i := start; i < end; i++ {
			if _, err := stmt.Exec(t.Keys[i], t.Win[i], t.Tie[i]); err != nil {
				stmt.Close()
				tx.Rollback()
				return fmt.Errorf("insert row %d: %w", i, err)
			}
		}
		stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// winRateHeader reads only the 24-byte header of a runtime file, so the packer
// can validate a level without decoding its rows.
func winRateHeader(n int, path string) (rows, total uint64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return 0, 0, err
	}
	defer dec.Close()

	var hdr [24]byte
	if _, err := dec.Read(hdr[:]); err != nil {
		return 0, 0, fmt.Errorf("%s: %w", path, err)
	}
	if string(hdr[0:4]) != winRateMagic {
		return 0, 0, fmt.Errorf("%s: bad magic %q", path, string(hdr[0:4]))
	}
	if int(hdr[4]) != n {
		return 0, 0, fmt.Errorf("%s: file holds level %d, expected %d", path, hdr[4], n)
	}
	return binary.LittleEndian.Uint64(hdr[8:16]), binary.LittleEndian.Uint64(hdr[16:24]), nil
}
