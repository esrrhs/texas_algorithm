// Mirror of Java's TexasAlgorithmUtilTest (the Go suite ports the same cases).
// Tests that need the generated data tables skip gracefully when the files
// are absent, exactly like the Java and Go suites.

#include "gen_util.h"
#include "tests/test_util.h"
#include "texas_algorithm_util.h"
#include "texas_card_util.h"

using namespace texas_algorithm;
using namespace txtest;

namespace
{

  bool g_dataAvailable = false;
  bool g_probDataAvailable = false;

  void SetUp()
  {
    if (FileExists("texas_data_color.txt"))
    {
      Load();
      g_dataAvailable = IsLoaded();
    }

    if (FileExists("texas_data_opt_2.txt"))
    {
      LoadProbility();
      g_probDataAvailable = IsProbabilityLoaded();
    }
  }

  bool Contains(const std::string &s, const std::string &sub) { return s.find(sub) != std::string::npos; }

  void TestConversions()
  {
    std::string original = "方A,黑K,红10,梅J,鬼";
    std::vector<int> pokes = StrToPokes(original);
    CheckEq((long long)pokes.size(), 5, "pokes size");

    std::string reconstructed = PokesToStr(pokes);
    CheckEqStr(reconstructed, "方A黑K红10梅J鬼", "PokesToStr round trip");

    int64_t key = GenCardBindBytes(pokes);
    std::vector<int> fromKey = KeyToPoke(key);
    CheckEq((long long)fromKey.size(), 5, "KeyToPoke size");
  }

  void TestGetMaxWithoutGui()
  {
    std::string best = GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6").first;
    Check(!best.empty(), "best is not empty");
    // Best 5 cards should form a straight: 2, 3, 4, 5, 6
    for (const char *want : {"2", "3", "4", "5", "6"})
    {
      Check(Contains(best, want), std::string("best should contain ") + want);
    }
  }

  void TestGetMaxWithGui()
  {
    std::pair<std::string, std::vector<int>> res = GetMaxStr("方2,梅3,黑2,黑4,鬼");
    Check(!res.first.empty(), "best is not empty");
    // Wild card should turn this hand into three of a kind or full house
    Check(!res.second.empty(), "guiTrans should not be empty");

    res = GetMaxStr("方2,方3,方4,鬼,鬼,黑6,红6");
    Check(!res.first.empty(), "best2 is not empty");
    // Two wild cards with 2, 3, 4 of diamonds should transform into straight
    // flush (5, 6 of diamonds)
    CheckEq((long long)res.second.size(), 2, "guiTrans size");
  }

  void TestTableLookup()
  {
    std::string cards = "方4,方A,黑2,黑A,黑3,黑5,黑6";
    std::string cards1 = "红8,方A,方2,黑8,黑3,黑5,黑7";

    int pos = GetWinPositionStr(cards);
    Check(pos > 0, "pos > 0");
    CheckEq(pos, 4010, "pos");

    double prob = GetWinProbabilityStr(cards);
    Check(prob > 0.8 && prob < 1.0, "prob in (0.8, 1.0)");

    int type = GetWinTypeStr(cards);
    CheckEq(type, TexasCardTypeTongHua, "type");

    int pos1 = GetWinPositionStr(cards1);
    CheckEq(pos1, 1143, "pos1");
    int type1 = GetWinTypeStr(cards1);
    CheckEq(type1, TexasCardTypeDuiZi, "type1");

    int cmp = CompareStr(cards, cards1);
    Check(cmp > 0, "cards should beat cards1");
  }

  void TestProbabilityEstimation()
  {
    float p1 = GetHandProbabilityStr("方3,方A", "黑2,黑4,黑5,黑K");
    Check(p1 > 0.6f && p1 < 0.9f, "p1 should be around 0.74");

    float p2 = GetHandProbabilityStr("方2,方3", "");
    Check(p2 > 0.3f && p2 < 0.6f, "p2 should be around 0.45");
  }

  void TestUnloadedProbabilityGraceful()
  {
    // When querying an invalid or unloaded card key
    const ProbilityData *data = GetHandProbabilityKey(1);
    Check(data == nullptr, "GetHandProbabilityKey(1) should be null");
  }

} // namespace

int main()
{
  SetUp();

  RunTest("TestConversions", true, TestConversions);
  RunTest("TestGetMaxWithoutGui", g_dataAvailable, TestGetMaxWithoutGui);
  RunTest("TestGetMaxWithGui", g_dataAvailable, TestGetMaxWithGui);
  RunTest("TestTableLookup", g_dataAvailable, TestTableLookup);
  RunTest("TestProbabilityEstimation", g_probDataAvailable, TestProbabilityEstimation);
  RunTest("TestUnloadedProbabilityGraceful", true, TestUnloadedProbabilityGraceful);
  return ExitCode();
}
