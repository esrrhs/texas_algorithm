package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Package-level state for the N-card table optimization, mirroring the static
// fields of Java's GenExtraOptUtil. Set GenExtraOptN before optimizing.
var (
	GenExtraOptN          = 6
	genExtraOptTotal      int64
	genExtraOptTotalKey   int64
	genExtraOptLastPrint  int
	genExtraOptBeginPrint int64
	genExtraOptKeys       = map[int64]bool{}
)

// GenExtraOptColorData writes texas_data_extra_color_N.txt from
// texas_data_extra_N.txt, deduplicating the suit-normalized flush keys.
func GenExtraOptColorData() {
	genExtraOptTotal = 1
	for i := 0; i < GenExtraOptN; i++ {
		genExtraOptTotal *= GenNum - int64(i)
	}
	for i := GenExtraOptN; i >= 1; i-- {
		genExtraOptTotal /= int64(i)
	}

	in, err := os.Open("texas_data_extra_" + strconv.Itoa(GenExtraOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_extra_color_" + strconv.Itoa(GenExtraOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	genExtraOptTotalKey = 0
	genExtraOptLastPrint = 0
	genExtraOptBeginPrint = time.Now().UnixMilli()
	for k := range genExtraOptKeys {
		delete(genExtraOptKeys, k)
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

			if !genExtraOptKeys[colorKey] {
				str := fmt.Sprint(colorKey) + " " + fmt.Sprint(i) + " " + fmt.Sprint(index) + " " + fmt.Sprint(total) + " " +
					ToString(colorKey) + " " + fmt.Sprint(Max(colorKey)) + " " + ToString(Max(colorKey)) + " " +
					fmt.Sprint(maxType) + "\n"
				w.WriteString(str)
				genExtraOptKeys[colorKey] = true
			}
		}

		genExtraOptTotalKey++

		cur := int(genExtraOptTotalKey * 100 / genExtraOptTotal)
		if cur != genExtraOptLastPrint {
			genExtraOptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genExtraOptBeginPrint) / float64(genExtraOptTotalKey)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(genExtraOptTotal-genExtraOptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genExtraOptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genExtraOptTotalKey)/(float64(now-genExtraOptBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()
	for k := range genExtraOptKeys {
		delete(genExtraOptKeys, k)
	}

	fmt.Println("optData finish", genExtraOptTotalKey)
}

// GenExtraOptNormalData writes texas_data_extra_normal_N.txt from
// texas_data_extra_N.txt, deduplicating the suit-stripped non-flush keys.
func GenExtraOptNormalData() {
	genExtraOptTotal = 1
	for i := 0; i < GenExtraOptN; i++ {
		genExtraOptTotal *= GenNum - int64(i)
	}
	for i := GenExtraOptN; i >= 1; i-- {
		genExtraOptTotal /= int64(i)
	}

	in, err := os.Open("texas_data_extra_" + strconv.Itoa(GenExtraOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_extra_normal_" + strconv.Itoa(GenExtraOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	genExtraOptTotalKey = 0
	genExtraOptLastPrint = 0
	genExtraOptBeginPrint = time.Now().UnixMilli()
	for k := range genExtraOptKeys {
		delete(genExtraOptKeys, k)
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
			if !genExtraOptKeys[removeKey] {
				str := fmt.Sprint(removeKey) + " " + fmt.Sprint(i) + " " + fmt.Sprint(index) + " " + fmt.Sprint(total) + " " +
					keystr + " " + fmt.Sprint(max) + " " + maxstr + " " + fmt.Sprint(maxType) + "\n"
				w.WriteString(str)
				genExtraOptKeys[removeKey] = true
			}
		}

		genExtraOptTotalKey++

		cur := int(genExtraOptTotalKey * 100 / genExtraOptTotal)
		if cur != genExtraOptLastPrint {
			genExtraOptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genExtraOptBeginPrint) / float64(genExtraOptTotalKey)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(genExtraOptTotal-genExtraOptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genExtraOptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genExtraOptTotalKey)/(float64(now-genExtraOptBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()
	for k := range genExtraOptKeys {
		delete(genExtraOptKeys, k)
	}

	fmt.Println("optData finish", genExtraOptTotalKey)
}
