package texas_algorithm

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"time"
)

// Number of wild cards in the deck, aligned with Java's GenUtil.guiNum.
const GuiNum = 2

// GenNum is the total number of cards in the deck (52 + 2 jokers).
const GenNum = 52 + GuiNum

// Total is the number of 7-card combinations: C(54, 7).
var Total int64 = (GenNum * (GenNum - 1) * (GenNum - 2) * (GenNum - 3) * (GenNum - 4) * (GenNum - 5) * (GenNum - 6)) /
	(7 * 6 * 5 * 4 * 3 * 2)

// Package-level state used while generating the data tables, mirroring the
// static fields of Java's GenUtil.
var (
	GenTotalKey   int64
	GenLastPrint  int
	GenBeginPrint int64
	// UseOpt switches GenCompare to the lookup-table comparator once the
	// tables are loaded, mirroring Java's GenUtil.useOpt.
	UseOpt bool
	// Keys accumulates the generated 7-card keys.
	Keys []int64
	// Progress counts sorted elements, updated by the parallel quicksort.
	Progress atomic.Int64
	// LastPrint is the last printed progress permille.
	LastPrint atomic.Int64
	// BeginPrint is the millisecond timestamp when sorting started.
	BeginPrint atomic.Int64
)

// GenAllPokes returns all 54 cards (52 regular cards + 2 jokers).
func GenAllPokes() []Poke {
	list := []Poke{}
	for _, t := range genAllCards() {
		list = append(list, NewPokeFromByte(byte(t)))
	}
	return list
}

// GenAllCards returns all 54 cards as encoded bytes, sorted ascending.
func GenAllCards() []int {
	return genAllCards()
}

// AllCards is the full 54-card deck as encoded bytes, sorted ascending,
// mirroring Java's GenUtil.allCards.
var AllCards = genAllCards()

func genAllCards() []int {
	var list []int
	for i := byte(0); i < 4; i++ {
		for j := byte(0); j < byte(GenNum/4); j++ {
			list = append(list, int(NewPoke(i, j+2).ToByte()))
		}
	}
	for i := 0; i < GuiNum; i++ {
		list = append(list, int(GUI.ToByte()))
	}
	sort.Ints(list)
	return list
}

// GenKey enumerates all C(54, 7) combinations into Keys.
func GenKey() {
	GenBeginPrint = time.Now().UnixMilli()

	genCard()

	fmt.Println("genKey finish", Total)
}

func genCard() {
	list := genAllCards()

	tmp := make([]int, 7)
	Permutation(list, 0, 0, 7, tmp, func(tmp []int) {
		genCardSave(tmp)
	})
}

// Permutation generates combinations of `except` elements from `a` with
// strictly increasing indices, calling run for each combination,
// mirroring Java's GenUtil.permutation.
func Permutation(a []int, count, count2, except int, tmp []int, run func(tmp []int)) {
	if count2 == except {
		run(tmp)
	} else {
		for i := count; i < len(a); i++ {
			tmp[count2] = a[i]
			Permutation(a, i+1, count2+1, except, tmp, run)
		}
	}
}

func genCardSave(tmp []int) {
	c := GenCardBindInts(tmp)

	Keys = append(Keys, c)
	GenTotalKey++

	cur := int(GenTotalKey * 100 / Total)
	if cur != GenLastPrint {
		GenLastPrint = cur

		now := time.Now().UnixMilli()
		per := float64(now-GenBeginPrint) / float64(GenTotalKey)
		fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(Total-GenTotalKey)/60/1000) + "分" +
			" 用时" + fmt.Sprint((now-GenBeginPrint)/60/1000) + "分" + " 速度" +
			fmt.Sprint(float64(GenTotalKey)/(float64(now-GenBeginPrint)/1000)) + "条/秒")
	}
}

// GenCardBindInts packs card bytes into a key: ret = ret * 100 + card.
func GenCardBindInts(tmp []int) int64 {
	var ret int64
	for _, i := range tmp {
		ret = ret*100 + int64(i)
	}
	return ret
}

// GenCardBindBytes packs card bytes into a key: ret = ret * 100 + card.
func GenCardBindBytes(tmp []byte) int64 {
	var ret int64
	for _, i := range tmp {
		ret = ret*100 + int64(i)
	}
	return ret
}

// OutputData sorts Keys by poker strength and writes texas_data.txt.
func OutputData() {
	begin := time.Now().UnixMilli()

	if UseOpt {
		Load()
	}

	f, err := os.Create("texas_data.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	w := bufio.NewWriter(f)

	GenBeginPrint = time.Now().UnixMilli()
	GenLastPrint = 0

	Quicksort(Keys)

	GenTotalKey = 0
	GenLastPrint = 0
	GenBeginPrint = time.Now().UnixMilli()
	i := 0
	var iindex int64
	var index int64
	var lastMax int64
	size := int64(len(Keys))
	for _, k := range Keys {
		curMax := Max(Keys[index])
		var str string
		row := func() string {
			return fmt.Sprint(k) + " " + fmt.Sprint(i) + " " + fmt.Sprint(iindex) + " " + fmt.Sprint(size) + " " +
				ToString(Keys[index]) + " " + fmt.Sprint(curMax) + " " + ToString(curMax) + " " +
				fmt.Sprint(MaxType(Keys[index])) + "\n"
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

		cur := int(index * 100 / Total)
		if cur != GenLastPrint {
			GenLastPrint = cur

			now := time.Now().UnixMilli()
			per := float64(now-GenBeginPrint) / float64(index)
			fmt.Println(fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(Total-index)/60/1000) + "分" +
				" 用时" + fmt.Sprint((now-GenBeginPrint)/60/1000) + "分" + " 速度" +
				fmt.Sprint(float64(index)/(float64(now-GenBeginPrint)/1000)) + "条/秒")
		}
	}

	w.Flush()
	f.Close()

	fmt.Println("outputData finish " + fmt.Sprint(Total) + " time:" +
		fmt.Sprint((time.Now().UnixMilli()-begin)/1000/60) + "分")

	Keys = nil
}

// pickBest extracts the best 5 cards out of 5/6/7 cards.
func pickBest(cs []Poke) []Poke {
	var pickedCards []Poke
	switch len(cs) {
	case 7:
		pickedCards = FiveFromSeven(cs)
	case 6:
		pickedCards = FiveFromSix(cs)
	case 5:
		pickedCards = FiveFromFive(cs)
	}
	return pickedCards
}

// Max returns the encoded key of the best 5-card hand within key k.
func Max(k int64) int64 {
	cs := ToArray(k)
	pickedCards := pickBest(cs)

	var ret int64
	for _, p := range pickedCards {
		ret = ret*100 + int64(p.ToByte())
	}
	return ret
}

// MaxType returns the card type of the best 5-card hand within key k.
func MaxType(k int64) int {
	cs := ToArray(k)
	pickedCards := pickBest(cs)

	return GetCardTypeUnorderedWithoutGui(pickedCards)
}

// ToArray decodes a key back into its card list, from the most significant
// card down, mirroring Java's GenUtil.toArray.
func ToArray(k int64) []Poke {
	var cs []Poke
	if k > 1000000000000 {
		cs = append(cs, NewPokeFromByte(byte(k%100000000000000/1000000000000)))
	}
	if k > 10000000000 {
		cs = append(cs, NewPokeFromByte(byte(k%1000000000000/10000000000)))
	}
	if k > 100000000 {
		cs = append(cs, NewPokeFromByte(byte(k%10000000000/100000000)))
	}
	if k > 1000000 {
		cs = append(cs, NewPokeFromByte(byte(k%100000000/1000000)))
	}
	if k > 10000 {
		cs = append(cs, NewPokeFromByte(byte(k%1000000/10000)))
	}
	if k > 100 {
		cs = append(cs, NewPokeFromByte(byte(k%10000/100)))
	}
	if k > 1 {
		cs = append(cs, NewPokeFromByte(byte(k%100/1)))
	}
	return cs
}

// ToString renders a key as its concatenated card strings, e.g. "方A黑K".
func ToString(k int64) string {
	cs := ToArray(k)
	ret := ""
	for _, poke := range cs {
		ret += poke.String()
	}
	return ret
}

// GenCompare reports whether key k1 is weaker than key k2 by poker strength,
// mirroring Java's GenUtil.compare(long, long). When UseOpt is set and the
// tables are loaded, the lookup table is used as the comparator.
func GenCompare(k1, k2 int64) bool {
	if UseOpt {
		return CompareKey(k1, k2) < 0
	}

	pickedCards1 := pickBest(ToArray(k1))
	pickedCards2 := pickBest(ToArray(k2))

	return CompareCardsWithoutGui(pickedCards1, pickedCards2) < 0
}

// Equal reports whether the two 5-card best-hand keys are equal in strength,
// mirroring Java's GenUtil.equal.
func Equal(k1, k2 int64) bool {
	cs1 := []Poke{
		NewPokeFromByte(byte(k1 % 10000000000 / 100000000)),
		NewPokeFromByte(byte(k1 % 100000000 / 1000000)),
		NewPokeFromByte(byte(k1 % 1000000 / 10000)),
		NewPokeFromByte(byte(k1 % 10000 / 100)),
		NewPokeFromByte(byte(k1 % 100 / 1)),
	}
	sortPokesDesc(cs1)

	cs2 := []Poke{
		NewPokeFromByte(byte(k2 % 10000000000 / 100000000)),
		NewPokeFromByte(byte(k2 % 100000000 / 1000000)),
		NewPokeFromByte(byte(k2 % 1000000 / 10000)),
		NewPokeFromByte(byte(k2 % 10000 / 100)),
		NewPokeFromByte(byte(k2 % 100 / 1)),
	}
	sortPokesDesc(cs2)

	return CompareCardsWithoutGui(cs1, cs2) == 0
}
