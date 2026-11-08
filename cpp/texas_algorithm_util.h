#pragma once

#include <cstdint>
#include <string>
#include <unordered_map>
#include <utility>
#include <vector>

namespace texas_algorithm
{

  // KeyData is the lookup result of a 7-card hand key. postion keeps Java's
  // spelling of the original field (getPostion).
  struct KeyData
  {
    int index = 0;
    int postion = 0;
    int64_t max = 0;
    int type = 0;
  };

  // WinRateTable is one exact-count table: keys sorted ascending, with the
  // constant per-situation enumeration total held once in the header. Rows
  // store only win and tie because total is shared by every row.
  class WinRateTable
  {
  public:
    std::vector<int64_t> keys;
    std::vector<uint32_t> win;
    std::vector<uint32_t> tie;
    uint32_t total = 0;

    size_t size() const { return keys.size(); }

    // Find returns the row of a canonical key, or false when absent.
    bool Find(int64_t key, uint32_t &winOut, uint32_t &tieOut, uint32_t &totalOut) const;
  };

  extern std::unordered_map<int64_t, KeyData> ColorMap;
  extern std::unordered_map<int64_t, KeyData> NormalMap;

  // WinRateMap[n] holds the exact counts for n known public cards. Entries 0..3
  // are loaded by default. Levels 4 and 5 ship with the repository as zstd
  // files but stay out of memory unless LoadProbilityAll or LoadWinRateRange
  // asks for them: decoded they need about 223 MiB and 1.97 GiB, while the
  // enumeration answers a level 4 query in ~2 ms and a level 5 query in ~0.2 ms.
  extern WinRateTable WinRateMap[6];

  // LoadWinRateFile loads one exact win-rate table from its zstd-compressed
  // runtime file, decompressed transparently while reading.
  bool LoadWinRateFile(int n, const std::string &path);

  // LoadWinRateRange loads levels from..to (inclusive) of the exact win-rate
  // tables into memory, from the directory last passed to LoadProbilityDir
  // (the current directory by default). LoadWinRateRange(0, 5) is the "all in
  // memory" mode. A level that fails to load leaves the others untouched.
  void LoadWinRateRange(int from, int to);

  // LoadWinRateRangeDir is LoadWinRateRange with an explicit directory.
  void LoadWinRateRangeDir(const std::string &dir, int from, int to);

  // LoadSqlite opens the packed win-rate database produced by cmd/packsqlite
  // and answers levels from..to from it. Rows stay on disk: each query is a
  // primary-key descent of a WITHOUT ROWID table, so only the page cache is
  // resident. Levels outside from..to are ignored. A missing table inside the
  // range simply falls back to enumeration. Returns false when the database
  // cannot be opened or no table could be prepared.
  //
  // SQLite support is a build option (CMake TEXAS_SQLITE, default auto). A
  // build compiled without it still links: LoadSqlite then prints an
  // explanatory message and returns false, and every level falls back to
  // enumeration. IsSqliteSupported() reports which build this is.
  //
  // The in-memory table always wins over the database, so the two can
  // coexist: LoadWinRateRange(0, 2) plus LoadSqlite(db, 0, 5) reads 0-2 from
  // memory and 3-5 from the database.
  bool LoadSqlite(const std::string &path, int from, int to);

  // IsSqliteSupported reports whether this build includes SQLite support.
  bool IsSqliteSupported();

  // LoadProbilityAll loads all six levels, including the 4 and 5 tables that
  // the default startup deliberately leaves out.
  void LoadProbilityAll(const std::string &dir);

  // IsWinRateLoaded reports whether the exact win-rate tables are loaded.
  bool IsWinRateLoaded();

  // Gen pipeline entry points, mirroring the private helpers of Java's
  // TexasAlgorithmUtil.main.
  void Gen();        // generate texas_data.txt if missing
  void GenExtra();   // generate texas_data_extra_6/5.txt if missing
  void GenOpt();     // generate texas_data_color/normal.txt if missing
  void GenExtraOpt(); // generate the extra optimized tables if missing
  void GenTrans();   // generate texas_data_N.txt (N = 6..2)
  void GenTransOpt(); // generate texas_data_opt_N.txt (N = 6..2)
  void GenHand();    // generate the exhaustive 1v1 win-rate files for 0-4 pub cards

  // IsLoaded reports whether the lookup tables are loaded.
  bool IsLoaded();

  // IsProbabilityLoaded reports whether the probability tables are loaded.
  bool IsProbabilityLoaded();

  // Load loads the lookup tables from the current directory.
  void Load();

  // LoadDir loads the lookup tables from the given directory. Like the Java
  // version, loading stops at the first missing file.
  void LoadDir(const std::string &dir);

  // LoadProbility loads the probability tables from the current directory.
  void LoadProbility();

  // LoadProbilityDir loads the probability tables from the given directory.
  void LoadProbilityDir(const std::string &dir);

  // CheckPoke validates one encoded card byte. Only the standard 52-card deck
  // is legal: color 0..3 and value 2..14; the joker byte 88 is rejected.
  // Throws std::invalid_argument whose message starts with "无法识别的牌".
  void CheckPoke(int poke);

  // CheckPokes validates every card byte in the slice.
  void CheckPokes(const std::vector<int> &pokes);

  // CheckKey validates the cards encoded in a hand key.
  void CheckKey(int64_t key);

  // StrToPokeValue converts a rank string ("2".."10", "J", "Q", "K", "A") into
  // its byte value. Throws std::invalid_argument on unknown input, mirroring
  // Java's Byte.parseByte.
  int StrToPokeValue(const std::string &str);

  // StrToPoke converts a card string ("方A", "黑10") into its byte. Throws
  // std::invalid_argument for an empty string, an unknown suit or rank, and
  // the removed joker card ("鬼").
  int StrToPoke(const std::string &str);

  // KeyToStr renders a key as its concatenated card strings.
  std::string KeyToStr(int64_t key);

  // KeyToPoke decodes a key into its card bytes, most significant card first.
  std::vector<int> KeyToPoke(int64_t k);

  // StrToPokes converts a comma-separated card string into card bytes.
  std::vector<int> StrToPokes(const std::string &str);

  // PokesToStr renders card bytes as their concatenated card strings.
  std::string PokesToStr(const std::vector<int> &pokes);

  // GetKeyDataStr looks up a 7-card comma-separated string. Throws
  // std::invalid_argument when a card cannot be recognised or the string is
  // not exactly 7 cards.
  const KeyData *GetKeyDataStr(const std::string &str);

  // GetKeyData looks up the best-hand data of the given cards. Throws
  // std::invalid_argument when a card is illegal or no data is found.
  const KeyData *GetKeyData(const std::vector<int> &pokes);

  // GetKeyDataKey looks up the best-hand data of the given 7-card key. The key
  // is checked against the 52-card deck first, so a key holding the removed
  // joker byte 88 throws std::invalid_argument.
  const KeyData *GetKeyDataKey(int64_t key);

  // LookupKeyData resolves a key without validating its cards, the unvalidated
  // hot path used by the table generator. Returns nullptr when the key is
  // absent from the loaded tables.
  const KeyData *LookupKeyData(int64_t key);

  // CompareKeyLookup compares two keys without validating them. Returns a
  // positive value when k1 is stronger, 0 when equal, negative when weaker.
  int CompareKeyLookup(int64_t k1, int64_t k2);

  // GetMaxStr returns the best 5-card hand of a comma-separated card list
  // (5-7 cards). Throws std::invalid_argument for illegal or wrongly sized
  // input.
  std::string GetMaxStr(const std::string &str);

  // GetMaxStrHandPub returns the best 5-card hand of 2 hole cards and 3-5
  // community cards. Throws std::invalid_argument for illegal or wrongly sized
  // input.
  std::string GetMaxStrHandPub(const std::string &hand, const std::string &pub);

  // GetMax returns the best 5-card hand of the given cards (2 hole cards
  // followed by 3-5 community cards). Throws std::invalid_argument for illegal
  // or wrongly sized input.
  std::vector<int> GetMax(const std::vector<int> &pokes);

  // GetMaxHandPub returns the best 5-card hand of 2 hole cards and 3-5
  // community cards. It maps the stored best-hand key back onto the actual
  // cards that were dealt. Throws std::invalid_argument for illegal or wrongly
  // sized input.
  std::vector<int> GetMaxHandPub(const std::vector<int> &hand, const std::vector<int> &pub);

  // GetWinPosition returns the absolute rank of the given cards among all
  // combinations of the same size.
  int GetWinPosition(const std::vector<int> &pokes);
  int GetWinPositionStr(const std::string &str);

  // GetWinProbability returns the rank ratio of the given cards among all
  // combinations of the same size.
  double GetWinProbability(const std::vector<int> &pokes);
  double GetWinProbabilityStr(const std::string &str);

  // GetWinMax returns the encoded key of the best 5-card hand.
  int64_t GetWinMax(const std::vector<int> &pokes);
  int64_t GetWinMaxStr(const std::string &str);

  // GetWinType returns the card type constant of the given cards.
  int GetWinType(const std::vector<int> &pokes);
  int GetWinTypeStr(const std::string &str);

  // CompareKey compares two hand keys by rank. Returns a positive value when
  // k1 is stronger, 0 when equal, negative when weaker.
  int CompareKey(int64_t k1, int64_t k2);
  int CompareBytes(const std::vector<int> &bytes1, const std::vector<int> &bytes2);
  int CompareStr(const std::string &str1, const std::string &str2);

  // GetHandProbabilityKeyPub is GetHandProbability with hand and public cards
  // given as encoded keys.
  float GetHandProbabilityKeyPub(int64_t hand, int64_t pub);

  // CanonKey returns the suit-canonical key of a situation made of the first
  // nHole cards plus the remaining public cards. Suits are labelled 0..3 by
  // sorting them on the ranks each suit holds (and which of those are hole
  // cards), which is invariant under the 24 suit permutations. Hole and public
  // cards occupy separate key positions because the roles differ. Matches the
  // Go implementation and the generator.
  int64_t CanonKey(const std::vector<int> &cards, int nHole);

  // EquityExact enumerates every completion of the community cards and every
  // opponent hole-card pair, counting wins, ties and losses of the given 2
  // hole cards. Every (board, opponent) pair is counted exactly once and no
  // opponent hand contains a hero card.
  void EquityExact(const std::vector<int> &hand, const std::vector<int> &pub,
                   int64_t &win, int64_t &tie, int64_t &loss);

  // GetHandProbability returns the exact 1v1 win probability of 2 hole cards
  // plus 0-5 community cards against one random opponent hand. Ties count as
  // half a win: P = (W + 0.5*T) / (W + T + L). The answer comes from the first
  // source available for that level: the in-memory table, the packed SQLite
  // database opened by LoadSqlite, then on-the-fly enumeration.
  float GetHandProbability(const std::vector<int> &hand, const std::vector<int> &pub);
  float GetHandProbabilityStr(const std::string &hand, const std::string &pub);

} // namespace texas_algorithm
