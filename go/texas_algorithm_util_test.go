package texas_algorithm

import (
	"strings"
	"testing"
)

func TestConversions(t *testing.T) {
	original := "方A,黑K,红10,梅J,鬼"
	pokes := StrToPokes(original)
	if len(pokes) != 5 {
		t.Fatalf("len = %d, want 5", len(pokes))
	}

	reconstructed := PokesToStr(pokes)
	if reconstructed != "方A黑K红10梅J鬼" {
		t.Errorf("PokesToStr = %q, want %q", reconstructed, "方A黑K红10梅J鬼")
	}

	key := GenCardBindBytes(pokes)
	fromKey := KeyToPoke(key)
	if len(fromKey) != 5 {
		t.Fatalf("len = %d, want 5", len(fromKey))
	}
}

func TestGetMaxWithoutGui(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping testGetMaxWithoutGui: data files not present")
	}
	best, _ := GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6")
	if best == "" {
		t.Fatalf("best is empty")
	}
	// Best 5 cards should form a straight: 2, 3, 4, 5, 6
	for _, want := range []string{"2", "3", "4", "5", "6"} {
		if !strings.Contains(best, want) {
			t.Errorf("best = %q, want it to contain %q", best, want)
		}
	}
}

func TestGetMaxWithGui(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping testGetMaxWithGui: data files not present")
	}
	best, guiTrans := GetMaxStr("方2,梅3,黑2,黑4,鬼")
	if best == "" {
		t.Fatalf("best is empty")
	}
	// Wild card should turn this hand into three of a kind or full house
	if len(guiTrans) == 0 {
		t.Errorf("guiTrans should not be empty")
	}

	best, guiTrans = GetMaxStr("方2,方3,方4,鬼,鬼,黑6,红6")
	if best == "" {
		t.Fatalf("best is empty")
	}
	// Two wild cards with 2, 3, 4 of diamonds should transform into straight
	// flush (5, 6 of diamonds)
	if len(guiTrans) != 2 {
		t.Errorf("guiTrans len = %d, want 2", len(guiTrans))
	}
}

func TestTableLookup(t *testing.T) {
	if !dataAvailable {
		t.Skip("Skipping testTableLookup: data files not present")
	}

	cards := "方4,方A,黑2,黑A,黑3,黑5,黑6"
	cards1 := "红8,方A,方2,黑8,黑3,黑5,黑7"

	pos := GetWinPositionStr(cards)
	if pos <= 0 {
		t.Errorf("pos = %d, want > 0", pos)
	}
	if pos != 4010 {
		t.Errorf("pos = %d, want 4010", pos)
	}

	prob := GetWinProbabilityStr(cards)
	if prob <= 0.8 || prob >= 1.0 {
		t.Errorf("prob = %f, want in (0.8, 1.0)", prob)
	}

	typ := GetWinTypeStr(cards)
	if typ != TexasCardTypeTongHua {
		t.Errorf("type = %d, want %d", typ, TexasCardTypeTongHua)
	}

	pos1 := GetWinPositionStr(cards1)
	if pos1 != 1143 {
		t.Errorf("pos1 = %d, want 1143", pos1)
	}
	typ1 := GetWinTypeStr(cards1)
	if typ1 != TexasCardTypeDuiZi {
		t.Errorf("type1 = %d, want %d", typ1, TexasCardTypeDuiZi)
	}

	cmp := CompareStr(cards, cards1)
	if cmp <= 0 {
		t.Errorf("cards should beat cards1, cmp = %d", cmp)
	}
}

func TestProbabilityEstimation(t *testing.T) {
	if !probDataAvailable {
		t.Skip("Skipping testProbabilityEstimation: prob data files not present")
	}

	p1 := GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K")
	if p1 <= 0.6 || p1 >= 0.9 {
		t.Errorf("p1 should be around 0.74, got: %f", p1)
	}

	p2 := GetHandProbabilityStr("方2,方3", "")
	if p2 <= 0.3 || p2 >= 0.6 {
		t.Errorf("p2 should be around 0.45, got: %f", p2)
	}
}

func TestUnloadedProbabilityGraceful(t *testing.T) {
	// When querying an invalid or unloaded card key
	data := GetHandProbabilityKey(1)
	if data != nil {
		t.Errorf("GetHandProbabilityKey(1) should be nil, got %+v", data)
	}
}
