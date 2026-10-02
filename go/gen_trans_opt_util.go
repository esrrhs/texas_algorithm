package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Package-level state for the N-card win-rate table optimization, mirroring
// the static fields of Java's GenTransOptUtil. Set GenTransOptN first.
var (
	GenTransOptN                = 6
	genTransOptTotal      int64 = 1
	genTransOptTotalKey   int64
	genTransOptLastPrint  int
	genTransOptBeginPrint int64
	genTransOptKeys       = map[int64]string{}
	genTransOptOptKeys    = map[int64]*GenTransOptKeyData{}
)

// GenTransOptKeyData groups the win-rate strings of all keys sharing the same
// suit-stripped key, tracking the most frequent one.
type GenTransOptKeyData struct {
	Max string
	Ps  map[string]int
}

// GenTransOptData reads texas_data_N.txt and writes texas_data_opt_N.txt:
// rows whose win-rate differs from the most frequent one of their
// suit-stripped group are written with flag 0, group winners with flag 1.
func GenTransOptData() {
	genTransOptTotal = 1
	for i := 0; i < GenTransOptN; i++ {
		genTransOptTotal *= GenNum - int64(i)
	}
	for i := GenTransOptN; i >= 1; i-- {
		genTransOptTotal /= int64(i)
	}

	in, err := os.Open("texas_data_" + strconv.Itoa(GenTransOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	f, err := os.Create("texas_data_opt_" + strconv.Itoa(GenTransOptN) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	genTransOptTotalKey = 0
	genTransOptLastPrint = 0
	genTransOptBeginPrint = time.Now().UnixMilli()
	for k := range genTransOptKeys {
		delete(genTransOptKeys, k)
	}
	for k := range genTransOptOptKeys {
		delete(genTransOptOptKeys, k)
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, _ := strconv.ParseInt(params[0], 10, 64)
		p := params[1]
		min := params[2]
		max := params[3]
		value := p + " " + min + " " + max
		genTransOptKeys[key] = value
		removeKey := RemoveColor(key)
		optKeyData := genTransOptOptKeys[removeKey]
		if optKeyData == nil {
			optKeyData = &GenTransOptKeyData{Ps: map[string]int{}}
			genTransOptOptKeys[removeKey] = optKeyData
		}
		optKeyData.Ps[value]++

		genTransOptTotalKey++

		cur := int(genTransOptTotalKey * 100 / genTransOptTotal)
		if cur != genTransOptLastPrint {
			genTransOptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genTransOptBeginPrint) / float64(genTransOptTotalKey)
			fmt.Println("step1 N" + strconv.Itoa(GenTransOptN) + " " + fmt.Sprint(cur) + "% 需要" +
				fmt.Sprint(per*float64(genTransOptTotal-genTransOptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genTransOptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genTransOptTotalKey)/(float64(now-genTransOptBeginPrint)/1000)) + "条/秒")
		}
	}

	for _, optKeyData := range genTransOptOptKeys {
		max := 0
		for v, num := range optKeyData.Ps {
			if num > max {
				optKeyData.Max = v
				max = num
			}
		}
	}

	genTransOptTotalKey = 0
	genTransOptLastPrint = 0
	genTransOptBeginPrint = time.Now().UnixMilli()

	for key, p := range genTransOptKeys {
		removeKey := RemoveColor(key)
		optKeyData := genTransOptOptKeys[removeKey]
		if optKeyData == nil || optKeyData.Max != p {
			tmp := fmt.Sprint(key) + " 0 " + p + " " + ToString(key) + "\n"
			w.WriteString(tmp)
		}

		genTransOptTotalKey++

		cur := int(genTransOptTotalKey * 100 / genTransOptTotal)
		if cur != genTransOptLastPrint {
			genTransOptLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genTransOptBeginPrint) / float64(genTransOptTotalKey)
			fmt.Println("step2 N" + strconv.Itoa(GenTransOptN) + " " + fmt.Sprint(cur) + "% 需要" +
				fmt.Sprint(per*float64(genTransOptTotal-genTransOptTotalKey)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genTransOptBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(genTransOptTotalKey)/(float64(now-genTransOptBeginPrint)/1000)) + "条/秒")
		}
	}

	for key, optKeyData := range genTransOptOptKeys {
		tmp := fmt.Sprint(key) + " 1 " + optKeyData.Max + " " + ToString(key) + "\n"
		w.WriteString(tmp)
	}
	w.Flush()
	f.Close()
	for k := range genTransOptKeys {
		delete(genTransOptKeys, k)
	}
	for k := range genTransOptOptKeys {
		delete(genTransOptOptKeys, k)
	}

	fmt.Println("optData finish", genTransOptTotalKey)
}
