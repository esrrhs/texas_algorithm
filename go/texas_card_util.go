package texas_algorithm

import "sort"

// Card type constants, aligned with Java's TexasCardUtil.
const (
	TexasCardTypeGaoPai          = 1  // 高牌 high card
	TexasCardTypeDuiZi           = 2  // 对子 one pair
	TexasCardTypeLiangDui        = 3  // 两对 two pairs
	TexasCardTypeSanTiao         = 4  // 三条 three of a kind
	TexasCardTypeShunZi          = 5  // 顺子 straight
	TexasCardTypeTongHua         = 6  // 同花 flush
	TexasCardTypeHuLu            = 7  // 葫芦 full house
	TexasCardTypeSiTiao          = 8  // 四条 four of a kind
	TexasCardTypeTongHuaShun     = 9  // 同花顺 straight flush
	TexasCardTypeKingTongHuaShun = 10 // 皇家同花顺 royal flush
)

type analyseResult struct {
	fourCount       int
	threeCount      int
	twoCount        int
	oneCount        int
	fourLogicValue  []int
	threeLogicValue []int
	twoLogicValue   []int
	oneLogicValue   []int
}

// sortPokesDesc sorts cards by logic value descending, the same order as
// Java's TexasPokeLogicValueComparator.
func sortPokesDesc(cards []Poke) {
	sort.SliceStable(cards, func(i, j int) bool { return cards[i].Value > cards[j].Value })
}

// GetCardTypeUnorderedWithoutGui sorts the cards by value descending in place
// and returns their type, mirroring Java's getCardTypeUnorderedWithoutGui.
func GetCardTypeUnorderedWithoutGui(cards []Poke) int {
	sortPokesDesc(cards)
	return getCardType(cards)
}

// getCardType returns the type of 5 sorted cards, mirroring Java's private
// getCardType. cards must be sorted descending by value and have length 5.
func getCardType(cards []Poke) int {
	if len(cards) != 5 {
		return 0
	}
	sameColor := true
	lineCard := true
	firstColor := cards[0].Color
	firstValue := int(cards[0].Value)

	// 牌型分析
	for i := 1; i < 5; i++ {
		if cards[i].Color != firstColor {
			sameColor = false
		}
		if int(cards[i].Value)+i != firstValue {
			lineCard = false
		}
		if !sameColor && !lineCard {
			break
		}
	}

	// 最小顺子
	if !lineCard && firstValue == 14 {
		i := 1
		for ; i < 5; i++ {
			value := int(cards[i].Value)
			if value+i+8 != firstValue {
				break
			}
		}
		if i == 5 {
			lineCard = true
		}
	}

	// 皇家同花顺
	if sameColor && lineCard && int(cards[1].Value) == 13 {
		return TexasCardTypeKingTongHuaShun
	}

	// 同花顺
	if sameColor && lineCard {
		return TexasCardTypeTongHuaShun
	}

	// 顺子
	if !sameColor && lineCard {
		return TexasCardTypeShunZi
	}

	// 同花
	if sameColor && !lineCard {
		return TexasCardTypeTongHua
	}

	// 分析相同牌有几个
	analyseRes := analyseCards(cards)
	if analyseRes.fourCount == 1 {
		return TexasCardTypeSiTiao // 四条
	}
	if analyseRes.threeCount == 1 && analyseRes.twoCount == 1 {
		return TexasCardTypeHuLu // 葫芦
	}
	if analyseRes.threeCount == 1 && analyseRes.twoCount == 0 {
		return TexasCardTypeSanTiao // 三条
	}
	if analyseRes.twoCount == 2 {
		return TexasCardTypeLiangDui // 两对
	}
	if analyseRes.twoCount == 1 && analyseRes.oneCount == 3 {
		return TexasCardTypeDuiZi // 对子
	}

	return TexasCardTypeGaoPai // 高牌
}

// analyseCards groups cards by value; cards must be sorted descending by value.
func analyseCards(cards []Poke) analyseResult {
	var res analyseResult
	for i := 0; i < len(cards); i++ {
		sameCount := 1
		cardValue := int(cards[i].Value)

		for j := i + 1; j < len(cards); j++ {
			if int(cards[j].Value) != cardValue {
				break
			}
			sameCount++
		}

		switch sameCount {
		case 1:
			res.oneCount++
			res.oneLogicValue = append(res.oneLogicValue, cardValue)
		case 2:
			res.twoCount++
			res.twoLogicValue = append(res.twoLogicValue, cardValue)
		case 3:
			res.threeCount++
			res.threeLogicValue = append(res.threeLogicValue, cardValue)
		case 4:
			res.fourCount++
			res.fourLogicValue = append(res.fourLogicValue, cardValue)
		}
		i += sameCount - 1
	}

	return res
}

// FiveFromSeven picks the best 5 cards out of 7, handling wild cards.
func FiveFromSeven(cards []Poke) []Poke {
	gui := 0
	for _, poke := range cards {
		if poke.IsGui() {
			gui++
		}
	}

	if gui == 0 {
		return FiveFromSevenWithoutGui(cards)
	}

	var left []Poke
	leftMap := map[int]bool{}
	for _, poke := range cards {
		if !poke.IsGui() {
			left = append(left, poke)
			leftMap[int(poke.ToByte())] = true
		}
	}

	maxCards := []Poke{}
	tmp := make([]int, gui)
	Permutation(AllCards, 0, 0, gui, tmp, func(tmp []int) {
		for _, t := range tmp {
			if leftMap[t] || t == int(GUI.ToByte()) {
				return
			}
		}

		last := make([]Poke, 0, len(left)+len(tmp))
		last = append(last, left...)
		for _, t := range tmp {
			last = append(last, NewPokeFromByte(byte(t)))
		}
		picked := FiveFromSevenWithoutGui(last)
		if len(maxCards) == 0 || CompareCardsWithoutGui(picked, maxCards) > 0 {
			maxCards = picked
		}
	})

	return maxCards
}

// FiveFromSevenWithoutGui picks the best 5 cards out of 7 without wild cards.
func FiveFromSevenWithoutGui(cards []Poke) []Poke {
	tmpSevenCards := make([]Poke, len(cards))
	copy(tmpSevenCards, cards)
	// 对七张牌排序
	sortPokesDesc(tmpSevenCards)

	pickedCards := append([]Poke{}, tmpSevenCards[:5]...)

	for i := 0; i < 3; i++ {
		for j := i + 1; j < 4; j++ {
			for k := j + 1; k < 5; k++ {
				for l := k + 1; l < 6; l++ {
					for m := l + 1; m < 7; m++ {
						tmpPickedCards := []Poke{tmpSevenCards[i], tmpSevenCards[j], tmpSevenCards[k], tmpSevenCards[l], tmpSevenCards[m]}

						if CompareCardsWithoutGui(tmpPickedCards, pickedCards) == 1 {
							// 找到更大的牌 进行替换
							pickedCards = tmpPickedCards
						}
					}
				}
			}
		}
	}
	return pickedCards
}

// FiveFromFive evaluates 5 cards, using wild cards if present.
func FiveFromFive(cards []Poke) []Poke {
	gui := 0
	for _, poke := range cards {
		if poke.IsGui() {
			gui++
		}
	}

	if gui == 0 {
		return FiveFromFiveWithoutGui(cards)
	}

	var left []Poke
	leftMap := map[int]bool{}
	for _, poke := range cards {
		if !poke.IsGui() {
			left = append(left, poke)
			leftMap[int(poke.ToByte())] = true
		}
	}

	maxCards := []Poke{}
	tmp := make([]int, gui)
	Permutation(AllCards, 0, 0, gui, tmp, func(tmp []int) {
		for _, t := range tmp {
			if leftMap[t] || t == int(GUI.ToByte()) {
				return
			}
		}

		last := make([]Poke, 0, len(left)+len(tmp))
		last = append(last, left...)
		for _, t := range tmp {
			last = append(last, NewPokeFromByte(byte(t)))
		}
		picked := FiveFromFiveWithoutGui(last)
		if len(maxCards) == 0 || CompareCardsWithoutGui(picked, maxCards) > 0 {
			maxCards = picked
		}
	})

	return maxCards
}

// FiveFromFiveWithoutGui sorts the 5 cards by value descending.
func FiveFromFiveWithoutGui(cards []Poke) []Poke {
	pickedCards := make([]Poke, len(cards))
	copy(pickedCards, cards)
	sortPokesDesc(pickedCards)
	return pickedCards
}

// FiveFromSix picks the best 5 cards out of 6, handling wild cards.
func FiveFromSix(cards []Poke) []Poke {
	gui := 0
	for _, poke := range cards {
		if poke.IsGui() {
			gui++
		}
	}

	if gui == 0 {
		return FiveFromSixWithoutGui(cards)
	}

	var left []Poke
	leftMap := map[int]bool{}
	for _, poke := range cards {
		if !poke.IsGui() {
			left = append(left, poke)
			leftMap[int(poke.ToByte())] = true
		}
	}

	maxCards := []Poke{}
	tmp := make([]int, gui)
	Permutation(AllCards, 0, 0, gui, tmp, func(tmp []int) {
		for _, t := range tmp {
			if leftMap[t] || t == int(GUI.ToByte()) {
				return
			}
		}

		last := make([]Poke, 0, len(left)+len(tmp))
		last = append(last, left...)
		for _, t := range tmp {
			last = append(last, NewPokeFromByte(byte(t)))
		}
		picked := FiveFromSixWithoutGui(last)
		if len(maxCards) == 0 || CompareCardsWithoutGui(picked, maxCards) > 0 {
			maxCards = picked
		}
	})

	return maxCards
}

// FiveFromSixWithoutGui picks the best 5 cards out of 6 without wild cards.
func FiveFromSixWithoutGui(cards []Poke) []Poke {
	tmpSixCards := make([]Poke, len(cards))
	copy(tmpSixCards, cards)
	// 对六张牌排序
	sortPokesDesc(tmpSixCards)

	pickedCards := append([]Poke{}, tmpSixCards[:5]...)

	for i := 0; i < 3; i++ {
		for j := i + 1; j < 4; j++ {
			for k := j + 1; k < 5; k++ {
				for l := k + 1; l < 6; l++ {
					for m := l + 1; m < 6; m++ {
						tmpPickedCards := []Poke{tmpSixCards[i], tmpSixCards[j], tmpSixCards[k], tmpSixCards[l], tmpSixCards[m]}

						if CompareCardsWithoutGui(tmpPickedCards, pickedCards) == 1 {
							// 找到更大的牌 进行替换
							pickedCards = tmpPickedCards
						}
					}
				}
			}
		}
	}
	return pickedCards
}

// CompareCardsWithoutGui compares two 5-card hands of the same card count.
// Both hands must be sorted descending by value. Returns 1 if firstCards is
// stronger, -1 if weaker, 0 if equal.
func CompareCardsWithoutGui(firstCards, secondCards []Poke) int {
	firstType := getCardType(firstCards)
	secondType := getCardType(secondCards)

	if firstType > secondType {
		return 1
	}
	if firstType < secondType {
		return -1
	}
	switch firstType {
	case TexasCardTypeGaoPai: // 单牌
		for i := 0; i < 5; i++ {
			secondValue := int(secondCards[i].Value)
			firstValue := int(firstCards[i].Value)
			if firstValue > secondValue {
				return 1
			} else if firstValue < secondValue {
				return -1
			}
		}
		return 0
	case TexasCardTypeDuiZi, TexasCardTypeLiangDui, TexasCardTypeSanTiao, TexasCardTypeSiTiao, TexasCardTypeHuLu:
		secondAnalyseRes := analyseCards(secondCards)
		firstAnalyseRes := analyseCards(firstCards)
		// 四条比较
		if firstAnalyseRes.fourCount > 0 {
			secondValue := secondAnalyseRes.fourLogicValue[0]
			firesValue := firstAnalyseRes.fourLogicValue[0]
			// 比较四条的值
			if firesValue != secondValue {
				if firesValue > secondValue {
					return 1
				}
				return -1
			}
			// 比较单牌的值
			secondValue = secondAnalyseRes.oneLogicValue[0]
			firesValue = firstAnalyseRes.oneLogicValue[0]
			if firesValue != secondValue {
				if firesValue > secondValue {
					return 1
				}
				return -1
			}
			return 0
		}

		// 三条比较
		if firstAnalyseRes.threeCount > 0 {
			secondValue := secondAnalyseRes.threeLogicValue[0]
			firstValue := firstAnalyseRes.threeLogicValue[0]
			// 比较三条
			if firstValue != secondValue {
				if firstValue > secondValue {
					return 1
				}
				return -1
			}

			// 葫芦牌型
			if firstType == TexasCardTypeHuLu {
				secondValue = secondAnalyseRes.twoLogicValue[0]
				firstValue = firstAnalyseRes.twoLogicValue[0]
				if firstValue != secondValue {
					if firstValue > secondValue {
						return 1
					}
					return -1
				}
				return 0
			}
			// 三条带单
			for i := 0; i < firstAnalyseRes.oneCount; i++ {
				secondValue := secondAnalyseRes.oneLogicValue[i]
				firstValue := firstAnalyseRes.oneLogicValue[i]
				if firstValue > secondValue {
					return 1
				} else if firstValue < secondValue {
					return -1
				}
			}
			return 0
		}

		// 对子
		for i := 0; i < firstAnalyseRes.twoCount; i++ {
			secondValue := secondAnalyseRes.twoLogicValue[i]
			firstValue := firstAnalyseRes.twoLogicValue[i]
			if firstValue > secondValue {
				return 1
			} else if firstValue < secondValue {
				return -1
			}
		}
		// 比较单牌
		for i := 0; i < firstAnalyseRes.oneCount; i++ {
			secondValue := secondAnalyseRes.oneLogicValue[i]
			firstValue := firstAnalyseRes.oneLogicValue[i]
			if firstValue > secondValue {
				return 1
			} else if firstValue < secondValue {
				return -1
			}
		}
		return 0
	case TexasCardTypeShunZi, TexasCardTypeTongHuaShun:
		secondValue := int(secondCards[0].Value)
		firstValue := int(firstCards[0].Value)

		// 是最小顺子吗 第一张5 第五张14
		firstMin := int(firstCards[4].Value) == (int(firstCards[0].Value) - 12)
		secondMin := int(secondCards[4].Value) == (int(secondCards[0].Value) - 12)

		if firstMin && !secondMin {
			// 第一个是最小顺子 第一个小于第二个
			return -1
		}
		if !firstMin && secondMin {
			return 1
		}
		if firstValue == secondValue {
			return 0
		}
		if firstValue > secondValue {
			return 1
		}
		return -1
	case TexasCardTypeTongHua:
		for i := 0; i < 5; i++ {
			secondValue := int(secondCards[i].Value)
			firstValue := int(firstCards[i].Value)

			if firstValue == secondValue {
				continue
			}
			if firstValue > secondValue {
				return 1
			}
			return -1
		}
		return 0
	default:
		return 0
	}
}
