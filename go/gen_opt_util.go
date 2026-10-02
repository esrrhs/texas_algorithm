package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Package-level state used while generating the optimized tables, mirroring
// the static fields of Java's GenOptUtil.
var (
	OptTotalKey   int64
	OptLastPrint  int
	OptBeginPrint int64
	optKeys       = map[int64]bool{}
)

// OptColorData writes texas_data_color.txt: for every flush-type row of
// texas_data.txt, the suit-normalized key, deduplicated.
func OptColorData() {
	in, err := os.Open("texas_data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_color.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	OptTotalKey = 0
	OptLastPrint = 0
	OptBeginPrint = time.Now().UnixMilli()
	for k := range optKeys {
		delete(optKeys, k)
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, _ := strconv.ParseInt(params[0], 10, 64)
		i, _ := strconv.ParseInt(params[1], 10, 64)
		index, _ := strconv.ParseInt(params[2], 10, 64)
		total, _ := strconv.ParseInt(params[3], 10, 64)
		maxType, _ := strconv.ParseInt(params[7], 10, 64)

		if maxType == TexasCardTypeTongHua || maxType == TexasCardTypeTongHuaShun || maxType == TexasCardTypeKingTongHuaShun {
			colorKey := ChangeColor(key)

			if !optKeys[colorKey] {
				str := fmt.Sprint(colorKey) + " " + fmt.Sprint(i) + " " + fmt.Sprint(index) + " " + fmt.Sprint(total) + " " +
					ToString(colorKey) + " " + fmt.Sprint(Max(colorKey)) + " " + ToString(Max(colorKey)) + " " +
					fmt.Sprint(maxType) + "\n"
				w.WriteString(str)
				optKeys[colorKey] = true
			}
		}

		OptTotalKey++

		cur := int(OptTotalKey * 100 / Total)
		if cur != OptLastPrint {
			OptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-OptBeginPrint) / float64(OptTotalKey)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(Total-OptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-OptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(OptTotalKey)/(float64(now-OptBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()
	for k := range optKeys {
		delete(optKeys, k)
	}

	fmt.Println("optData finish", OptTotalKey)
}

// OptNormalData writes texas_data_normal.txt: for every non-flush-type row of
// texas_data.txt, the suit-stripped key, deduplicated.
func OptNormalData() {
	in, err := os.Open("texas_data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_normal.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	OptTotalKey = 0
	OptLastPrint = 0
	OptBeginPrint = time.Now().UnixMilli()
	for k := range optKeys {
		delete(optKeys, k)
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, _ := strconv.ParseInt(params[0], 10, 64)
		i, _ := strconv.ParseInt(params[1], 10, 64)
		index, _ := strconv.ParseInt(params[2], 10, 64)
		total, _ := strconv.ParseInt(params[3], 10, 64)
		keystr := params[4]
		max, _ := strconv.ParseInt(params[5], 10, 64)
		maxstr := params[6]
		maxType, _ := strconv.ParseInt(params[7], 10, 64)
		removeKey := RemoveColor(key)
		if !(maxType == TexasCardTypeTongHua || maxType == TexasCardTypeTongHuaShun || maxType == TexasCardTypeKingTongHuaShun) {
			if !optKeys[removeKey] {
				str := fmt.Sprint(removeKey) + " " + fmt.Sprint(i) + " " + fmt.Sprint(index) + " " + fmt.Sprint(total) + " " +
					keystr + " " + fmt.Sprint(max) + " " + maxstr + " " + fmt.Sprint(maxType) + "\n"
				w.WriteString(str)
				optKeys[removeKey] = true
			}
		}

		OptTotalKey++

		cur := int(OptTotalKey * 100 / Total)
		if cur != OptLastPrint {
			OptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-OptBeginPrint) / float64(OptTotalKey)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(Total-OptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-OptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(OptTotalKey)/(float64(now-OptBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()
	for k := range optKeys {
		delete(optKeys, k)
	}

	fmt.Println("optData finish", OptTotalKey)
}

// RemoveColor collapses all suit colors of the non-wild cards to 方 (diamonds),
// sorts and re-binds the key, mirroring Java's GenOptUtil.removeColor.
func RemoveColor(k int64) int64 {
	cs := keyToCards(k)

	for i, c := range cs {
		if !IsGuiByte(c) {
			cs[i] = int(PokeColorFang)<<4 | (c % 16)
		}
	}

	sort.Ints(cs)

	return GenCardBindInts(cs)
}

// ChangeColor normalizes the dominant suit of the key to 黑 (spades) and every
// other suit to 方 (diamonds), keeping wild cards, then sorts and re-binds,
// mirroring Java's GenOptUtil.changeColor.
func ChangeColor(k int64) int64 {
	cs := keyToCards(k)

	var color [4]int
	for _, i := range cs {
		if !IsGuiByte(i) {
			color[byte(i)>>4]++
		}
	}

	maxColor := 0
	maxColorNum := 0
	for i := 0; i < len(color); i++ {
		if color[i] > maxColorNum {
			maxColor = i
			maxColorNum = color[i]
		}
	}

	for i, c := range cs {
		if !IsGuiByte(c) {
			if c>>4 == maxColor {
				cs[i] = int(PokeColorHei)<<4 | (c % 16)
			} else {
				cs[i] = int(PokeColorFang)<<4 | (c % 16)
			}
		}
	}

	sort.Ints(cs)

	return GenCardBindInts(cs)
}

// keyToCards decodes a key into its card bytes, most significant first.
func keyToCards(k int64) []int {
	var cs []int
	if k > 1000000000000 {
		cs = append(cs, int(k%100000000000000/1000000000000))
	}
	if k > 10000000000 {
		cs = append(cs, int(k%1000000000000/10000000000))
	}
	if k > 100000000 {
		cs = append(cs, int(k%10000000000/100000000))
	}
	if k > 1000000 {
		cs = append(cs, int(k%100000000/1000000))
	}
	if k > 10000 {
		cs = append(cs, int(k%1000000/10000))
	}
	if k > 100 {
		cs = append(cs, int(k%10000/100))
	}
	if k > 1 {
		cs = append(cs, int(k%100/1))
	}
	return cs
}
