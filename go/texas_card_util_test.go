package texas_algorithm

import (
	"os"
	"testing"
)

// The two test classes below mirror Java's TexasCardUtilTest and
// TexasAlgorithmUtilTest one to one. Like the Java version, tests that need
// the generated data tables skip gracefully when the files are absent.

var (
	dataAvailable     = false
	probDataAvailable = false
)

func TestMain(m *testing.M) {
	if fileExists("texas_data_color.txt") {
		Load()
		dataAvailable = IsLoaded()
	}

	if fileExists("texas_data_opt_2.txt") {
		LoadProbility()
		probDataAvailable = IsProbabilityLoaded()
	}

	os.Exit(m.Run())
}

func TestPokeParsing(t *testing.T) {
	b := StrToPoke("方A")
	p := NewPokeFromByte(b)
	if p.Color != PokeColorFang {
		t.Errorf("color = %d, want %d", p.Color, PokeColorFang)
	}
	if p.Value != PokeValueA {
		t.Errorf("value = %d, want %d", p.Value, PokeValueA)
	}
	if p.String() != "方A" {
		t.Errorf("String() = %q, want %q", p.String(), "方A")
	}

	guiByte := StrToPoke("鬼")
	gui := NewPokeFromByte(guiByte)
	if !gui.IsGui() {
		t.Errorf("鬼 should be a wild card")
	}
	if gui.ToByte() != GUI.ToByte() {
		t.Errorf("gui byte = %d, want %d", gui.ToByte(), GUI.ToByte())
	}

	pokes := StrToPokes("方2,梅3,黑2,黑4,鬼")
	if len(pokes) != 5 {
		t.Fatalf("len = %d, want 5", len(pokes))
	}
	if got := PokesToStr(pokes); got != "方2梅3黑2黑4鬼" {
		t.Errorf("PokesToStr = %q, want %q", got, "方2梅3黑2黑4鬼")
	}
}

func TestCardTypes(t *testing.T) {
	// 1. Royal Flush
	checkCardType(t, "黑10,黑J,黑Q,黑K,黑A", TexasCardTypeKingTongHuaShun)

	// 2. Straight Flush
	checkCardType(t, "红9,红10,红J,红Q,红K", TexasCardTypeTongHuaShun)

	// 3. Four of a kind
	checkCardType(t, "黑A,红A,梅A,方A,黑K", TexasCardTypeSiTiao)

	// 4. Full House
	checkCardType(t, "黑A,红A,梅A,黑K,红K", TexasCardTypeHuLu)

	// 5. Flush
	checkCardType(t, "黑2,黑4,黑6,黑8,黑K", TexasCardTypeTongHua)

	// 6. Straight
	checkCardType(t, "黑10,红J,梅Q,方K,黑A", TexasCardTypeShunZi)
	checkCardType(t, "黑A,红2,梅3,方4,黑5", TexasCardTypeShunZi) // A-2-3-4-5

	// 7. Three of a kind
	checkCardType(t, "黑A,红A,梅A,黑K,红Q", TexasCardTypeSanTiao)

	// 8. Two Pair
	checkCardType(t, "黑A,红A,黑K,红K,黑Q", TexasCardTypeLiangDui)

	// 9. One Pair
	checkCardType(t, "黑A,红A,黑K,红Q,黑J", TexasCardTypeDuiZi)

	// 10. High Card
	checkCardType(t, "黑A,红K,黑Q,红J,黑9", TexasCardTypeGaoPai)
}

func TestWildCardEvaluation(t *testing.T) {
	// 4 cards + 1 Gui -> Royal Flush
	checkCardTypeWithGui(t, "黑10,黑J,黑Q,黑K,鬼", TexasCardTypeKingTongHuaShun)

	// 3 cards + 2 Gui -> Straight Flush
	checkCardTypeWithGui(t, "方2,方3,方4,鬼,鬼", TexasCardTypeTongHuaShun)

	// Three of a kind + 1 Gui -> Four of a kind
	checkCardTypeWithGui(t, "黑A,红A,梅A,黑K,鬼", TexasCardTypeSiTiao)

	// One pair + 1 Gui -> Three of a kind
	checkCardTypeWithGui(t, "黑A,红A,黑K,红Q,鬼", TexasCardTypeSanTiao)
}

func TestCardComparison(t *testing.T) {
	royalFlush := toSortedPokeList("黑10,黑J,黑Q,黑K,黑A")
	straightFlush := toSortedPokeList("红9,红10,红J,红Q,红K")
	fourOfAKind := toSortedPokeList("黑A,红A,梅A,方A,黑K")
	fullHouse := toSortedPokeList("黑A,红A,梅A,黑K,红K")

	if CompareCardsWithoutGui(royalFlush, straightFlush) <= 0 {
		t.Errorf("royal flush should beat straight flush")
	}
	if CompareCardsWithoutGui(straightFlush, fourOfAKind) <= 0 {
		t.Errorf("straight flush should beat four of a kind")
	}
	if CompareCardsWithoutGui(fourOfAKind, fullHouse) <= 0 {
		t.Errorf("four of a kind should beat full house")
	}
	if CompareCardsWithoutGui(royalFlush, royalFlush) != 0 {
		t.Errorf("royal flush should tie with itself")
	}
}

func checkCardType(t *testing.T, cardsStr string, expectedType int) {
	t.Helper()
	pokes := toPokeList(cardsStr)
	actualType := GetCardTypeUnorderedWithoutGui(pokes)
	if actualType != expectedType {
		t.Errorf("Failed for cards: %s, got type %d, want %d", cardsStr, actualType, expectedType)
	}
}

func checkCardTypeWithGui(t *testing.T, cardsStr string, expectedType int) {
	t.Helper()
	pokes := toPokeList(cardsStr)
	picked := FiveFromFive(pokes)
	actualType := GetCardTypeUnorderedWithoutGui(picked)
	if actualType != expectedType {
		t.Errorf("Failed with gui for cards: %s, got type %d, want %d", cardsStr, actualType, expectedType)
	}
}

func toPokeList(cardsStr string) []Poke {
	bytes := StrToPokes(cardsStr)
	return ToArray(GenCardBindBytes(bytes))
}

func toSortedPokeList(cardsStr string) []Poke {
	list := toPokeList(cardsStr)
	return FiveFromFiveWithoutGui(list)
}
