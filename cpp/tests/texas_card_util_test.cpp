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

  void TestPokeParsing()
  {
    int b = StrToPoke("方A");
    Poke p = PokeFromByte(b);
    CheckEq(p.color, PokeColorFang, "color");
    CheckEq(p.value, PokeValueA, "value");
    CheckEqStr(PokeToString(p), "方A", "PokeToString");

    std::vector<int> pokes = StrToPokes("方2,梅3,黑2,黑4,黑5");
    CheckEq((long long)pokes.size(), 5, "pokes size");
    CheckEqStr(PokesToStr(pokes), "方2梅3黑2黑4黑5", "PokesToStr");
  }

  // TestJokerRejected pins the new contract: the joker card was removed from
  // the deck, so every entry point must throw std::invalid_argument instead of
  // silently returning a zero value, nullptr or false.
  void TestJokerRejected()
  {
    CheckThrowsInvalidArgument([]()
                               { StrToPoke("鬼"); },
                               "牌面 \"鬼\"", "StrToPoke(鬼)");
    CheckThrowsInvalidArgument([]()
                               { StrToPokes("方2,梅3,黑2,黑4,鬼"); },
                               "牌面 \"鬼\"", "StrToPokes with 5 cards and 鬼");
    CheckThrowsInvalidArgument([]()
                               { StrToPokes("方2,梅3,黑2,黑4,黑5,鬼"); },
                               "牌面 \"鬼\"", "StrToPokes with 6 cards and 鬼");
    CheckThrowsInvalidArgument([]()
                               { StrToPokes("方2,梅3,黑2,黑4,黑5,黑6,鬼"); },
                               "牌面 \"鬼\"", "StrToPokes with 7 cards and 鬼");
    CheckThrowsInvalidArgument([]()
                               { GetMaxStr("方2,梅3,黑2,黑4,鬼"); },
                               "牌面 \"鬼\"", "GetMaxStr with 鬼");
    CheckThrowsInvalidArgument([]()
                               { GetMaxStrHandPub("方2,鬼", "黑2,黑4,黑5"); },
                               "牌面 \"鬼\"", "GetMaxStrHandPub with 鬼");
    CheckThrowsInvalidArgument([]()
                               { GetWinPositionStr("方2,梅3,黑2,黑4,鬼"); },
                               "牌面 \"鬼\"", "GetWinPositionStr with 鬼");
    CheckThrowsInvalidArgument([]()
                               { GetWinTypeStr("方2,梅3,黑2,黑4,鬼"); },
                               "牌面 \"鬼\"", "GetWinTypeStr with 鬼");
    CheckThrowsInvalidArgument([]()
                               { CompareStr("方2,梅3,黑2,黑4,鬼", "方2,梅3,黑2,黑4,黑5"); },
                               "牌面 \"鬼\"", "CompareStr with 鬼");

    // The joker byte 0x58 = 88 belongs to color 5, which the 52-card deck does
    // not have; both byte-slice and hand-key entry points must reject it.
    std::vector<int> jokerBytes{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x58};
    CheckThrowsInvalidArgument([&]()
                               { CheckPokes(jokerBytes); },
                               "牌字节 88", "CheckPokes with byte 88");
    CheckThrowsInvalidArgument([&]()
                               { GetMax(jokerBytes); },
                               "牌字节 88", "GetMax with byte 88");
    CheckThrowsInvalidArgument([&]()
                               { GetWinPosition(jokerBytes); },
                               "牌字节 88", "GetWinPosition with byte 88");
    CheckThrowsInvalidArgument([]()
                               { GetMaxHandPub({0x12, 0x58}, {0x13, 0x24, 0x25, 0x36}); },
                               "牌字节 88", "GetMaxHandPub with joker in hand");
    CheckThrowsInvalidArgument([&]()
                               { CompareBytes(jokerBytes, std::vector<int>{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x48}); },
                               "牌字节 88", "CompareBytes with byte 88");

    // A hand key containing a joker card, e.g. 方2 梅3 黑2 黑4 黑5 黑6 鬼, is
    // rejected by the key based entry points too.
    int64_t jokerKey = GenCardBindBytes(jokerBytes);
    CheckThrowsInvalidArgument([&]()
                               { CheckKey(jokerKey); },
                               std::to_string(jokerKey), "CheckKey with joker key");
    CheckThrowsInvalidArgument([&]()
                               { GetKeyDataKey(jokerKey); },
                               std::to_string(jokerKey), "GetKeyDataKey with joker key");
    CheckThrowsInvalidArgument([&]()
                               { CompareKey(jokerKey, GenCardBindBytes(std::vector<int>{0x12, 0x13, 0x24, 0x25, 0x36, 0x37, 0x48})); },
                               std::to_string(jokerKey), "CompareKey with joker key");
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
  RunTest("TestJokerRejected", true, TestJokerRejected);
  RunTest("TestCardTypes", true, TestCardTypes);
  RunTest("TestCardComparison", true, TestCardComparison);
  return ExitCode();
}
