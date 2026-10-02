package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// Package-level state for the N-card (5 or 6) table generation, mirroring the
// static fields of Java's GenExtraUtil. Set GenExtraN before generating.
var (
	GenExtraN          = 6
	genExtraTotal      int64
	genExtraTotalKey   int64
	genExtraLastPrint  int
	genExtraBeginPrint int64
	genExtraKeys       []int64
	genExtraProgress   atomic.Int64
)

// GenExtraGenKey enumerates all C(54, GenExtraN) combinations into genExtraKeys.
func GenExtraGenKey() {
	genExtraTotal = 1
	for i := 0; i < GenExtraN; i++ {
		genExtraTotal *= GenNum - int64(i)
	}
	for i := GenExtraN; i >= 1; i-- {
		genExtraTotal /= int64(i)
	}
	genExtraBeginPrint = time.Now().UnixMilli()

	genExtraCard()

	fmt.Println("genKey finish", genExtraTotal)
}

func genExtraCard() {
	list := genAllCards()

	tmp := make([]int, GenExtraN)
	Permutation(list, 0, 0, GenExtraN, tmp, func(tmp []int) {
		genExtraCardSave(tmp)
	})
}

func genExtraCardSave(tmp []int) {
	c := GenCardBindInts(tmp)

	genExtraKeys = append(genExtraKeys, c)
	genExtraTotalKey++

	cur := int(genExtraTotalKey * 100 / genExtraTotal)
	if cur != genExtraLastPrint {
		genExtraLastPrint = cur

		now := time.Now().UnixMilli()
		per := float64(now-genExtraBeginPrint) / float64(genExtraTotalKey)
		fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(genExtraTotal-genExtraTotalKey)/60/1000) + "分" +
			" 用时" + fmt.Sprint((now-genExtraBeginPrint)/60/1000) + "分" + " 速度" +
			fmt.Sprint(float64(genExtraTotalKey)/(float64(now-genExtraBeginPrint)/1000)) + "条/秒")
	}
}

// GenExtraOutputData sorts genExtraKeys by poker strength and writes
// texas_data_extra_N.txt.
func GenExtraOutputData() {
	begin := time.Now().UnixMilli()

	if UseOpt {
		Load()
	}

	name := "texas_data_extra_" + strconv.Itoa(GenExtraN) + ".txt"
	f, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	genExtraBeginPrint = time.Now().UnixMilli()
	genExtraLastPrint = 0

	Quicksort(genExtraKeys)

	genExtraTotalKey = 0
	genExtraLastPrint = 0
	genExtraBeginPrint = time.Now().UnixMilli()
	i := 0
	var iindex int64
	var index int64
	var lastMax int64
	size := int64(len(genExtraKeys))
	for _, k := range genExtraKeys {
		curMax := Max(genExtraKeys[index])
		var str string
		row := func() string {
			return fmt.Sprint(k) + " " + fmt.Sprint(i) + " " + fmt.Sprint(iindex) + " " + fmt.Sprint(size) + " " +
				ToString(genExtraKeys[index]) + " " + fmt.Sprint(curMax) + " " + ToString(curMax) + " " +
				fmt.Sprint(MaxType(genExtraKeys[index])) + "\n"
		}
		if lastMax == 0 {
			str = row()
			lastMax = curMax
			iindex = index
		} else {
			if Equal(lastMax, curMax) {
				str = row()
				lastMax = curMax
			} else {
				i++
				iindex = index
				str = row()
				lastMax = curMax
			}
		}

		w.WriteString(str)
		index++

		cur := int(index * 100 / genExtraTotal)
		if cur != genExtraLastPrint {
			genExtraLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-genExtraBeginPrint) / float64(index)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(genExtraTotal-index)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-genExtraBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(index)/(float64(now-genExtraBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()

	fmt.Println("outputData finish " + fmt.Sprint(genExtraTotal) + " time:" +
		fmt.Sprint((time.Now().UnixMilli()-begin)/1000/60) + "分 " + fmt.Sprint(genExtraProgress.Load()))

	genExtraKeys = nil
}
