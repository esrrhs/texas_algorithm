package texas_algorithm

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

// KeyData is the lookup result of a 7-card hand key. Postion keeps Java's
// spelling of the original field (getPostion).
type KeyData struct {
	Index   int
	Postion int
	Max     int64
	Type    int
}

var (
	ColorMap  = map[int64]*KeyData{}
	NormalMap = map[int64]*KeyData{}

	// WinRateMap[n] holds exact "2 hole cards + n known community cards vs
	// one random opponent" counts, keyed by the suit-canonical card key.
	// Levels 0-3 are loaded by default. Level 4 and 5 ship with the repository
	// but stay out of memory unless something asks for them (LoadProbilityAll
	// or LoadWinRateRange), because decoded they need about 223 MiB and
	// 1.97 GiB while the enumeration answers a whole query in ~2 ms and ~0.2 ms.
	WinRateMap [6]*WinRateTable

	// winRateDir is the directory the probability files are read from. It is
	// written by LoadProbilityLevels so a later LoadWinRateRange call reads the
	// same place as the startup load.
	winRateDir = "."
	winRateMu  sync.Mutex
)

// WinRateTable is one exact-count table: rows sorted by canonical key, plus the
// constant per-situation enumeration total from the file header. Rows store
// only win and tie because total is shared by every row.
type WinRateTable struct {
	Keys  []int64
	Win   []uint32
	Tie   []uint32
	Total uint32
}

// Len returns the number of rows, which is also the "loaded" signal: an empty
// table means the file was missing or unreadable.
func (t *WinRateTable) Len() int {
	if t == nil {
		return 0
	}
	return len(t.Keys)
}

// Get returns the counts of a canonical key by binary search.
func (t *WinRateTable) Get(key int64) (win, tie, total uint32, ok bool) {
	if t == nil {
		return 0, 0, 0, false
	}
	lo, hi := 0, len(t.Keys)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		if t.Keys[mid] < key {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= len(t.Keys) || t.Keys[lo] != key {
		return 0, 0, 0, false
	}
	return t.Win[lo], t.Tie[lo], t.Total, true
}

// Gen generates texas_data.txt if it does not exist yet.
func Gen() {
	if !fileExists("texas_data.txt") {
		GenKey()
		OutputData()
	}
}

// GenExtra generates texas_data_extra_6.txt and texas_data_extra_5.txt if they
// do not exist yet.
func GenExtra() {
	if !fileExists("texas_data_extra_6.txt") {
		GenExtraN = 6
		GenExtraGenKey()
		GenExtraOutputData()
	}
	if !fileExists("texas_data_extra_5.txt") {
		GenExtraN = 5
		GenExtraGenKey()
		GenExtraOutputData()
	}
}

// GenOpt generates texas_data_color.txt and texas_data_normal.txt from
// texas_data.txt if they do not exist yet.
func GenOpt() {
	if fileExists("texas_data.txt") && !fileExists("texas_data_color.txt") && !fileExists("texas_data_normal.txt") {
		OptNormalData()
		OptColorData()
	}
}

// GenExtraOpt generates the extra optimized tables if they do not exist yet.
func GenExtraOpt() {
	if fileExists("texas_data_extra_6.txt") && !fileExists("texas_data_extra_color_6.txt") && !fileExists("texas_data_extra_normal_6.txt") {
		GenExtraOptN = 6
		GenExtraOptNormalData()
		GenExtraOptColorData()
	}
	if fileExists("texas_data_extra_5.txt") && !fileExists("texas_data_extra_color_5.txt") && !fileExists("texas_data_extra_normal_5.txt") {
		GenExtraOptN = 5
		GenExtraOptNormalData()
		GenExtraOptColorData()
	}
}

// GenHand generates the exhaustive 1v1 win-rate files for 0 to 4 public cards.
func GenHand() {
	for i := 0; i <= 4; i++ {
		GenHandN = i
		GenHandGenKey()
	}
}

// GenTrans generates texas_data_N.txt (N = 6..2) from texas_data.txt.
func GenTrans() {
	if !fileExists("texas_data.txt") {
		return
	}
	for i := 6; i >= 2; i-- {
		if !fileExists("texas_data_" + strconv.Itoa(i) + ".txt") {
			GenTransN = i
			GenTransGenKey()
			GenTransTransData()
		}
	}
}

// GenTransOpt generates texas_data_opt_N.txt (N = 6..2).
func GenTransOpt() {
	if !fileExists("texas_data.txt") {
		return
	}
	for i := 6; i >= 2; i-- {
		if fileExists("texas_data_"+strconv.Itoa(i)+".txt") && !fileExists("texas_data_opt_"+strconv.Itoa(i)+".txt") {
			GenTransOptN = i
			GenTransOptData()
		}
	}
}

// IsLoaded reports whether the lookup tables are loaded.
func IsLoaded() bool {
	return len(NormalMap) != 0 && len(ColorMap) != 0
}

// IsProbabilityLoaded reports whether the exact win-rate tables for 0-3 known
// community cards are in memory. Only the N <= 3 lookups depend on them; N = 4
// and 5 are enumerated on the fly and work without any table.
func IsProbabilityLoaded() bool {
	for _, t := range WinRateMap[:4] {
		if t.Len() == 0 {
			return false
		}
	}
	return true
}

// Load loads the lookup tables from the current directory.
func Load() {
	LoadDir(".")
}

// LoadDir loads the lookup tables from the given directory. Like the Java
// version, loading stops at the first missing file.
func LoadDir(dir string) {
	begin := time.Now().UnixMilli()

	files := []struct {
		name  string
		color bool
	}{
		{"texas_data_color.txt", true},
		{"texas_data_normal.txt", false},
		{"texas_data_extra_color_6.txt", true},
		{"texas_data_extra_normal_6.txt", false},
		{"texas_data_extra_color_5.txt", true},
		{"texas_data_extra_normal_5.txt", false},
	}
	for _, f := range files {
		err := loadKeyFile(filepath.Join(dir, f.name), f.color)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("load time", time.Now().UnixMilli()-begin)
}

// LoadProbility loads the probability tables from the current directory.
func LoadProbility() {
	LoadProbilityDir(".")
}

// LoadProbilityDir loads the default set of exact win-rate tables (0 to 3 known
// community cards) from the given directory. The files are zstd-compressed and
// decompressed transparently on load, so nothing has to be unpacked by hand.
// The tables are independent: a missing texas_data_winrate_N.zst only makes
// the matching N-card lookup return ErrNoProbabilityData; everything else keeps
// working.
//
// Levels 4 and 5 are deliberately left out of the default set. Their files ship
// with the repository, but decoded they need about 223 MiB and 1.97 GiB of
// resident memory, while the on-the-fly enumeration already answers a level 4
// query in ~2 ms and a level 5 query in ~0.2 ms. Call LoadProbilityAll or
// LoadWinRateRange(4, 5) to opt in, or load them from a packed SQLite
// database with LoadSqlite, which answers them without touching the heap.
func LoadProbilityDir(dir string) {
	LoadProbilityLevels(dir, 0, 1, 2, 3)
}

// LoadProbilityAll loads all six exact win-rate tables, including levels 4 and
// 5. It is the explicit switch for deployments that prefer the table over the
// enumeration; expect roughly 223 MiB extra for level 4 and 1.97 GiB for level 5.
func LoadProbilityAll(dir string) {
	LoadProbilityLevels(dir, 0, 1, 2, 3, 4, 5)
}

// LoadProbilityLevels loads the given levels from dir. A missing or corrupt
// file is reported and skipped, and every other level keeps working.
func LoadProbilityLevels(dir string, levels ...int) {
	begin := time.Now().UnixMilli()

	winRateMu.Lock()
	winRateDir = dir
	winRateMu.Unlock()

	loaded := 0
	for _, n := range levels {
		if n < 0 || n >= len(WinRateMap) {
			continue
		}
		if err := LoadWinRateFile(n, filepath.Join(dir, winRateFileName(n))); err != nil {
			fmt.Println(err)
			continue
		}
		loaded++
	}

	fmt.Println("load time", time.Now().UnixMilli()-begin, "tables", loaded)
}

// winRateFileName is the runtime file of level n: one zstd-compressed file per
// level, matching the opt_2..opt_6 layout of the original tables.
func winRateFileName(n int) string {
	return "texas_data_winrate_" + strconv.Itoa(n) + ".zst"
}

// winRateMagic is the payload magic of the runtime win-rate files.
const winRateMagic = "TXV1"

// maxWinRateRows is a sanity bound on the row count, about 1.6x the level 5
// table, so a corrupt header cannot ask for an absurd allocation.
const maxWinRateRows = 200_000_000

// LoadWinRateFile loads one exact win-rate table from its zstd-compressed
// runtime file. The payload is columnar:
//
//	"TXV1" (4) | n (1) | reserved (3) | rows (u64 LE) | total (u64 LE)
//	rows x uvarint   key delta against the previous row
//	rows x uvarint   win count
//	rows x uvarint   tie count
//
// `total` is the same for every row of a level: the generator derives it from
// the orbit size and the per-hand enumeration count, so it is stored once in
// the header, and loss = total - win - tie is never written. Keys are strictly
// increasing after delta decoding, so the three columns become three parallel
// arrays that are binary searched.
//
// The zstd stream is decoded incrementally, so the 487 MB decompressed payload
// of level 5 never has to sit in memory on top of the decoded arrays.
func LoadWinRateFile(n int, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f,
		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderMaxWindow(64<<20),
	)
	if err != nil {
		return err
	}
	defer dec.Close()

	in := newStreamReader(dec, 1<<17)

	var hdr [24]byte
	if err := in.readFull(hdr[:]); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if string(hdr[0:4]) != winRateMagic {
		return fmt.Errorf("%s: bad magic %q", path, string(hdr[0:4]))
	}
	if int(hdr[4]) != n {
		return fmt.Errorf("%s: file holds level %d, expected %d", path, hdr[4], n)
	}
	rows := binary.LittleEndian.Uint64(hdr[8:16])
	total := binary.LittleEndian.Uint64(hdr[16:24])
	if total == 0 || total > math.MaxUint32 {
		return fmt.Errorf("%s: total %d does not fit the uint32 rows", path, total)
	}
	if rows == 0 || rows > maxWinRateRows {
		return fmt.Errorf("%s: implausible row count %d", path, rows)
	}
	// total is derived from the orbit size, never stored per row; the header
	// value is only a cross check against the formula the generator uses.
	if total != uint64(winRateTotal(n)) {
		return fmt.Errorf("%s: total %d does not match the derived %d", path, total, winRateTotal(n))
	}

	t := &WinRateTable{
		Keys:  make([]int64, rows),
		Win:   make([]uint32, rows),
		Tie:   make([]uint32, rows),
		Total: uint32(total),
	}

	// The payload is columnar: every key delta first, then every win, then
	// every tie, so the three loops below each read one homogeneous stream.
	var prev int64
	for i := uint64(0); i < rows; i++ {
		gap, err := in.uvarint()
		if err != nil {
			return fmt.Errorf("%s: row %d key: %w", path, i, err)
		}
		key := prev + int64(gap)
		if key <= prev {
			return fmt.Errorf("%s: row %d key %d is not increasing", path, i, key)
		}
		t.Keys[i] = key
		prev = key
	}
	for i := uint64(0); i < rows; i++ {
		w, err := in.uvarint()
		if err != nil {
			return fmt.Errorf("%s: row %d win: %w", path, i, err)
		}
		if w > total {
			return fmt.Errorf("%s: row %d win %d above total %d", path, i, w, total)
		}
		t.Win[i] = uint32(w)
	}
	for i := uint64(0); i < rows; i++ {
		ti, err := in.uvarint()
		if err != nil {
			return fmt.Errorf("%s: row %d tie: %w", path, i, err)
		}
		if uint64(t.Win[i])+ti > total {
			return fmt.Errorf("%s: row %d has win+tie=%d above total %d", path, i, uint64(t.Win[i])+ti, total)
		}
		t.Tie[i] = uint32(ti)
	}

	WinRateMap[n] = t
	return nil
}

// WinRateCount is the row count reported by CountWinRateFile.
type WinRateCount struct {
	Rows  uint64
	Total uint64
}

// CountWinRateFile decodes one packed table from the first key to the last and
// reports how many rows it holds. It is the streaming counterpart of
// LoadWinRateFile: the header and the three columns are validated the same way,
// but no arrays are allocated and WinRateMap is left untouched, so counting
// level 5 costs a few megabytes instead of about 1.9 GiB.
func CountWinRateFile(n int, path string) (WinRateCount, error) {
	f, err := os.Open(path)
	if err != nil {
		return WinRateCount{}, err
	}
	defer f.Close()

	dec, err := zstd.NewReader(f,
		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderMaxWindow(64<<20),
	)
	if err != nil {
		return WinRateCount{}, err
	}
	defer dec.Close()

	in := newStreamReader(dec, 1<<17)

	var hdr [24]byte
	if err := in.readFull(hdr[:]); err != nil {
		return WinRateCount{}, fmt.Errorf("%s: %w", path, err)
	}
	if string(hdr[0:4]) != winRateMagic {
		return WinRateCount{}, fmt.Errorf("%s: bad magic %q", path, string(hdr[0:4]))
	}
	if int(hdr[4]) != n {
		return WinRateCount{}, fmt.Errorf("%s: file holds level %d, expected %d", path, hdr[4], n)
	}
	rows := binary.LittleEndian.Uint64(hdr[8:16])
	total := binary.LittleEndian.Uint64(hdr[16:24])
	if total == 0 || total > math.MaxUint32 {
		return WinRateCount{}, fmt.Errorf("%s: total %d does not fit the uint32 rows", path, total)
	}
	if rows == 0 || rows > maxWinRateRows {
		return WinRateCount{}, fmt.Errorf("%s: implausible row count %d", path, rows)
	}
	if total != uint64(winRateTotal(n)) {
		return WinRateCount{}, fmt.Errorf("%s: total %d does not match the derived %d", path, total, winRateTotal(n))
	}

	// Same column layout as LoadWinRateFile: key deltas, then wins, then ties.
	// Keys must strictly increase and every count must fit inside total; only
	// the running key is kept.
	var prev int64
	for i := uint64(0); i < rows; i++ {
		gap, err := in.uvarint()
		if err != nil {
			return WinRateCount{}, fmt.Errorf("%s: row %d key: %w", path, i, err)
		}
		key := prev + int64(gap)
		if key <= prev {
			return WinRateCount{}, fmt.Errorf("%s: row %d key %d is not increasing", path, i, key)
		}
		prev = key
	}
	for i := uint64(0); i < rows; i++ {
		c, err := in.uvarint()
		if err != nil {
			return WinRateCount{}, fmt.Errorf("%s: row %d win: %w", path, i, err)
		}
		if c > total {
			return WinRateCount{}, fmt.Errorf("%s: row %d win %d above total %d", path, i, c, total)
		}
	}
	for i := uint64(0); i < rows; i++ {
		c, err := in.uvarint()
		if err != nil {
			return WinRateCount{}, fmt.Errorf("%s: row %d tie: %w", path, i, err)
		}
		if c > total {
			return WinRateCount{}, fmt.Errorf("%s: row %d tie %d above total %d", path, i, c, total)
		}
	}
	return WinRateCount{Rows: rows, Total: total}, nil
}

// errVarintOverflow is returned when a uvarint does not fit 64 bits.
var errVarintOverflow = errors.New("varint overflows uint64")

// streamReader buffers a decompression stream and decodes uvarints straight
// out of the buffer. encoding/binary.ReadUvarint would route every byte
// through an interface call, which is far too slow for the 123M-row level 5.
type streamReader struct {
	src io.Reader
	buf []byte
	pos int
	end int
}

func newStreamReader(src io.Reader, size int) *streamReader {
	return &streamReader{src: src, buf: make([]byte, size)}
}

// fill refills the buffer, tolerating readers that return (0, nil).
func (r *streamReader) fill() error {
	for {
		n, err := r.src.Read(r.buf)
		if n > 0 {
			r.pos, r.end = 0, n
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (r *streamReader) readFull(dst []byte) error {
	for len(dst) > 0 {
		if r.pos == r.end {
			if err := r.fill(); err != nil {
				return err
			}
		}
		n := copy(dst, r.buf[r.pos:r.end])
		r.pos += n
		dst = dst[n:]
	}
	return nil
}

func (r *streamReader) uvarint() (uint64, error) {
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if r.pos == r.end {
			if err := r.fill(); err != nil {
				return 0, err
			}
		}
		c := r.buf[r.pos]
		r.pos++
		if c < 0x80 {
			if shift == 63 && c > 1 {
				return 0, errVarintOverflow
			}
			return v | uint64(c)<<shift, nil
		}
		if shift >= 63 {
			return 0, errVarintOverflow
		}
		v |= uint64(c&0x7f) << shift
	}
}

// LoadWinRateRange loads levels from..to (inclusive) of the exact win-rate
// tables into memory, from the directory last given to LoadProbilityLevels
// (current directory by default). LoadWinRateRange(0, 5) is the "everything in
// memory" mode: about 2.2 GiB of resident memory for all six levels.
func LoadWinRateRange(from, to int) error {
	return LoadWinRateRangeDir(winRateDir, from, to)
}

// LoadWinRateRangeDir is LoadWinRateRange with an explicit directory.
func LoadWinRateRangeDir(dir string, from, to int) error {
	if from < 0 {
		from = 0
	}
	if to >= len(WinRateMap) {
		to = len(WinRateMap) - 1
	}
	for n := from; n <= to; n++ {
		if err := LoadWinRateFile(n, filepath.Join(dir, winRateFileName(n))); err != nil {
			return err
		}
	}
	return nil
}

// IsWinRateLoaded reports whether the exact win-rate tables are loaded.
func IsWinRateLoaded() bool {
	return WinRateMap[0].Len() != 0
}

func loadKeyFile(path string, color bool) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			return err
		}
		i, err := strconv.Atoi(params[1])
		if err != nil {
			return err
		}
		index, err := strconv.Atoi(params[2])
		if err != nil {
			return err
		}
		max, err := strconv.ParseInt(params[5], 10, 64)
		if err != nil {
			return err
		}
		typ, err := strconv.Atoi(params[7])
		if err != nil {
			return err
		}

		keyData := &KeyData{Index: index, Postion: i, Max: max, Type: typ}
		if color {
			ColorMap[key] = keyData
		} else {
			NormalMap[key] = keyData
		}
	}
	return scanner.Err()
}

// ErrUnknownPoke is returned when a card string, card byte or hand key cannot
// be recognised as cards of the 52-card deck. It replaces the old silent
// zero/null result that joker input used to produce.
var ErrUnknownPoke = errors.New("无法识别的牌")

// CheckPoke validates one encoded card byte. Only the 52-card deck is legal:
// color must be 0..3 and value 2..14. The joker byte 88 (color 5) is rejected.
func CheckPoke(p byte) error {
	if p>>4 > 3 || p&0x0f < PokeValue2 || p&0x0f > PokeValueA {
		return fmt.Errorf("%w: 牌字节 %d", ErrUnknownPoke, p)
	}
	return nil
}

// CheckPokes validates every card byte in the slice.
func CheckPokes(pokes []byte) error {
	for _, p := range pokes {
		if err := CheckPoke(p); err != nil {
			return err
		}
	}
	return nil
}

// CheckKey validates the cards encoded in a hand key.
func CheckKey(key int64) error {
	for _, c := range keyToCards(key) {
		if err := CheckPoke(byte(c)); err != nil {
			return fmt.Errorf("%w: 牌 key %d", ErrUnknownPoke, key)
		}
	}
	return nil
}

// StrToPokeValue converts a rank string ("2".."10", "J", "Q", "K", "A") into
// its byte value.
func StrToPokeValue(str string) (byte, error) {
	switch str {
	case "A":
		return PokeValueA, nil
	case "K":
		return PokeValueK, nil
	case "Q":
		return PokeValueQ, nil
	case "J":
		return PokeValueJ, nil
	}
	v, err := strconv.ParseInt(str, 10, 8)
	if err != nil || byte(v) < PokeValue2 || byte(v) > PokeValueA {
		return 0, fmt.Errorf("%w: 牌面 %q", ErrUnknownPoke, str)
	}
	return byte(v), nil
}

// StrToPoke converts a card string ("方A", "黑10") into its byte. An empty
// string, an unknown suit or rank, and the removed joker card ("鬼") all
// return ErrUnknownPoke.
func StrToPoke(str string) (byte, error) {
	r := []rune(str)
	if len(r) == 0 {
		return 0, fmt.Errorf("%w: 空牌面", ErrUnknownPoke)
	}
	var color byte
	switch r[0] {
	case '方':
		color = PokeColorFang
	case '梅':
		color = PokeColorMei
	case '红':
		color = PokeColorHong
	case '黑':
		color = PokeColorHei
	default:
		return 0, fmt.Errorf("%w: 牌面 %q", ErrUnknownPoke, str)
	}
	value, err := StrToPokeValue(string(r[1:]))
	if err != nil {
		return 0, fmt.Errorf("%w: 牌面 %q", ErrUnknownPoke, str)
	}
	return NewPoke(color, value).ToByte(), nil
}

// KeyToStr renders a key as its concatenated card strings.
func KeyToStr(key int64) string {
	return ToString(key)
}

// KeyToPoke decodes a key into its card bytes, most significant card first.
func KeyToPoke(k int64) []byte {
	var cs []byte
	if k > 1000000000000 {
		cs = append(cs, byte(k%100000000000000/1000000000000))
	}
	if k > 10000000000 {
		cs = append(cs, byte(k%1000000000000/10000000000))
	}
	if k > 100000000 {
		cs = append(cs, byte(k%10000000000/100000000))
	}
	if k > 1000000 {
		cs = append(cs, byte(k%100000000/1000000))
	}
	if k > 10000 {
		cs = append(cs, byte(k%1000000/10000))
	}
	if k > 100 {
		cs = append(cs, byte(k%10000/100))
	}
	if k > 1 {
		cs = append(cs, byte(k%100/1))
	}
	return cs
}

// StrToPokes converts a comma-separated card string into card bytes. Any
// unrecognised card makes the whole conversion fail with ErrUnknownPoke.
func StrToPokes(str string) ([]byte, error) {
	var ret []byte
	if len(str) == 0 {
		return ret, nil
	}
	for _, s := range strings.Split(str, ",") {
		p, err := StrToPoke(s)
		if err != nil {
			return nil, err
		}
		ret = append(ret, p)
	}
	return ret, nil
}

// PokesToStr renders card bytes as their concatenated card strings.
func PokesToStr(pokes []byte) string {
	return KeyToStr(GenCardBindBytes(pokes))
}

// GetKeyDataStr looks up a 7-card comma-separated string. It returns
// ErrUnknownPoke for an unrecognised card and an error when the string does
// not contain exactly 7 cards.
func GetKeyDataStr(str string) (*KeyData, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return nil, err
	}
	if len(pokes) != 7 {
		return nil, fmt.Errorf("%w: 需要 7 张牌，实际 %d 张", ErrUnknownPoke, len(pokes))
	}
	return GetKeyData(pokes)
}

// GetKeyData looks up the best-hand data of the given cards.
func GetKeyData(pokes []byte) (*KeyData, error) {
	if err := CheckPokes(pokes); err != nil {
		return nil, err
	}
	data := LookupKeyData(GenCardBindBytes(pokes))
	if data == nil {
		return nil, fmt.Errorf("%w: 未找到 key %d 的牌力数据", ErrUnknownPoke, GenCardBindBytes(pokes))
	}
	return data, nil
}

// GetKeyDataKey looks up the best-hand data of the given 7-card key. The key is
// checked against the 52-card deck first, so a key holding the removed joker
// byte 88 is rejected with ErrUnknownPoke.
func GetKeyDataKey(key int64) (*KeyData, error) {
	if err := CheckKey(key); err != nil {
		return nil, err
	}
	data := LookupKeyData(key)
	if data == nil {
		return nil, fmt.Errorf("%w: 未找到 key %d 的牌力数据", ErrUnknownPoke, key)
	}
	return data, nil
}

// LookupKeyData looks up the best-hand data of a 7-card key without validating
// the key. It is the unvalidated hot path used by the table generator, and
// returns nil when the key is absent from the loaded tables.
func LookupKeyData(key int64) *KeyData {
	colorKey := ChangeColor(key)
	color := ColorMap[colorKey]
	normalKey := RemoveColor(key)
	normal := NormalMap[normalKey]
	if color == nil {
		return normal
	}
	if normal == nil {
		return color
	}
	if color.Index > normal.Index {
		return color
	}
	return normal
}

// GetMaxStr returns the best 5-card hand of a comma-separated card list
// (5-7 cards).
func GetMaxStr(str string) (string, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return "", err
	}
	best, err := GetMax(pokes)
	if err != nil {
		return "", err
	}
	return PokesToStr(best), nil
}

// GetMaxStrHandPub returns the best 5-card hand of 2 hole cards and 3-5
// community cards.
func GetMaxStrHandPub(hand, pub string) (string, error) {
	handPokes, err := StrToPokes(hand)
	if err != nil {
		return "", err
	}
	pubPokes, err := StrToPokes(pub)
	if err != nil {
		return "", err
	}
	best, err := GetMaxHandPub(handPokes, pubPokes)
	if err != nil {
		return "", err
	}
	return PokesToStr(best), nil
}

// GetMax returns the best 5-card hand of the given cards (2 hole cards
// followed by 3-5 community cards).
func GetMax(pokes []byte) ([]byte, error) {
	if len(pokes) < 5 || len(pokes) > 7 {
		return nil, fmt.Errorf("%w: 需要 5-7 张牌，实际 %d 张", ErrUnknownPoke, len(pokes))
	}
	if err := CheckPokes(pokes); err != nil {
		return nil, err
	}
	hand := []byte{pokes[0], pokes[1]}
	pub := append([]byte{}, pokes[2:]...)
	return GetMaxHandPub(hand, pub)
}

// GetMaxHandPub returns the best 5-card hand of 2 hole cards and 3-5
// community cards. It maps the stored best-hand key back onto the actual cards
// that were dealt.
func GetMaxHandPub(hand, pub []byte) ([]byte, error) {
	if len(hand) != 2 {
		return nil, fmt.Errorf("%w: 手牌需要 2 张，实际 %d 张", ErrUnknownPoke, len(hand))
	}
	if len(pub) < 3 || len(pub) > 5 {
		return nil, fmt.Errorf("%w: 公牌需要 3-5 张，实际 %d 张", ErrUnknownPoke, len(pub))
	}
	if err := CheckPokes(hand); err != nil {
		return nil, err
	}
	if err := CheckPokes(pub); err != nil {
		return nil, err
	}
	ret := maxHandPub(hand, pub)
	if len(ret) == 0 {
		return nil, fmt.Errorf("%w: 未找到这手牌的牌力数据", ErrUnknownPoke)
	}
	return ret, nil
}

// maxHandPub is the unvalidated implementation behind GetMaxHandPub.
func maxHandPub(hand, pub []byte) []byte {
	var ret []byte
	tmp := append(append([]byte{}, hand...), pub...)
	keyData := LookupKeyData(GenCardBindBytes(tmp))
	if keyData == nil {
		return ret
	}

	max := KeyToPoke(keyData.Max)

	pubtmp := append([]byte{}, pub...)
	handtmp := append([]byte{}, hand...)

	if keyData.Type == TexasCardTypeTongHua ||
		keyData.Type == TexasCardTypeTongHuaShun ||
		keyData.Type == TexasCardTypeKingTongHuaShun {
		var srccolor [4]int
		for _, c := range tmp {
			srccolor[c>>4]++
		}

		srcmaxColor := 0
		srcmaxColorNum := 0
		for i := 0; i < len(srccolor); i++ {
			if srccolor[i] >= srcmaxColorNum {
				srcmaxColor = i
				srcmaxColorNum = srccolor[i]
			}
		}

		for i := 0; i < len(max); i++ {
			for j := 0; j < len(pubtmp); j++ {
				if int(pubtmp[j]%16) == int(max[i]%16) && int(pubtmp[j]>>4) == srcmaxColor &&
					max[i] != 0 && pubtmp[j] != 0 {
					ret = append(ret, pubtmp[j])

					max[i] = 0
					pubtmp[j] = 0
					break
				}
			}
		}

		if len(ret) < 5 {
			for i := 0; i < len(max); i++ {
				for j := 0; j < len(handtmp); j++ {
					if int(handtmp[j]%16) == int(max[i]%16) && int(handtmp[j]>>4) == srcmaxColor &&
						max[i] != 0 && handtmp[j] != 0 {
						ret = append(ret, handtmp[j])

						max[i] = 0
						handtmp[j] = 0
						break
					}
				}
			}
		}

		for i := 0; i < len(max); i++ {
			if max[i] != 0 {
				max[i] = byte(srcmaxColor)<<4 | (max[i] % 16)
			}
		}
	} else {
		for j := 0; j < len(pubtmp); j++ {
			for i := 0; i < len(max); i++ {
				if pubtmp[j] == max[i] && max[i] != 0 && pubtmp[j] != 0 {
					ret = append(ret, pubtmp[j])

					max[i] = 0
					pubtmp[j] = 0
					break
				}
			}
		}

		for j := 0; j < len(handtmp); j++ {
			for i := 0; i < len(max); i++ {
				if handtmp[j] == max[i] && max[i] != 0 && handtmp[j] != 0 {
					ret = append(ret, handtmp[j])

					max[i] = 0
					handtmp[j] = 0
					break
				}
			}
		}

		for i := 0; i < len(max); i++ {
			for j := 0; j < len(pubtmp); j++ {
				if int(pubtmp[j]%16) == int(max[i]%16) && max[i] != 0 && pubtmp[j] != 0 {
					ret = append(ret, pubtmp[j])

					max[i] = 0
					pubtmp[j] = 0
					break
				}
			}
		}

		if len(ret) < 5 {
			for i := 0; i < len(max); i++ {
				for j := 0; j < len(handtmp); j++ {
					if int(handtmp[j]%16) == int(max[i]%16) && max[i] != 0 && handtmp[j] != 0 {
						ret = append(ret, handtmp[j])

						max[i] = 0
						handtmp[j] = 0
						break
					}
				}
			}
		}
	}

	sort.Slice(ret, func(a, b int) bool { return ret[a] < ret[b] })

	return ret
}

// GetWinPosition returns the absolute rank of the given cards among all
// combinations of the same size.
func GetWinPosition(pokes []byte) (int, error) {
	keyData, err := GetKeyData(pokes)
	if err != nil {
		return 0, err
	}
	return keyData.Postion, nil
}

// GetWinPositionStr is GetWinPosition for a comma-separated card string.
func GetWinPositionStr(str string) (int, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return 0, err
	}
	return GetWinPosition(pokes)
}

// GetWinProbability returns the rank ratio of the given cards among all
// combinations of the same size.
func GetWinProbability(pokes []byte) (float64, error) {
	keyData, err := GetKeyData(pokes)
	if err != nil {
		return 0, err
	}
	total := int64(1)
	for i := 0; i < len(pokes); i++ {
		total = total * (GenNum - int64(i))
	}
	for i := len(pokes); i >= 1; i-- {
		total = total / int64(i)
	}
	return float64(keyData.Index) / float64(total), nil
}

// GetWinProbabilityStr is GetWinProbability for a comma-separated string.
func GetWinProbabilityStr(str string) (float64, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return 0, err
	}
	return GetWinProbability(pokes)
}

// GetWinMax returns the encoded key of the best 5-card hand.
func GetWinMax(pokes []byte) (int64, error) {
	keyData, err := GetKeyData(pokes)
	if err != nil {
		return 0, err
	}
	return keyData.Max, nil
}

// GetWinMaxStr is GetWinMax for a comma-separated card string.
func GetWinMaxStr(str string) (int64, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return 0, err
	}
	return GetWinMax(pokes)
}

// GetWinType returns the card type constant of the given cards.
func GetWinType(pokes []byte) (int, error) {
	keyData, err := GetKeyData(pokes)
	if err != nil {
		return 0, err
	}
	return keyData.Type, nil
}

// GetWinTypeStr is GetWinType for a comma-separated card string.
func GetWinTypeStr(str string) (int, error) {
	pokes, err := StrToPokes(str)
	if err != nil {
		return 0, err
	}
	return GetWinType(pokes)
}

// CompareKey compares two hand keys by rank. Returns a positive value when k1
// is stronger, 0 when equal, negative when weaker. Keys holding cards outside
// the 52-card deck are rejected.
func CompareKey(k1, k2 int64) (int, error) {
	if err := CheckKey(k1); err != nil {
		return 0, err
	}
	if err := CheckKey(k2); err != nil {
		return 0, err
	}
	return compareKey(k1, k2), nil
}

// compareKey is the unvalidated implementation behind CompareKey; it is the
// hot path used while sorting generated keys.
func compareKey(k1, k2 int64) int {
	keyData1 := LookupKeyData(k1)
	keyData2 := LookupKeyData(k2)
	if keyData1 == nil && keyData2 == nil {
		return 0
	}
	if keyData1 == nil {
		return -1
	}
	if keyData2 == nil {
		return 1
	}
	return keyData1.Postion - keyData2.Postion
}

// CompareBytes compares two hands given as card bytes.
func CompareBytes(bytes1, bytes2 []byte) (int, error) {
	if err := CheckPokes(bytes1); err != nil {
		return 0, err
	}
	if err := CheckPokes(bytes2); err != nil {
		return 0, err
	}
	return compareKey(GenCardBindBytes(bytes1), GenCardBindBytes(bytes2)), nil
}

// CompareStr compares two hands given as comma-separated card strings.
func CompareStr(str1, str2 string) (int, error) {
	pokes1, err := StrToPokes(str1)
	if err != nil {
		return 0, err
	}
	pokes2, err := StrToPokes(str2)
	if err != nil {
		return 0, err
	}
	return CompareBytes(pokes1, pokes2)
}

// ErrNoProbabilityData is returned when the win-rate tables that
// GetHandProbability needs are not loaded or do not contain the hand.
var ErrNoProbabilityData = errors.New("缺少胜率表数据")

// GetHandProbabilityKeyPub is GetHandProbability with hand and public cards
// given as encoded keys.
func GetHandProbabilityKeyPub(hand, pub int64) (float32, error) {
	if err := CheckKey(hand); err != nil {
		return 0, err
	}
	if err := CheckKey(pub); err != nil {
		return 0, err
	}
	return GetHandProbability(KeyToPoke(hand), KeyToPoke(pub))
}

// GetHandProbability returns the exact 1v1 win probability of 2 hole cards
// plus 0-5 community cards against one random opponent hand. Ties count as
// half a win: P = (W + 0.5*T) / (W + T + L).
//
// The answer comes from the first source available for the level:
//
//  1. the in-memory exact table, when LoadWinRateRange loaded that level;
//  2. the packed sqlite database, when LoadSqlite opened one covering it;
//  3. on-the-fly enumeration, comparing the Index of the strength tables
//     instead of re-evaluating hands.
//
// Levels 0-3 are loaded into memory by default (LoadProbilityDir); levels 4 and
// 5 are not, because decoded they need ~223 MiB and ~1.97 GiB of resident
// memory, while the enumeration answers them in ~2 ms and ~0.2 ms.
func GetHandProbability(hand, pub []byte) (float32, error) {
	if err := CheckPokes(hand); err != nil {
		return 0, err
	}
	if err := CheckPokes(pub); err != nil {
		return 0, err
	}
	if len(hand) != 2 || len(pub) > 5 {
		return 0, ErrNoProbabilityData
	}

	win, tie, total, ok := exactCounts(hand, pub)
	if !ok || total == 0 {
		return 0, ErrNoProbabilityData
	}
	return float32((2*float64(win) + float64(tie)) / (2 * float64(total))), nil
}

// exactCounts answers (win, tie, total) for a situation from the first
// available source: the in-memory table, the sqlite database, enumeration.
func exactCounts(hand, pub []byte) (win, tie, total uint64, ok bool) {
	n := len(pub)
	if t := WinRateMap[n]; t.Len() != 0 {
		cards := append(append([]byte{}, hand...), pub...)
		w, ti, tot, found := t.Get(CanonKey(cards, len(hand)))
		if !found {
			return 0, 0, 0, false
		}
		return uint64(w), uint64(ti), uint64(tot), true
	}
	if w, ti, found := sqliteCounts(n, hand, pub); found {
		return uint64(w), uint64(ti), uint64(winRateTotal(n)), true
	}
	w, ti, l := EquityExact(hand, pub)
	if w+ti+l == 0 {
		return 0, 0, 0, false
	}
	return uint64(w), uint64(ti), uint64(w + ti + l), true
}

// winRateTotal returns the constant enumeration total of level n. It is the
// orbit size times the 990 opponent pairs: after 2 hole cards and n community
// cards, the remaining 5-n board cards can be completed in C(50-n, 5-n) ways.
// The generator uses the same formula, so the table header is only a cross
// check; the sqlite rows do not have to store it at all.
func winRateTotal(n int) uint32 {
	if n < 0 || n > 5 {
		return 0
	}
	c := uint64(1)
	for i := 0; i < 5-n; i++ {
		c = c * uint64(50-n-i) / uint64(i+1)
	}
	return uint32(c * 990)
}

// deck52 returns the 52 card bytes in ascending order.
func deck52() [52]byte {
	var d [52]byte
	n := 0
	for c := byte(0); c < 4; c++ {
		for v := byte(PokeValue2); v <= PokeValueA; v++ {
			d[n] = c<<4 | v
			n++
		}
	}
	return d
}

// CanonKey returns the suit-canonical key of a situation made of the first
// nHole cards plus the remaining public cards. Each suit is labelled 0..3 by
// sorting the suits on its signature (which ranks it holds, and which of those
// are hole cards), which is invariant under the 24-element suit permutation
// group; suits sharing a signature hold the same (rank, role) pairs, so any
// tie-break yields the same labelled cards. Hole cards and public cards occupy
// separate positions of the key because the two roles are not interchangeable.
// It is the key of the exact win-rate tables and must stay in sync with the
// generator in cmd/genrate.
func CanonKey(cards []byte, nHole int) int64 {
	var mask, hole [4]int
	for i, c := range cards {
		mask[c>>4] |= 1 << (c & 0x0f)
		if i < nHole {
			hole[c>>4] |= 1 << (c & 0x0f)
		}
	}
	var label [4]int8
	used := [4]bool{}
	for k := int8(0); k < 4; k++ {
		best, bestSig := -1, 0
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
	for i, c := range cards {
		b := byte(label[c>>4])<<4 | (c & 0x0f)
		if i < nHole {
			holeCards[nh] = b
			nh++
		} else {
			pubCards[np] = b
			np++
		}
	}
	sortCardsAsc(holeCards[:nh])
	sortCardsAsc(pubCards[:np])

	var k int64
	for i := 0; i < nh; i++ {
		k = k*100 + int64(holeCards[i])
	}
	for i := 0; i < np; i++ {
		k = k*100 + int64(pubCards[i])
	}
	return k
}

// key7Board merges two ascending cards x<y into the ascending 5-card board and
// returns the 7-card key.
func key7Board(b *[5]byte, x, y byte) int64 {
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
		k = k*100 + int64(buf[n])
	}
	return k
}

// sortCardsAsc sorts a small card slice ascending.
func sortCardsAsc(cs []byte) {
	for i := 1; i < len(cs); i++ {
		for j := i; j > 0 && cs[j] < cs[j-1]; j-- {
			cs[j], cs[j-1] = cs[j-1], cs[j]
		}
	}
}

// EquityExact enumerates every completion of the community cards and every
// opponent hole-card pair, and counts wins, ties and losses of the given 2
// hole cards. Ties count as half a win in GetHandProbability. Opponent cards
// never include a hero card, and every (board, opponent) pair is counted
// exactly once.
func EquityExact(hand, pub []byte) (win, tie, loss int64) {
	var known [7]byte
	copy(known[:], hand)
	copy(known[len(hand):], pub)
	used := make(map[byte]bool, len(known))
	for _, c := range known {
		used[c] = true
	}
	var rest []byte
	d := deck52()
	for _, c := range d {
		if !used[c] {
			rest = append(rest, c)
		}
	}

	need := 5 - len(pub)
	var board [5]byte
	copy(board[:], pub)
	var boards [][5]byte
	var gen func(start, depth int)
	gen = func(start, depth int) {
		if depth == need {
			b := board
			sortCardsAsc(b[:])
			boards = append(boards, b)
			return
		}
		for i := start; i <= len(rest)-(need-depth); i++ {
			board[len(pub)+depth] = rest[i]
			gen(i+1, depth+1)
		}
	}
	gen(0, 0)

	heroCards := []byte{hand[0], hand[1]}
	sortCardsAsc(heroCards)

	var wg sync.WaitGroup
	workers := 1
	// 20k evaluations is roughly the point where the split pays for the
	// goroutine hand-off: N>=4 (45,540 pairs) benefits, N=5 (990) does not.
	if len(boards)*990 >= 20000 {
		workers = runtime.NumCPU()
	}
	results := make([]struct{ w, t, l int64 }, workers)
	chunk := (len(boards) + workers - 1) / workers
	for w := 0; w < workers; w++ {
		from := w * chunk
		to := from + chunk
		if to > len(boards) {
			to = len(boards)
		}
		if from >= to {
			continue
		}
		wg.Add(1)
		go func(slot, from, to int) {
			defer wg.Done()
			var w, t, l int64
			for _, b := range boards[from:to] {
				var opp [45]byte
				n := 0
				for _, c := range rest {
					if c == b[0] || c == b[1] || c == b[2] || c == b[3] || c == b[4] {
						continue
					}
					opp[n] = c
					n++
				}
				hero := LookupKeyData(key7Board(&b, heroCards[0], heroCards[1]))
				if hero == nil {
					continue
				}
				for i := 0; i < n; i++ {
					for j := i + 1; j < n; j++ {
						data := LookupKeyData(key7Board(&b, opp[i], opp[j]))
						if data == nil {
							continue
						}
						switch {
						case hero.Index > data.Index:
							w++
						case hero.Index == data.Index:
							t++
						default:
							l++
						}
					}
				}
			}
			results[slot].w, results[slot].t, results[slot].l = w, t, l
		}(w, from, to)
	}
	wg.Wait()
	for _, r := range results {
		win += r.w
		tie += r.t
		loss += r.l
	}
	return win, tie, loss
}

// GetHandProbabilityStr is GetHandProbability for comma-separated strings.
func GetHandProbabilityStr(hand, pub string) (float32, error) {
	handPokes, err := StrToPokes(hand)
	if err != nil {
		return 0, err
	}
	pubPokes, err := StrToPokes(pub)
	if err != nil {
		return 0, err
	}
	return GetHandProbability(handPokes, pubPokes)
}
