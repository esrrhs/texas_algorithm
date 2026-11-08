package texas_algorithm

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
	// Only n <= 2 is precomputed; n >= 3 is enumerated on the fly.
	WinRateMap [3]map[int64]*WinRateData
)

// WinRateData is the exact win/tie/total count of one canonical situation.
type WinRateData struct {
	Win   int64
	Tie   int64
	Total int64
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

// IsProbabilityLoaded reports whether all three exact win-rate tables
// (texas_data_winrate_0..2.txt) are in memory. Only the N <= 2 lookups depend
// on them; N >= 3 is enumerated on the fly and works without any table.
func IsProbabilityLoaded() bool {
	for _, m := range WinRateMap {
		if len(m) == 0 {
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

// LoadProbilityDir loads the exact win-rate tables from the given directory.
// The three tables are independent: a missing texas_data_winrate_N.txt only
// makes the matching N-card lookup return ErrNoProbabilityData, everything
// else keeps working.
func LoadProbilityDir(dir string) {
	begin := time.Now().UnixMilli()

	loaded := 0
	for i := 0; i <= 2; i++ {
		err := LoadWinRateFile(i, filepath.Join(dir, "texas_data_winrate_"+strconv.Itoa(i)+".txt"))
		if err != nil {
			fmt.Println(err)
			continue
		}
		loaded++
	}

	fmt.Println("load time", time.Now().UnixMilli()-begin, "tables", loaded)
}

// LoadWinRateFile loads one exact win-rate table: "<key> <win> <tie> <total>".
func LoadWinRateFile(n int, path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	m := make(map[int64]*WinRateData, 1<<10)
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Fields(scanner.Text())
		if len(params) < 4 {
			continue
		}
		key, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			return err
		}
		w, err := strconv.ParseInt(params[1], 10, 64)
		if err != nil {
			return err
		}
		t, err := strconv.ParseInt(params[2], 10, 64)
		if err != nil {
			return err
		}
		total, err := strconv.ParseInt(params[3], 10, 64)
		if err != nil {
			return err
		}
		m[key] = &WinRateData{Win: w, Tie: t, Total: total}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	WinRateMap[n] = m
	return nil
}

// IsWinRateLoaded reports whether the exact win-rate tables are loaded.
func IsWinRateLoaded() bool {
	return len(WinRateMap[0]) != 0
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
// n community cards below 3 are answered from the precomputed exact tables
// (texas_data_winrate_0..2.txt, keyed by the suit-canonical card key). Three
// or more community cards are enumerated on the fly, comparing the Index of
// the new strength tables instead of re-evaluating hands.
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
	if len(pub) >= 3 {
		win, tie, loss := EquityExact(hand, pub)
		total := win + tie + loss
		if total == 0 {
			return 0, ErrNoProbabilityData
		}
		return float32((2*float64(win) + float64(tie)) / (2 * float64(total))), nil
	}
	m := WinRateMap[len(pub)]
	if len(m) == 0 {
		return 0, ErrNoProbabilityData
	}
	cards := append(append([]byte{}, hand...), pub...)
	data := m[CanonKey(cards, len(hand))]
	if data == nil || data.Total == 0 {
		return 0, ErrNoProbabilityData
	}
	return float32((2*float64(data.Win) + float64(data.Tie)) / (2 * float64(data.Total))), nil
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
