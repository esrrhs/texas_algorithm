package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

// ProbilityData holds win-rate statistics of one hand key.
type ProbilityData struct {
	Avg float32
	Min float32
	Max float32
}

var (
	ColorMap  = map[int64]*KeyData{}
	NormalMap = map[int64]*KeyData{}

	ProbilityMap    [7]map[int64]*ProbilityData
	OptProbilityMap [7]map[int64]*ProbilityData
)

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

// IsProbabilityLoaded reports whether the probability tables are loaded.
func IsProbabilityLoaded() bool {
	return ProbilityMap[2] != nil && len(ProbilityMap[2]) != 0
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

// LoadProbilityDir loads the probability tables from the given directory.
func LoadProbilityDir(dir string) {
	begin := time.Now().UnixMilli()

	for i := 6; i >= 2; i-- {
		err := LoadProbilityFile(i, filepath.Join(dir, "texas_data_opt_"+strconv.Itoa(i)+".txt"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("load time", time.Now().UnixMilli()-begin)
}

// LoadProbilityFile loads one probability table file into ProbilityMap[n] /
// OptProbilityMap[n]. Rows with type 0 go to the plain table, others to the
// optimized table.
func LoadProbilityFile(n int, path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	return loadProbilityFrom(n, in)
}

func loadProbilityFrom(n int, in *os.File) error {
	ProbilityMap[n] = map[int64]*ProbilityData{}
	OptProbilityMap[n] = map[int64]*ProbilityData{}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			return err
		}
		typ, err := strconv.ParseInt(params[1], 10, 64)
		if err != nil {
			return err
		}
		probility, err := strconv.ParseFloat(params[2], 32)
		if err != nil {
			return err
		}
		min, err := strconv.ParseFloat(params[3], 32)
		if err != nil {
			return err
		}
		max, err := strconv.ParseFloat(params[4], 32)
		if err != nil {
			return err
		}

		data := &ProbilityData{Avg: float32(probility), Min: float32(min), Max: float32(max)}
		if typ == 0 {
			ProbilityMap[n][key] = data
		} else {
			OptProbilityMap[n][key] = data
		}
	}
	return scanner.Err()
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

// StrToPokeValue converts a rank string ("2".."10", "J", "Q", "K", "A") into
// its byte value.
func StrToPokeValue(str string) byte {
	switch str {
	case "A":
		return PokeValueA
	case "K":
		return PokeValueK
	case "Q":
		return PokeValueQ
	case "J":
		return PokeValueJ
	}
	v, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		panic(err)
	}
	return byte(v)
}

// StrToPoke converts a card string ("方A", "黑10", "鬼"...) into its byte.
func StrToPoke(str string) byte {
	r := []rune(str)
	if len(r) == 0 {
		return 0
	}
	switch r[0] {
	case '方':
		return NewPoke(PokeColorFang, StrToPokeValue(string(r[1:]))).ToByte()
	case '梅':
		return NewPoke(PokeColorMei, StrToPokeValue(string(r[1:]))).ToByte()
	case '红':
		return NewPoke(PokeColorHong, StrToPokeValue(string(r[1:]))).ToByte()
	case '黑':
		return NewPoke(PokeColorHei, StrToPokeValue(string(r[1:]))).ToByte()
	case '鬼':
		return NewPoke(GUI.Color, GUI.Value).ToByte()
	}
	return 0
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

// StrToPokes converts a comma-separated card string into card bytes.
func StrToPokes(str string) []byte {
	var ret []byte
	if len(str) == 0 {
		return ret
	}
	for _, s := range strings.Split(str, ",") {
		ret = append(ret, StrToPoke(s))
	}
	return ret
}

// PokesToStr renders card bytes as their concatenated card strings.
func PokesToStr(pokes []byte) string {
	return KeyToStr(GenCardBindBytes(pokes))
}

// GetKeyDataStr looks up a 7-card comma-separated string; returns nil when the
// string is not exactly 7 cards.
func GetKeyDataStr(str string) *KeyData {
	pokes := StrToPokes(str)
	if len(pokes) != 7 {
		return nil
	}
	return GetKeyData(pokes)
}

// GetKeyData looks up the best-hand data of the given cards.
func GetKeyData(pokes []byte) *KeyData {
	return GetKeyDataKey(GenCardBindBytes(pokes))
}

// GetKeyDataKey looks up the best-hand data of the given 7-card key: the
// suit-normalized flush table and the suit-stripped normal table are both
// consulted, and the entry with the higher rank wins.
func GetKeyDataKey(key int64) *KeyData {
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
// (5-7 cards) together with the wild-card substitution list.
func GetMaxStr(str string) (string, []byte) {
	best, guiTrans := GetMax(StrToPokes(str))
	return PokesToStr(best), guiTrans
}

// GetMaxStrHandPub returns the best 5-card hand of 2 hole cards and 3-5
// community cards, together with the wild-card substitution list.
func GetMaxStrHandPub(hand, pub string) (string, []byte) {
	best, guiTrans := GetMaxHandPub(StrToPokes(hand), StrToPokes(pub))
	return PokesToStr(best), guiTrans
}

// GetMax returns the best 5-card hand of the given cards (2 hole cards
// followed by 3-5 community cards) together with the wild-card substitution
// list.
func GetMax(pokes []byte) ([]byte, []byte) {
	if len(pokes) < 5 || len(pokes) > 7 {
		return []byte{}, nil
	}
	hand := []byte{pokes[0], pokes[1]}
	pub := append([]byte{}, pokes[2:]...)
	return GetMaxHandPub(hand, pub)
}

// GetMaxHandPub returns the best 5-card hand of 2 hole cards and 3-5
// community cards, together with the wild-card substitution list. It maps the
// stored best-hand key back onto the actual cards that were dealt.
func GetMaxHandPub(hand, pub []byte) ([]byte, []byte) {
	var ret []byte
	if len(hand) != 2 {
		return ret, nil
	}
	if len(pub) < 3 || len(pub) > 5 {
		return ret, nil
	}
	tmp := append(append([]byte{}, hand...), pub...)
	keyData := GetKeyData(tmp)
	if keyData == nil {
		return ret, nil
	}

	max := KeyToPoke(keyData.Max)

	pubtmp := append([]byte{}, pub...)
	handtmp := append([]byte{}, hand...)

	if keyData.Type == TexasCardTypeTongHua ||
		keyData.Type == TexasCardTypeTongHuaShun ||
		keyData.Type == TexasCardTypeKingTongHuaShun {
		var srccolor [4]int
		for _, c := range tmp {
			if !IsGuiByte(int(c)) {
				srccolor[c>>4]++
			}
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
				if pubtmp[j] == max[i] && max[i] != 0 && pubtmp[j] != 0 && !IsGuiByte(int(pubtmp[j])) {
					ret = append(ret, pubtmp[j])

					max[i] = 0
					pubtmp[j] = 0
					break
				}
			}
		}

		for j := 0; j < len(handtmp); j++ {
			for i := 0; i < len(max); i++ {
				if handtmp[j] == max[i] && max[i] != 0 && handtmp[j] != 0 && !IsGuiByte(int(handtmp[j])) {
					ret = append(ret, handtmp[j])

					max[i] = 0
					handtmp[j] = 0
					break
				}
			}
		}

		for i := 0; i < len(max); i++ {
			for j := 0; j < len(pubtmp); j++ {
				if int(pubtmp[j]%16) == int(max[i]%16) && max[i] != 0 && pubtmp[j] != 0 && !IsGuiByte(int(pubtmp[j])) {
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
					if int(handtmp[j]%16) == int(max[i]%16) && max[i] != 0 && handtmp[j] != 0 && !IsGuiByte(int(handtmp[j])) {
						ret = append(ret, handtmp[j])

						max[i] = 0
						handtmp[j] = 0
						break
					}
				}
			}
		}
	}

	for len(ret) < 5 {
		ret = append(ret, GUI.ToByte())
	}

	var guiTrans []byte
	for _, m := range max {
		if m != 0 {
			guiTrans = append(guiTrans, m)
		}
	}

	sort.Slice(ret, func(a, b int) bool { return ret[a] < ret[b] })

	return ret, guiTrans
}

// GetWinPosition returns the absolute rank of the given cards among all
// combinations of the same size.
func GetWinPosition(pokes []byte) int {
	keyData := GetKeyData(pokes)
	if keyData == nil {
		return 0
	}
	return keyData.Postion
}

// GetWinPositionStr is GetWinPosition for a comma-separated card string.
func GetWinPositionStr(str string) int {
	return GetWinPosition(StrToPokes(str))
}

// GetWinProbability returns the rank ratio of the given cards among all
// combinations of the same size.
func GetWinProbability(pokes []byte) float64 {
	keyData := GetKeyData(pokes)
	if keyData == nil {
		return 0
	}
	total := int64(1)
	for i := 0; i < len(pokes); i++ {
		total = total * (GenNum - int64(i))
	}
	for i := len(pokes); i >= 1; i-- {
		total = total / int64(i)
	}
	return float64(keyData.Index) / float64(total)
}

// GetWinProbabilityStr is GetWinProbability for a comma-separated string.
func GetWinProbabilityStr(str string) float64 {
	return GetWinProbability(StrToPokes(str))
}

// GetWinMax returns the encoded key of the best 5-card hand.
func GetWinMax(pokes []byte) int64 {
	keyData := GetKeyData(pokes)
	if keyData == nil {
		return 0
	}
	return keyData.Max
}

// GetWinMaxStr is GetWinMax for a comma-separated card string.
func GetWinMaxStr(str string) int64 {
	return GetWinMax(StrToPokes(str))
}

// GetWinType returns the card type constant of the given cards.
func GetWinType(pokes []byte) int {
	keyData := GetKeyData(pokes)
	if keyData == nil {
		return 0
	}
	return keyData.Type
}

// GetWinTypeStr is GetWinType for a comma-separated card string.
func GetWinTypeStr(str string) int {
	return GetWinType(StrToPokes(str))
}

// CompareKey compares two hand keys by rank. Returns a positive value when k1
// is stronger, 0 when equal, negative when weaker.
func CompareKey(k1, k2 int64) int {
	keyData1 := GetKeyDataKey(k1)
	keyData2 := GetKeyDataKey(k2)
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
func CompareBytes(bytes1, bytes2 []byte) int {
	return CompareKey(GenCardBindBytes(bytes1), GenCardBindBytes(bytes2))
}

// CompareStr compares two hands given as comma-separated card strings.
func CompareStr(str1, str2 string) int {
	return CompareBytes(StrToPokes(str1), StrToPokes(str2))
}

// GetHandProbabilityKey returns the stored win-rate statistics of a key with
// 2-6 cards, or nil when out of range or not loaded.
func GetHandProbabilityKey(k int64) *ProbilityData {
	num := 0
	if k > 10000000000 {
		num = 6
	} else if k > 100000000 {
		num = 5
	} else if k > 1000000 {
		num = 4
	} else if k > 10000 {
		num = 3
	} else if k > 100 {
		num = 2
	}
	if num < 2 || num > 6 {
		return nil
	}
	if ProbilityMap[num] == nil || OptProbilityMap[num] == nil {
		return nil
	}

	probilityData := ProbilityMap[num][k]
	if probilityData == nil {
		k = RemoveColor(k)
		probilityData = OptProbilityMap[num][k]
	}
	return probilityData
}

// GetHandProbabilityKeyPub is GetHandProbability with hand and public cards
// given as encoded keys.
func GetHandProbabilityKeyPub(hand, pub int64) float32 {
	return GetHandProbability(KeyToPoke(hand), KeyToPoke(pub))
}

// GetHandProbability estimates the 1v1 win probability of 2 hole cards plus
// 0-4 community cards. Requires the probability tables to be loaded.
func GetHandProbability(hand, pub []byte) float32 {
	allCards := append(append([]byte{}, hand...), pub...)
	sort.Slice(allCards, func(a, b int) bool { return allCards[a] < allCards[b] })
	pubCopy := append([]byte{}, pub...)
	sort.Slice(pubCopy, func(a, b int) bool { return pubCopy[a] < pubCopy[b] })
	pubkey := GenCardBindBytes(pubCopy)

	pubProbilityData := GetHandProbabilityKey(pubkey)

	var avg float32
	if len(allCards) == 7 {
		avg = float32(GetWinProbability(allCards))
	} else {
		totalkey := GenCardBindBytes(allCards)
		totalProbilityData := GetHandProbabilityKey(totalkey)
		if totalProbilityData == nil {
			return 0
		}
		avg = totalProbilityData.Avg
	}

	if pubProbilityData == nil {
		return avg
	}

	p := float32(0.5)

	if avg > pubProbilityData.Avg {
		p += 0.5 * (avg - pubProbilityData.Avg) / (pubProbilityData.Max - pubProbilityData.Avg)
	} else {
		p += 0.5 * (avg - pubProbilityData.Avg) / (pubProbilityData.Avg - pubProbilityData.Min)
	}

	if p > 1 {
		p = 1
	}
	if p < 0 {
		p = 0
	}

	return p
}

// GetHandProbabilityStr is GetHandProbability for comma-separated strings.
func GetHandProbabilityStr(hand, pub string) float32 {
	return GetHandProbability(StrToPokes(hand), StrToPokes(pub))
}
