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

// demo mirrors Java's TestUtil.main. The 52-card deck has no joker card, so
// every API returns an error for input it cannot recognise.
func demo() {
	ta.Load()

	best, err := ta.GetMaxStr("方2,梅3,黑2,黑4,黑5")
	fmt.Println(best, err)

	best, err = ta.GetMaxStr("方2,方3,方4,方5,黑6,红6,黑7")
	fmt.Println(best, err)

	cards := "方4,方A,黑2,黑A,黑3,黑5,黑6"
	cards1 := "红8,方A,方2,黑8,黑3,黑5,黑7"
	showHand(cards)
	showHand(cards1)

	cmp, err := ta.CompareStr(cards, cards1)
	fmt.Println(cmp, err)

	cards2 := "红8,方A,方2,黑8,黑5,黑7"
	cards3 := "红8,方A,方2,黑5,黑7"
	showHand(cards2)
	showHand(cards3)

	for _, pub := range []string{
		"黑2,黑A,方3,黑5,黑6",
		"黑2,黑A,黑7,黑5,黑6",
		"方2,方A,黑7,黑5,黑6",
		"方2,方A,黑7,黑5",
		"方2,黑7,黑5",
	} {
		best, err := ta.GetMaxStrHandPub("黑2,黑3", pub)
		fmt.Println(best, err)
	}

	ta.LoadProbility()
	p, err := ta.GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K")
	fmt.Println(p, err)
	p, err = ta.GetHandProbabilityStr("方2,方3", "")
	fmt.Println(p, err)
	p, err = ta.GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K,方A")
	fmt.Println(p, err)

	// The removed joker card is no longer accepted anywhere.
	_, err = ta.GetMaxStr("方2,梅3,黑2,黑4,鬼")
	fmt.Println(err)
	_, err = ta.GetWinPosition([]byte{0x22, 0x13, 0x54, 0x58, 0x35, 0x36, 0x37})
	fmt.Println(err)

	compare("hand4/texas_hand_方3方10.txt")
}

func showHand(cards string) {
	pos, err := ta.GetWinPositionStr(cards)
	if err != nil {
		fmt.Println(err)
		return
	}
	prob, _ := ta.GetWinProbabilityStr(cards)
	max, _ := ta.GetWinMaxStr(cards)
	typ, _ := ta.GetWinTypeStr(cards)
	fmt.Println(pos)
	fmt.Println(prob)
	fmt.Println(ta.KeyToStr(max))
	fmt.Println(typ)
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

		p, err := ta.GetHandProbabilityKeyPub(key/100000000, key%100000000)
		if err != nil {
			fmt.Println(err)
			return
		}

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
