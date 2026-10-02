package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	ta "github.com/esrrhs/texas_algorithm/go"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		demo()
		return
	}

	ta.Gen()
	ta.GenExtra()
	ta.GenOpt()
	ta.GenExtraOpt()
	ta.GenTrans()
	ta.GenTransOpt()
	ta.GenHand()
	fmt.Println("done!")
}

// demo mirrors Java's TestUtil.main.
func demo() {
	ta.Load()

	best, gui := ta.GetMaxStr("方2,梅3,黑2,黑4,鬼")
	fmt.Println(best)
	fmt.Println(ta.PokesToStr(gui))

	best, gui = ta.GetMaxStr("方2,方3,方4,鬼,鬼,黑6,红6")
	fmt.Println(best)
	fmt.Println(ta.PokesToStr(gui))

	best, gui = ta.GetMaxStr("黑A,方J,黑8,红8,鬼")
	fmt.Println(best)
	fmt.Println(ta.PokesToStr(gui))

	best, gui = ta.GetMaxStr("方2,鬼,黑2,黑4,黑5,鬼")
	fmt.Println(best)
	fmt.Println(ta.PokesToStr(gui))

	cards := "方4,方A,黑2,黑A,黑3,黑5,黑6"
	cards1 := "红8,方A,方2,黑8,黑3,黑5,黑7"
	fmt.Println(ta.GetWinPositionStr(cards))
	fmt.Println(ta.GetWinProbabilityStr(cards))
	fmt.Println(ta.KeyToStr(ta.GetWinMaxStr(cards)))
	fmt.Println(ta.GetWinTypeStr(cards))

	fmt.Println(ta.GetWinPositionStr(cards1))
	fmt.Println(ta.GetWinProbabilityStr(cards1))
	fmt.Println(ta.KeyToStr(ta.GetWinMaxStr(cards1)))
	fmt.Println(ta.GetWinTypeStr(cards1))

	fmt.Println(ta.CompareStr(cards, cards1))

	cards2 := "红8,方A,方2,黑8,黑5,黑7"
	fmt.Println(ta.GetWinPositionStr(cards2))
	fmt.Println(ta.GetWinProbabilityStr(cards2))
	fmt.Println(ta.KeyToStr(ta.GetWinMaxStr(cards2)))
	fmt.Println(ta.GetWinTypeStr(cards2))

	cards3 := "红8,方A,方2,黑5,黑7"
	fmt.Println(ta.GetWinPositionStr(cards3))
	fmt.Println(ta.GetWinProbabilityStr(cards3))
	fmt.Println(ta.KeyToStr(ta.GetWinMaxStr(cards3)))
	fmt.Println(ta.GetWinTypeStr(cards3))

	best, _ = ta.GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6")
	fmt.Println(best)
	best, _ = ta.GetMaxStrHandPub("方4,方2", "黑2,黑A,黑7,黑5,黑6")
	fmt.Println(best)
	best, _ = ta.GetMaxStrHandPub("黑2,黑3", "方2,方A,黑7,黑5,黑6")
	fmt.Println(best)
	best, _ = ta.GetMaxStrHandPub("黑2,黑3", "方2,方A,黑7,黑5")
	fmt.Println(best)
	best, _ = ta.GetMaxStrHandPub("黑2,黑3", "方2,黑7,黑5")
	fmt.Println(best)

	ta.LoadProbility()
	fmt.Println(ta.GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K"))
	fmt.Println(ta.GetHandProbabilityStr("方2,方3", ""))
	fmt.Println(ta.GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K,方A"))

	compare("hand4/texas_hand_方3方10.txt")
}

// compare checks the stored exhaustive win rates against the estimated ones,
// mirroring Java's TestUtil.compare.
func compare(file string) {
	total := 0
	diff1 := 0
	diff2 := 0

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), " ")
		key, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		probility, err := strconv.ParseFloat(params[1], 32)
		if err != nil {
			fmt.Println(err)
			return
		}
		card := params[2]

		p := ta.GetHandProbabilityKeyPub(key/100000000, key%100000000)

		if float64(p)-probility > 0.1 || probility-float64(p) > 0.1 {
			fmt.Println("diff " + strconv.FormatFloat(float64(p)-probility, 'g', -1, 64) + " " + card + " " +
				strconv.FormatFloat(float64(p), 'g', -1, 64) + " " + strconv.FormatFloat(probility, 'g', -1, 64))
			diff1++
		}
		if float64(p)-probility > 0.2 || probility-float64(p) > 0.2 {
			fmt.Println("diff " + strconv.FormatFloat(float64(p)-probility, 'g', -1, 64) + " " + card + " " +
				strconv.FormatFloat(float64(p), 'g', -1, 64) + " " + strconv.FormatFloat(probility, 'g', -1, 64))
			diff2++
		}

		total++
	}

	fmt.Println("diff>0.1 = %" + strconv.Itoa(diff1*100/total))
	fmt.Println("diff>0.1 = " + strconv.Itoa(diff1))
	fmt.Println("diff>0.2 = %" + strconv.Itoa(diff2*100/total))
	fmt.Println("diff>0.2 = " + strconv.Itoa(diff2))
	fmt.Println("total " + strconv.Itoa(total))
}
