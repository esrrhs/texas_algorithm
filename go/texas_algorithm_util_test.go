package texas_algorithm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConversions(t *testing.T) {
	original := "方A,黑K,红10,梅J,黑9"
	pokes, err := StrToPokes(original)
	if err != nil {
		t.Fatalf("StrToPokes(%q) error: %v", original, err)
	}
	if len(pokes) != 5 {
		t.Fatalf("len = %d, want 5", len(pokes))
	}

	reconstructed := PokesToStr(pokes)
	if reconstructed != "方A黑K红10梅J黑9" {
		t.Errorf("PokesToStr = %q, want %q", reconstructed, "方A黑K红10梅J黑9")
	}

	key := GenCardBindBytes(pokes)
	fromKey := KeyToPoke(key)
	if len(fromKey) != 5 {
		t.Fatalf("len = %d, want 5", len(fromKey))
	}
}

func TestGetMax(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping TestGetMax: data files not present")
	}
	best, err := GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6")
	if err != nil {
		t.Fatalf("GetMaxStrHandPub error: %v", err)
	}
	if best == "" {
		t.Fatalf("best is empty")
	}
	// Best 5 cards should form a straight: 2, 3, 4, 5, 6
	for _, want := range []string{"2", "3", "4", "5", "6"} {
		if !strings.Contains(best, want) {
			t.Errorf("best = %q, want it to contain %q", best, want)
		}
	}
}

func TestTableLookup(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping TestTableLookup: data files not present")
	}

	cards := "方4,方A,黑2,黑A,黑3,黑5,黑6"
	cards1 := "红8,方A,方2,黑8,黑3,黑5,黑7"

	pos, err := GetWinPositionStr(cards)
	if err != nil {
		t.Fatalf("GetWinPositionStr error: %v", err)
	}
	if pos <= 0 {
		t.Errorf("pos = %d, want > 0", pos)
	}
	if pos != 4010 {
		t.Errorf("pos = %d, want 4010", pos)
	}

	prob, err := GetWinProbabilityStr(cards)
	if err != nil {
		t.Fatalf("GetWinProbabilityStr error: %v", err)
	}
	if prob <= 0.8 || prob >= 1.0 {
		t.Errorf("prob = %f, want in (0.8, 1.0)", prob)
	}

	typ, err := GetWinTypeStr(cards)
	if err != nil {
		t.Fatalf("GetWinTypeStr error: %v", err)
	}
	if typ != TexasCardTypeTongHua {
		t.Errorf("type = %d, want %d", typ, TexasCardTypeTongHua)
	}

	pos1, err := GetWinPositionStr(cards1)
	if err != nil {
		t.Fatalf("GetWinPositionStr error: %v", err)
	}
	if pos1 != 1143 {
		t.Errorf("pos1 = %d, want 1143", pos1)
	}
	typ1, err := GetWinTypeStr(cards1)
	if err != nil {
		t.Fatalf("GetWinTypeStr error: %v", err)
	}
	if typ1 != TexasCardTypeDuiZi {
		t.Errorf("type1 = %d, want %d", typ1, TexasCardTypeDuiZi)
	}

	cmp, err := CompareStr(cards, cards1)
	if err != nil {
		t.Fatalf("CompareStr error: %v", err)
	}
	if cmp <= 0 {
		t.Errorf("cards should beat cards1, cmp = %d", cmp)
	}
}

func TestProbabilityEstimation(t *testing.T) {
	if !probDataAvailable {
		t.Skip("Skipping TestProbabilityEstimation: prob data files not present")
	}

	p1, err := GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K")
	if err != nil {
		t.Fatalf("GetHandProbabilityStr error: %v", err)
	}
	if p1 <= 0.55 || p1 >= 0.56 {
		t.Errorf("p1 should be around 0.5533 (exact), got: %f", p1)
	}

	p2, err := GetHandProbabilityStr("方2,方3", "")
	if err != nil {
		t.Fatalf("GetHandProbabilityStr error: %v", err)
	}
	if p2 <= 0.35 || p2 >= 0.37 {
		t.Errorf("p2 should be around 0.3598 (exact), got: %f", p2)
	}
}

func TestUnloadedProbabilityGraceful(t *testing.T) {
	// A wrong number of hole cards must be reported as an error, never a panic.
	if _, err := GetHandProbability([]byte{0x02}, nil); err == nil {
		t.Error("GetHandProbability with one hole card should fail")
	}
}

// TestMissingWinRateTableGraceful checks that a level with no table in memory
// and no sqlite database falls back to on-the-fly enumeration instead of
// failing or panicking.
func TestMissingWinRateTableGraceful(t *testing.T) {
	if !dataAvailable {
		t.Skip("data files not present")
	}
	saved := WinRateMap
	WinRateMap = [6]*WinRateTable{}
	defer func() { WinRateMap = saved }()

	hand := []byte{0x02, 0x03}
	// A flop with no table is enumerated in tens of milliseconds.
	p, err := GetHandProbability(hand, []byte{0x04, 0x05, 0x06})
	if err != nil {
		t.Fatalf("flop should be enumerated without tables: %v", err)
	}
	if p <= 0 || p >= 1 {
		t.Errorf("flop probability out of range: %f", p)
	}

	p, err = GetHandProbability(hand, []byte{0x04, 0x05, 0x06, 0x07})
	if err != nil {
		t.Fatalf("river should be enumerated without tables: %v", err)
	}
	if p <= 0 || p >= 1 {
		t.Errorf("river probability out of range: %f", p)
	}
}

// TestCanonKeySuitInvariant checks that CanonKey is invariant under all 24
// suit permutations for 2 hole cards plus 0-2 public cards, that it separates
// the hole/public roles, and that the preflop table has the 169 known classes.
func TestCanonKeySuitInvariant(t *testing.T) {
	perms := [][4]byte{}
	p := []byte{0, 1, 2, 3}
	var rec func(k int)
	rec = func(k int) {
		if k == 4 {
			var a [4]byte
			copy(a[:], p)
			perms = append(perms, a)
			return
		}
		for i := k; i < 4; i++ {
			p[k], p[i] = p[i], p[k]
			rec(k + 1)
			p[k], p[i] = p[i], p[k]
		}
	}
	rec(0)

	var deck []byte
	for c := 0; c < 4; c++ {
		for v := 2; v <= 14; v++ {
			deck = append(deck, uint8(c<<4|v))
		}
	}
	for n := 0; n <= 2; n++ {
		for a := 0; a < 52; a++ {
			for b := a + 1; b < 52; b++ {
				base := []byte{deck[a], deck[b]}
				for c := 0; c < 52 && n > 0; c++ {
					if c == a || c == b {
						continue
					}
					base = append(base[:2], deck[c])
					want := CanonKey(base, 2)
					for _, pm := range perms {
						cs := make([]byte, len(base))
						for i, x := range base {
							cs[i] = pm[x>>4]<<4 | (x & 0x0f)
						}
						if got := CanonKey(cs, 2); got != want {
							t.Fatalf("n=%d cards=%v perm=%v: %d != %d", n, base, pm, got, want)
						}
					}
				}
				if n == 0 {
					CanonKey(base, 2)
				}
			}
		}
	}

	// role separation: same four cards, different hole/public split
	c1 := []byte{0x02, 0x0d, 0x13, 0x2c}
	c2 := []byte{0x02, 0x13, 0x0d, 0x2c}
	if CanonKey(c1, 2) == CanonKey(c2, 2) {
		t.Errorf("hole/public split is not part of the key: both %d", CanonKey(c1, 2))
	}

	keys := make(map[int64]bool)
	for a := 0; a < 52; a++ {
		for b := a + 1; b < 52; b++ {
			keys[CanonKey([]byte{deck[a], deck[b]}, 2)] = true
		}
	}
	if len(keys) != 169 {
		t.Errorf("distinct preflop keys = %d, want 169", len(keys))
	}
}

// TestEquityExactReference compares the exact enumeration with values produced
// by an independent brute-force evaluator (see the verification in the MR).
func TestEquityExactReference(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping TestEquityExactReference: data files not present")
	}
	cases := []struct {
		hand, pub string
		w, ti, l  int64
	}{
		{"方10,方5", "方4,黑4,方3,黑6", 19465, 1470, 24605},
		{"黑A,黑6", "红Q,黑4,黑K,黑10,红6", 990, 0, 0},
	}
	for _, c := range cases {
		hand, err := StrToPokes(c.hand)
		if err != nil {
			t.Fatalf("StrToPokes(%q) error: %v", c.hand, err)
		}
		pub, err := StrToPokes(c.pub)
		if err != nil {
			t.Fatalf("StrToPokes(%q) error: %v", c.pub, err)
		}
		w, ti, l := EquityExact(hand, pub)
		if w != c.w || ti != c.ti || l != c.l {
			t.Errorf("%s | %s: got (%d,%d,%d), want (%d,%d,%d)", c.hand, c.pub, w, ti, l, c.w, c.ti, c.l)
		}
		p, err := GetHandProbability(hand, pub)
		if err != nil {
			t.Fatalf("GetHandProbability error: %v", err)
		}
		want := float32(2*c.w+c.ti) / float32(2*(c.w+c.ti+c.l))
		if p != want {
			t.Errorf("%s | %s: p = %f, want %f", c.hand, c.pub, p, want)
		}
	}
}

// TestWinRateTableMatchesEnumeration checks that the precomputed n=2 table
// agrees with the on-the-fly enumeration (slow, skipped in -short mode).
func TestWinRateTableMatchesEnumeration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestWinRateTableMatchesEnumeration in short mode")
	}
	if !dataAvailable || !IsWinRateLoaded() {
		t.Skip("Skipping TestWinRateTableMatchesEnumeration: data files not present")
	}
	hand, err := StrToPokes("方3,梅10")
	if err != nil {
		t.Fatalf("StrToPokes error: %v", err)
	}
	pub, err := StrToPokes("黑2,红4")
	if err != nil {
		t.Fatalf("StrToPokes error: %v", err)
	}
	p, err := GetHandProbability(hand, pub)
	if err != nil {
		t.Fatalf("GetHandProbability error: %v", err)
	}
	w, ti, l := EquityExact(hand, pub)
	want := float32(2*w+ti) / float32(2*(w+ti+l))
	if p != want {
		t.Errorf("table p = %f, enumeration p = %f", p, want)
	}
}

// TestWinRate3TableMatchesEnumeration checks that the n=3 table agrees with the
// on-the-fly enumeration on a handful of flops. This is the correctness gate
// for the N <= 3 lookup path.
func TestWinRate3TableMatchesEnumeration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestWinRate3TableMatchesEnumeration in short mode")
	}
	if !dataAvailable || WinRateMap[3].Len() == 0 {
		t.Skip("Skipping TestWinRate3TableMatchesEnumeration: data files not present")
	}
	cases := []struct{ hand, pub string }{
		{"方3,梅10", "黑2,红4,方7"},
		{"方A,方K", "方Q,红9,梅5"},
		{"黑2,红7", "梅3,方8,红J"},
	}
	for _, c := range cases {
		hand, err := StrToPokes(c.hand)
		if err != nil {
			t.Fatalf("StrToPokes(%q) error: %v", c.hand, err)
		}
		pub, err := StrToPokes(c.pub)
		if err != nil {
			t.Fatalf("StrToPokes(%q) error: %v", c.pub, err)
		}
		p, err := GetHandProbability(hand, pub)
		if err != nil {
			t.Fatalf("GetHandProbability error: %v", err)
		}
		w, ti, l := EquityExact(hand, pub)
		want := float32(2*w+ti) / float32(2*(w+ti+l))
		if p != want {
			t.Errorf("%s | %s: table p = %f, enumeration p = %f", c.hand, c.pub, p, want)
		}
	}
}

// TestSqliteMatchesEnumeration checks the on-demand database path: a row read
// from the packed SQLite file must agree with the in-memory table, a level
// outside the loaded range must enumerate to the same value, and a missing
// file must be rejected instead of panicking. The database is about 2.3 GiB and
// is generated locally with cmd/packsqlite, so the test skips when it is absent.
func TestSqliteMatchesEnumeration(t *testing.T) {
	dbPath := filepath.Join(dataDir(), "texas_data_winrate.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Skip("Skipping TestSqliteMatchesEnumeration: packed database not present")
	}

	if err := LoadSqlite(dbPath, 3, 3); err != nil {
		t.Fatalf("LoadSqlite(%q, 3, 3) error: %v", dbPath, err)
	}
	defer closeSqlite()

	saved := WinRateMap[3]
	defer func() { WinRateMap[3] = saved }()

	WinRateMap[3] = nil
	fromDB, err := GetHandProbabilityStr("方3,梅10", "黑2,红4,方7")
	if err != nil {
		t.Fatalf("sqlite lookup error: %v", err)
	}
	WinRateMap[3] = saved
	fromTable, err := GetHandProbabilityStr("方3,梅10", "黑2,红4,方7")
	if err != nil {
		t.Fatalf("table lookup error: %v", err)
	}
	if fromDB != fromTable {
		t.Errorf("sqlite p = %f, table p = %f", fromDB, fromTable)
	}

	// Level 3 lies outside 0..2, so the lookup must enumerate; the exact answer
	// still has to match the table value.
	if err := LoadSqlite(dbPath, 0, 2); err != nil {
		t.Fatalf("LoadSqlite(%q, 0, 2) error: %v", dbPath, err)
	}
	WinRateMap[3] = nil
	fromEnum, err := GetHandProbabilityStr("方3,梅10", "黑2,红4,方7")
	if err != nil {
		t.Fatalf("enumeration fallback error: %v", err)
	}
	if fromEnum != fromTable {
		t.Errorf("level outside the sqlite range: p = %f, table p = %f", fromEnum, fromTable)
	}

	if err := LoadSqlite(filepath.Join(dataDir(), "texas_data_winrate_missing.db"), 0, 5); err == nil {
		t.Errorf("LoadSqlite on a missing file should fail")
	}
}

// benchCase holds a fixed situation per level so the benchmark measures only
// the cost of the query, not the construction of the inputs.
var benchCases = []struct {
	hand, pub string
}{
	{"方3,梅10", ""},
	{"方3,梅10", "黑2"},
	{"方3,梅10", "黑2,红4"},
	{"方3,梅10", "黑2,红4,方7"},
	{"方3,梅10", "黑2,红4,方7,梅9"},
	{"方3,梅10", "黑2,红4,方7,梅9,红J"},
}

// benchGetHandProbability reports the full cost of one GetHandProbability call
// for the given level. It is used to publish the per-level lookup latency in
// the README, so it must stay a plain call with no amortisation.
func benchGetHandProbability(b *testing.B, n int) {
	if !dataAvailable {
		b.Skip("data files not present")
	}
	hand, err := StrToPokes(benchCases[n].hand)
	if err != nil {
		b.Fatal(err)
	}
	var pub []byte
	if benchCases[n].pub != "" {
		pub, err = StrToPokes(benchCases[n].pub)
		if err != nil {
			b.Fatal(err)
		}
	}
	if n <= 3 {
		if WinRateMap[n].Len() == 0 {
			b.Skipf("winrate_%d not loaded", n)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := GetHandProbability(hand, pub); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetHandProbabilityN0(b *testing.B) { benchGetHandProbability(b, 0) }
func BenchmarkGetHandProbabilityN1(b *testing.B) { benchGetHandProbability(b, 1) }
func BenchmarkGetHandProbabilityN2(b *testing.B) { benchGetHandProbability(b, 2) }
func BenchmarkGetHandProbabilityN3(b *testing.B) { benchGetHandProbability(b, 3) }
func BenchmarkGetHandProbabilityN4(b *testing.B) { benchGetHandProbability(b, 4) }
func BenchmarkGetHandProbabilityN5(b *testing.B) { benchGetHandProbability(b, 5) }
