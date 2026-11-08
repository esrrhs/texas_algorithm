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
  bool g_sqliteSupported = false;
  bool g_sqliteAvailable = false;

  void SetUp()
  {
    if (FileExists("texas_data_color.txt"))
    {
      Load();
      g_dataAvailable = IsLoaded();
    }

    if (FileExists("texas_data_winrate_0.zst"))
    {
      LoadProbility();
      g_probDataAvailable = IsProbabilityLoaded();
    }

    // The packed database is not committed (about 2.3 GiB) and has to be
    // generated locally with go/cmd/packsqlite, so the SQLite test skips when
    // it is absent. A build configured with -DTEXAS_SQLITE=off also skips it:
    // LoadSqlite is then a stub that reports the missing feature.
    g_sqliteSupported = IsSqliteSupported();
    g_sqliteAvailable = g_sqliteSupported && FileExists("texas_data_winrate.db");
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

  // TestWinRate3TableMatchesEnumeration checks that the n=3 table agrees with
  // the on-the-fly enumeration. It is the correctness gate for the N <= 3
  // lookup path in this language.
  void TestWinRate3TableMatchesEnumeration()
  {
    float fromTable = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");

    WinRateTable saved = std::move(WinRateMap[3]);
    WinRateMap[3] = WinRateTable();
    float fromEnum = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");
    WinRateMap[3] = std::move(saved);

    Check(fromTable == fromEnum, "winrate_3 table should match the enumeration");
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
    // Without an in-memory table the lookup falls back to enumeration instead of
    // throwing, mirroring the Go and Java suites. Level 3 is used because its
    // enumeration is ~1M evaluations; the preflop one would take minutes.
    WinRateTable saved[6];
    for (int i = 0; i < 6; i++)
    {
      saved[i] = std::move(WinRateMap[i]);
      WinRateMap[i] = WinRateTable();
    }

    float p = GetHandProbability({0x02, 0x03}, {0x04, 0x05, 0x06});
    Check(p > 0 && p < 1, "flop should be enumerated without tables");

    float q = GetHandProbability({0x02, 0x03}, {0x04, 0x05, 0x06, 0x07});
    Check(q > 0 && q < 1, "turn should be enumerated without tables");

    for (int i = 0; i < 6; i++)
    {
      WinRateMap[i] = std::move(saved[i]);
    }
  }

  // TestSqliteMatchesEnumeration checks the on-demand database path: a row read
  // from the packed SQLite file must agree with the in-memory table, a level
  // outside the loaded range must fall back to enumeration, and a missing file
  // must be rejected instead of crashing.
  void TestSqliteMatchesEnumeration()
  {
    Check(LoadSqlite("texas_data_winrate.db", 3, 3), "open the packed database for level 3");

    WinRateTable saved = std::move(WinRateMap[3]);
    WinRateMap[3] = WinRateTable();
    float fromDb = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");
    WinRateMap[3] = std::move(saved);
    float fromTable = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");
    Check(fromDb == fromTable, "SQLite row should match the in-memory table");

    // Range limiting: level 3 is outside 0..2, so it must enumerate; the answer
    // is still exact and must equal the table value.
    Check(LoadSqlite("texas_data_winrate.db", 0, 2), "open the packed database for levels 0..2");
    WinRateTable saved3 = std::move(WinRateMap[3]);
    WinRateMap[3] = WinRateTable();
    float fromEnum = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");
    WinRateMap[3] = std::move(saved3);
    Check(fromEnum == fromTable, "a level outside the SQLite range should enumerate to the same value");

    Check(!LoadSqlite("texas_data_winrate_missing.db", 0, 5), "a missing database should be rejected");
  }

  // TestSqliteUnsupportedBuild checks the build configured without SQLite
  // support: LoadSqlite still links, reports the missing feature and returns
  // false, and the lookup keeps working through enumeration.
  void TestSqliteUnsupportedBuild()
  {
    Check(!LoadSqlite("texas_data_winrate.db", 0, 5),
          "LoadSqlite should fail in a build without SQLite support");

    float p = GetHandProbabilityStr("方3,梅10", "黑2,红4,方7");
    Check(p > 0 && p < 1, "enumeration should still answer without SQLite support");
  }

} // namespace

int main()
{
  SetUp();

  RunTest("TestConversions", true, TestConversions);
  RunTest("TestGetMax", g_dataAvailable, TestGetMax);
  RunTest("TestTableLookup", g_dataAvailable, TestTableLookup);
  RunTest("TestProbabilityEstimation", g_probDataAvailable, TestProbabilityEstimation);
  RunTest("TestWinRate3TableMatchesEnumeration", g_probDataAvailable, TestWinRate3TableMatchesEnumeration);
  RunTest("TestUnloadedProbabilityGraceful", true, TestUnloadedProbabilityGraceful);
  RunTest("TestMissingWinRateTableGraceful", g_dataAvailable, TestMissingWinRateTableGraceful);
  RunTest("TestSqliteMatchesEnumeration", g_sqliteAvailable, TestSqliteMatchesEnumeration);
  RunTest("TestSqliteUnsupportedBuild", !g_sqliteSupported, TestSqliteUnsupportedBuild);
  return ExitCode();
}
