// Mirror of Java's TexasCardUtilTest (the Go suite ports the same cases).

#include <vector>

#include "gen_util.h"
#include "tests/test_util.h"
#include "texas_algorithm_util.h"
#include "texas_card_util.h"

using namespace texas_algorithm;
using namespace txtest;

namespace
{

  std::vector<Poke> ToPokeList(const std::string &cardsStr)
  {
    std::vector<int> bytes = StrToPokes(cardsStr);
    return ToArray(GenCardBindBytes(bytes));
  }

  std::vector<Poke> ToSortedPokeList(const std::string &cardsStr)
  {
    std::vector<Poke> list = ToPokeList(cardsStr);
    return FiveFromFiveWithoutGui(list);
  }

  void CheckCardType(const std::string &cardsStr, int expectedType)
  {
    std::vector<Poke> pokes = ToPokeList(cardsStr);
    int actualType = GetCardTypeUnorderedWithoutGui(pokes);
    CheckEq(actualType, expectedType, "Failed for cards: " + cardsStr);
  }

  void CheckCardTypeWithGui(const std::string &cardsStr, int expectedType)
  {
    std::vector<Poke> pokes = ToPokeList(cardsStr);
    std::vector<Poke> picked = FiveFromFive(pokes);
    int actualType = GetCardTypeUnorderedWithoutGui(picked);
    CheckEq(actualType, expectedType, "Failed with gui for cards: " + cardsStr);
  }

  void TestPokeParsing()
  {
    int b = StrToPoke("方A");
    Poke p = PokeFromByte(b);
    CheckEq(p.color, PokeColorFang, "color");
    CheckEq(p.value, PokeValueA, "value");
    CheckEqStr(PokeToString(p), "方A", "PokeToString");

    int guiByte = StrToPoke("鬼");
    Poke gui = PokeFromByte(guiByte);
    Check(IsGui(gui), "IsGui(gui)");
    CheckEq(PokeToByte(gui), PokeToByte(GUI), "gui byte");

    std::vector<int> pokes = StrToPokes("方2,梅3,黑2,黑4,鬼");
    CheckEq((long long)pokes.size(), 5, "pokes size");
    CheckEqStr(PokesToStr(pokes), "方2梅3黑2黑4鬼", "PokesToStr");
  }

  void TestCardTypes()
  {
    // 1. Royal Flush
    CheckCardType("黑10,黑J,黑Q,黑K,黑A", TexasCardTypeKingTongHuaShun);

    // 2. Straight Flush
    CheckCardType("红9,红10,红J,红Q,红K", TexasCardTypeTongHuaShun);

    // 3. Four of a kind
    CheckCardType("黑A,红A,梅A,方A,黑K", TexasCardTypeSiTiao);

    // 4. Full House
    CheckCardType("黑A,红A,梅A,黑K,红K", TexasCardTypeHuLu);

    // 5. Flush
    CheckCardType("黑2,黑4,黑6,黑8,黑K", TexasCardTypeTongHua);

    // 6. Straight
    CheckCardType("黑10,红J,梅Q,方K,黑A", TexasCardTypeShunZi);
    CheckCardType("黑A,红2,梅3,方4,黑5", TexasCardTypeShunZi); // A-2-3-4-5

    // 7. Three of a kind
    CheckCardType("黑A,红A,梅A,黑K,红Q", TexasCardTypeSanTiao);

    // 8. Two Pair
    CheckCardType("黑A,红A,黑K,红K,黑Q", TexasCardTypeLiangDui);

    // 9. One Pair
    CheckCardType("黑A,红A,黑K,红Q,黑J", TexasCardTypeDuiZi);

    // 10. High Card
    CheckCardType("黑A,红K,黑Q,红J,黑9", TexasCardTypeGaoPai);
  }

  void TestWildCardEvaluation()
  {
    // 4 cards + 1 Gui -> Royal Flush
    CheckCardTypeWithGui("黑10,黑J,黑Q,黑K,鬼", TexasCardTypeKingTongHuaShun);

    // 3 cards + 2 Gui -> Straight Flush
    CheckCardTypeWithGui("方2,方3,方4,鬼,鬼", TexasCardTypeTongHuaShun);

    // Three of a kind + 1 Gui -> Four of a kind
    CheckCardTypeWithGui("黑A,红A,梅A,黑K,鬼", TexasCardTypeSiTiao);

    // One pair + 1 Gui -> Three of a kind
    CheckCardTypeWithGui("黑A,红A,黑K,红Q,鬼", TexasCardTypeSanTiao);
  }

  void TestCardComparison()
  {
    std::vector<Poke> royalFlush = ToSortedPokeList("黑10,黑J,黑Q,黑K,黑A");
    std::vector<Poke> straightFlush = ToSortedPokeList("红9,红10,红J,红Q,红K");
    std::vector<Poke> fourOfAKind = ToSortedPokeList("黑A,红A,梅A,方A,黑K");
    std::vector<Poke> fullHouse = ToSortedPokeList("黑A,红A,梅A,黑K,红K");

    Check(CompareCardsWithoutGui(royalFlush, straightFlush) > 0, "royal flush should beat straight flush");
    Check(CompareCardsWithoutGui(straightFlush, fourOfAKind) > 0, "straight flush should beat four of a kind");
    Check(CompareCardsWithoutGui(fourOfAKind, fullHouse) > 0, "four of a kind should beat full house");
    CheckEq(CompareCardsWithoutGui(royalFlush, royalFlush), 0, "royal flush should tie with itself");
  }

} // namespace

int main()
{
  RunTest("TestPokeParsing", true, TestPokeParsing);
  RunTest("TestCardTypes", true, TestCardTypes);
  RunTest("TestWildCardEvaluation", true, TestWildCardEvaluation);
  RunTest("TestCardComparison", true, TestCardComparison);
  return ExitCode();
}
