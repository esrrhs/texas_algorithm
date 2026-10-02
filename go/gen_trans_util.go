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

// Package-level state for the N-card win-rate table generation, mirroring the
// static fields of Java's GenTransUtil. Set GenTransN before generating.
var (
	GenTransN                = 6
	genTransTotal      int64 = 1
	genTransTotalKey   int64
	genTransLastPrint  int
	genTransBeginPrint int64
	genTransKeys       = map[int64]*GenTransKeyData{}
)

// GenTransKeyData accumulates the win statistics of one N-card key.
type GenTransKeyData struct {
	Win int64
	Num int64
	Min float64
	Max float64
}

func newGenTransKeyData() *GenTransKeyData {
	return &GenTransKeyData{Min: 1, Max: 0}
}

// GenTransGenKey enumerates all C(54, GenTransN) combinations as the key set.
func GenTransGenKey() {
	genTransTotal = 1
	for i := 0; i < GenTransN; i++ {
		genTransTotal *= GenNum - int64(i)
	}
	for i := GenTransN; i >= 1; i-- {
		genTransTotal /= int64(i)
	}
	genTransBeginPrint = time.Now().UnixMilli()
	for k := range genTransKeys {
		delete(genTransKeys, k)
	}
	genTransTotalKey = 0
	genTransLastPrint = 0

	genTransCard()

	fmt.Println("genKey finish", genTransTotal)
}

func genTransCard() {
	list := genAllCards()

	tmp := make([]int, GenTransN)
	Permutation(list, 0, 0, GenTransN, tmp, func(tmp []int) {
		genTransCardSave(tmp)
	})
}

func genTransCardSave(tmp []int) {
	c := GenCardBindInts(tmp)

	genTransKeys[c] = newGenTransKeyData()
	genTransTotalKey++

	cur := int(genTransTotalKey * 100 / genTransTotal)
	if cur != genTransLastPrint {
		genTransLastPrint = cur

		now := time.Now().UnixMilli()
		per := float64(now-genTransBeginPrint) / float64(genTransTotalKey)
		fmt.Println("N" + strconv.Itoa(GenTransN) + " " + fmt.Sprint(cur) + "% 需要" +
			fmt.Sprint(per*float64(genTransTotal-genTransTotalKey)/60/1000) + "分" +
			" 用时" + fmt.Sprint((now-genTransBeginPrint)/60/1000) + "分" + " 速度" +
			fmt.Sprint(float64(genTransTotalKey)/(float64(now-genTransBeginPrint)/1000)) + "条/秒")
	}
}

// GenTransTransData reads texas_data.txt and, for every N-card subset of each
// 7-card key, accumulates the average win rate into texas_data_N.txt.
func GenTransTransData() {
	in, err := os.Open("texas_data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_" + strconv.Itoa(GenTransN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	genTransTotalKey = 0
	genTransLastPrint = 0
	genTransBeginPrint = time.Now().UnixMilli()

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		key, _ := strconv.ParseInt(fields[0], 10, 64)
		win, _ := strconv.ParseInt(fields[2], 10, 64)
		tmp := GenTransGetKeyList(key)
		for _, l := range tmp {
			keyData := genTransKeys[l]
			if keyData == nil {
				continue
			}
			keyData.Win += win
			keyData.Num++
			p := float64(win) / float64(Total)
			if p < keyData.Min {
				keyData.Min = p
			}
			if p > keyData.Max {
				keyData.Max = p
			}
		}

		genTransTotalKey++

		cur := int(genTransTotalKey * 100 / Total)
		if cur != genTransLastPrint {
			genTransLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genTransBeginPrint) / float64(genTransTotalKey)
			fmt.Println("step1 N" + strconv.Itoa(GenTransN) + " " + fmt.Sprint(cur) + "% 需要" +
				fmt.Sprint(per*float64(Total-genTransTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genTransBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genTransTotalKey)/(float64(now-genTransBeginPrint)/1000)) + "条/秒")
		}
	}

	genTransTotalKey = 0
	genTransLastPrint = 0
	genTransBeginPrint = time.Now().UnixMilli()

	for key, e := range genTransKeys {
		win := float64(e.Win) / float64(e.Num) / float64(Total)

		tmp := fmt.Sprint(key) + " " + strconv.FormatFloat(win, 'g', -1, 64) + " " +
			strconv.FormatFloat(e.Min, 'g', -1, 64) + " " + strconv.FormatFloat(e.Max, 'g', -1, 64) + " " +
			ToString(key) + "\n"
		w.WriteString(tmp)
		genTransTotalKey++

		cur := int(genTransTotalKey * 100 / Total)
		if cur != genTransLastPrint {
			genTransLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genTransBeginPrint) / float64(genTransTotalKey)
			fmt.Println("step2 N" + strconv.Itoa(GenTransN) + " " + fmt.Sprint(cur) + "% 需要" +
				fmt.Sprint(per*float64(Total-genTransTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genTransBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genTransTotalKey)/(float64(now-genTransBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()
	for k := range genTransKeys {
		delete(genTransKeys, k)
	}

	fmt.Println("transData finish", genTransTotalKey)
}

// GenTransGetKeyList returns the distinct N-card subset keys contained in the
// given 7-card key, mirroring Java's GenTransUtil.getKeyList.
func GenTransGetKeyList(key int64) []int64 {
	var ret []int64
	var list []int
	for key > 100 {
		list = append(list, int(key%100))
		key /= 100
	}
	list = append(list, int(key))
	sort.Ints(list)

	tmp := make([]int, GenTransN)
	Permutation(list, 0, 0, GenTransN, tmp, func(tmp []int) {
		c := GenCardBindInts(tmp)
		found := false
		for _, r := range ret {
			if r == c {
				found = true
				break
			}
		}
		if !found {
			ret = append(ret, c)
		}
	})

	return ret
}
