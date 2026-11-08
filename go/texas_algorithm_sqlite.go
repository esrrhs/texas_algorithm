package texas_algorithm

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

	_ "modernc.org/sqlite"
)

// sqliteSource is an open packed win-rate database. Rows stay on disk: every
// query is a primary-key lookup on a WITHOUT ROWID table, so nothing but the
// page cache is resident.
type sqliteSource struct {
	db    *sql.DB
	from  int
	to    int
	stmts [6]*sql.Stmt
	avail [6]bool
}

var (
	sqliteMu  sync.RWMutex
	sqliteSrc *sqliteSource
)

// LoadSqlite opens a packed win-rate database (see cmd/packsqlite) and answers
// levels from..to from it. Rows are queried on demand and never loaded into
// memory. An in-memory table always wins over the database, so the two can
// coexist: LoadWinRateRange(0, 2) plus LoadSqlite(db, 0, 5) reads 0-2 from
// memory and 3-5 from disk.
func LoadSqlite(path string, from, to int) error {
	if from < 0 {
		from = 0
	}
	if to >= len(WinRateMap) {
		to = len(WinRateMap) - 1
	}
	if from > to {
		return fmt.Errorf("empty level range %d..%d", from, to)
	}

	// mode=ro opens the packed file read-only, so a read-only mount still works
	// and no journal or -wal file is ever created next to it. immutable=1
	// promises the file never changes, which lets SQLite skip locking and
	// shared-memory bookkeeping entirely; the packer only produces the file
	// once and then never touches it.
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	db, err := sql.Open("sqlite", u.String()+"?mode=ro&immutable=1")
	if err != nil {
		return err
	}
	// The packed file is immutable; a single connection keeps the page cache
	// small and avoids redundant handles.
	db.SetMaxOpenConns(1)

	src := &sqliteSource{db: db, from: from, to: to}
	for n := from; n <= to; n++ {
		table := "w" + strconv.Itoa(n)
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			db.Close()
			return err
		}
		stmt, err := db.Prepare("SELECT w, t FROM " + table + " WHERE k = ?")
		if err != nil {
			db.Close()
			return err
		}
		src.stmts[n] = stmt
		src.avail[n] = true
	}

	sqliteMu.Lock()
	old := sqliteSrc
	sqliteSrc = src
	sqliteMu.Unlock()
	if old != nil {
		old.close()
	}
	return nil
}

func (s *sqliteSource) close() {
	for _, stmt := range s.stmts {
		if stmt != nil {
			stmt.Close()
		}
	}
	s.db.Close()
}

// closeSqlite releases the packed database, if any. It keeps the test suite
// from leaking the connection between cases.
func closeSqlite() {
	sqliteMu.Lock()
	src := sqliteSrc
	sqliteSrc = nil
	sqliteMu.Unlock()
	if src != nil {
		src.close()
	}
}

// sqliteCounts answers (win, tie) of a situation from the packed database.
// found is false when no database is open, the level is out of the loaded
// range, the table is absent, or the key is not in the table; the caller then
// falls back to enumeration.
func sqliteCounts(n int, hand, pub []byte) (win, tie uint32, found bool) {
	sqliteMu.RLock()
	defer sqliteMu.RUnlock()
	src := sqliteSrc
	if src == nil || n < src.from || n > src.to || !src.avail[n] {
		return 0, 0, false
	}

	cards := append(append([]byte{}, hand...), pub...)
	key := CanonKey(cards, len(hand))
	var w, t int64
	if err := src.stmts[n].QueryRow(key).Scan(&w, &t); err != nil {
		return 0, 0, false
	}
	return uint32(w), uint32(t), true
}
