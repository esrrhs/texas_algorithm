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

    if (FileExists("texas_data_winrate_0.txt"))
    {
      LoadProbility();
      g_probDataAvailable = IsProbabilityLoaded();
    }
  }

  bool Contains(const std::string &s, const std::string &sub) { return s.find(sub) != std::string::npos; }

  void TestConversions()
  {
    std::string original = "方A,黑K,红10,梅J,黑9";
    std::vector<int> pokes = StrToPokes(original);
    CheckEq((long long)pokes.size(), 5, "pokes size");

    std::string reconstructed = PokesToStr(pokes);
    CheckEqStr(reconstructed, "方A黑K红10梅J黑9", "PokesToStr round trip");

    int64_t key = GenCardBindBytes(pokes);
    std::vector<int> fromKey = KeyToPoke(key);
    CheckEq((long long)fromKey.size(), 5, "KeyToPoke size");

    // The deck is the standard 52 cards, so the 7-card key set is C(52, 7).
    CheckEq((long long)Total, 133784560LL, "Total should be C(52, 7)");
  }

  void TestGetMax()
  {
    std::string best = GetMaxStrHandPub("方4,方2", "黑2,黑A,方3,黑5,黑6");
    Check(!best.empty(), "best is not empty");
    // Best 5 cards should form a straight: 2, 3, 4, 5, 6
    for (const char *want : {"2", "3", "4", "5", "6"})
    {
      Check(Contains(best, want), std::string("best should contain ") + want);
    }
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
    Check(p1 > 0.55f && p1 < 0.56f, "p1 should be around 0.5533 (exact)");

    float p2 = GetHandProbabilityStr("方2,方3", "");
    Check(p2 > 0.35f && p2 < 0.37f, "p2 should be around 0.3598 (exact)");
  }

  void TestUnloadedProbabilityGraceful()
  {
    // A wrong number of hole cards must be reported as an exception, never a crash.
    bool threw = false;
    try
    {
      GetHandProbability({0x02}, {});
    }
    catch (const std::invalid_argument &)
    {
      threw = true;
    }
    Check(threw, "GetHandProbability with one hole card should throw");
  }

  void TestMissingWinRateTableGraceful()
  {
    // Missing exact tables must throw with a clear message while N >= 3 keeps
    // working, mirroring the Go and Java suites.
    std::unordered_map<int64_t, WinRateData> saved[3];
    for (int i = 0; i < 3; i++)
    {
      saved[i] = WinRateMap[i];
      WinRateMap[i].clear();
    }

    bool threw = false;
    try
    {
      GetHandProbability({0x02, 0x03}, {});
    }
    catch (const std::invalid_argument &)
    {
      threw = true;
    }
    Check(threw, "preflop without winrate_0 should throw");

    float p = GetHandProbability({0x02, 0x03}, {0x04, 0x05, 0x06});
    Check(p > 0 && p < 1, "flop should be enumerated without tables");

    for (int i = 0; i < 3; i++)
    {
      WinRateMap[i] = saved[i];
    }
  }

} // namespace

int main()
{
  SetUp();

  RunTest("TestConversions", true, TestConversions);
  RunTest("TestGetMax", g_dataAvailable, TestGetMax);
  RunTest("TestTableLookup", g_dataAvailable, TestTableLookup);
  RunTest("TestProbabilityEstimation", g_probDataAvailable, TestProbabilityEstimation);
  RunTest("TestUnloadedProbabilityGraceful", true, TestUnloadedProbabilityGraceful);
  RunTest("TestMissingWinRateTableGraceful", g_dataAvailable, TestMissingWinRateTableGraceful);
  return ExitCode();
}
