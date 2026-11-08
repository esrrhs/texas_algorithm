package texas_algorithm

import (
	"errors"
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

	if fileExists("texas_data_winrate_0.txt") {
		LoadProbility()
		probDataAvailable = IsProbabilityLoaded()
	}

	os.Exit(m.Run())
}

func TestPokeParsing(t *testing.T) {
	b, err := StrToPoke("方A")
	if err != nil {
		t.Fatalf("StrToPoke(方A) error: %v", err)
	}
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

	pokes, err := StrToPokes("方2,梅3,黑2,黑4,黑5")
	if err != nil {
		t.Fatalf("StrToPokes error: %v", err)
	}
	if len(pokes) != 5 {
		t.Fatalf("len = %d, want 5", len(pokes))
	}
	if got := PokesToStr(pokes); got != "方2梅3黑2黑4黑5" {
		t.Errorf("PokesToStr = %q, want %q", got, "方2梅3黑2黑4黑5")
	}
}

// TestJokerRejected pins the new contract: the joker card was removed from the
// deck, so every entry point must report ErrUnknownPoke instead of silently
// returning a zero value or nil.
func TestJokerRejected(t *testing.T) {
	if _, err := StrToPoke("鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("StrToPoke(鬼) err = %v, want ErrUnknownPoke", err)
	}
	if _, err := StrToPokes("方2,梅3,黑2,黑4,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("StrToPokes with 5 cards and 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := StrToPokes("方2,梅3,黑2,黑4,黑5,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("StrToPokes with 6 cards and 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := StrToPokes("方2,梅3,黑2,黑4,黑5,黑6,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("StrToPokes with 7 cards and 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetMaxStr("方2,梅3,黑2,黑4,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetMaxStr with 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetWinPositionStr("方2,梅3,黑2,黑4,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetWinPositionStr with 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetWinTypeStr("方2,梅3,黑2,黑4,鬼"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetWinTypeStr with 鬼 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := CompareStr("方2,梅3,黑2,黑4,鬼", "方2,梅3,黑2,黑4,黑5"); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("CompareStr with 鬼 err = %v, want ErrUnknownPoke", err)
	}

	// The joker byte 0x58 = 88 belongs to color 5, which the 52-card deck does
	// not have; both byte-slice and key entry points must reject it.
	jokerBytes := []byte{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x58}
	if err := CheckPokes(jokerBytes); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("CheckPokes(88) err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetMax(jokerBytes); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetMax with byte 88 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetWinPosition(jokerBytes); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetWinPosition with byte 88 err = %v, want ErrUnknownPoke", err)
	}
	if _, err := GetMaxHandPub([]byte{0x12, 0x58}, []byte{0x13, 0x24, 0x25, 0x36}); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetMaxHandPub with joker in hand err = %v, want ErrUnknownPoke", err)
	}
	if _, err := CompareBytes(jokerBytes, []byte{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x48}); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("CompareBytes with byte 88 err = %v, want ErrUnknownPoke", err)
	}

	// A hand key containing a joker card, e.g. 方2 梅3 黑2 黑4 黑5 黑6 鬼,
	// is rejected by the key based entry points too.
	jokerKey := GenCardBindBytes(jokerBytes)
	if err := CheckKey(jokerKey); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("CheckKey(%d) err = %v, want ErrUnknownPoke", jokerKey, err)
	}
	if _, err := GetKeyDataKey(jokerKey); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("GetKeyDataKey(%d) err = %v, want ErrUnknownPoke", jokerKey, err)
	}
	if _, err := CompareKey(jokerKey, GenCardBindBytes([]byte{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x48})); !errors.Is(err, ErrUnknownPoke) {
		t.Errorf("CompareKey with joker key err = %v, want ErrUnknownPoke", err)
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

func TestCardComparison(t *testing.T) {
	royalFlush := toSortedPokeList(t, "黑10,黑J,黑Q,黑K,黑A")
	straightFlush := toSortedPokeList(t, "红9,红10,红J,红Q,红K")
	fourOfAKind := toSortedPokeList(t, "黑A,红A,梅A,方A,黑K")
	fullHouse := toSortedPokeList(t, "黑A,红A,梅A,黑K,红K")

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
	pokes := toPokeList(t, cardsStr)
	actualType := GetCardTypeUnorderedWithoutGui(pokes)
	if actualType != expectedType {
		t.Errorf("Failed for cards: %s, got type %d, want %d", cardsStr, actualType, expectedType)
	}
}

func toPokeList(t *testing.T, cardsStr string) []Poke {
	t.Helper()
	bytes, err := StrToPokes(cardsStr)
	if err != nil {
		t.Fatalf("StrToPokes(%q) error: %v", cardsStr, err)
	}
	return ToArray(GenCardBindBytes(bytes))
}

func toSortedPokeList(t *testing.T, cardsStr string) []Poke {
	t.Helper()
	list := toPokeList(t, cardsStr)
	return FiveFromFiveWithoutGui(list)
}
