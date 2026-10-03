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

  // ProbilityData holds win-rate statistics of one hand key.
  struct ProbilityData
  {
    float avg = 0;
    float min = 0;
    float max = 0;
  };

  extern std::unordered_map<int64_t, KeyData> ColorMap;
  extern std::unordered_map<int64_t, KeyData> NormalMap;

  extern std::unordered_map<int64_t, ProbilityData> ProbilityMap[7];
  extern std::unordered_map<int64_t, ProbilityData> OptProbilityMap[7];

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

  // StrToPokeValue converts a rank string ("2".."10", "J", "Q", "K", "A") into
  // its byte value. Throws std::invalid_argument on unknown input, mirroring
  // Java's Byte.parseByte.
  int StrToPokeValue(const std::string &str);

  // StrToPoke converts a card string ("方A", "黑10", "鬼"...) into its byte.
  int StrToPoke(const std::string &str);

  // KeyToStr renders a key as its concatenated card strings.
  std::string KeyToStr(int64_t key);

  // KeyToPoke decodes a key into its card bytes, most significant card first.
  std::vector<int> KeyToPoke(int64_t k);

  // StrToPokes converts a comma-separated card string into card bytes.
  std::vector<int> StrToPokes(const std::string &str);

  // PokesToStr renders card bytes as their concatenated card strings.
  std::string PokesToStr(const std::vector<int> &pokes);

  // GetKeyDataStr looks up a 7-card comma-separated string; returns nullptr
  // when the string is not exactly 7 cards.
  const KeyData *GetKeyDataStr(const std::string &str);

  // GetKeyData looks up the best-hand data of the given cards.
  const KeyData *GetKeyData(const std::vector<int> &pokes);

  // GetKeyDataKey looks up the best-hand data of the given 7-card key: the
  // suit-normalized flush table and the suit-stripped normal table are both
  // consulted, and the entry with the higher rank wins.
  const KeyData *GetKeyDataKey(int64_t key);

  // GetMaxStr returns the best 5-card hand of a comma-separated card list
  // (5-7 cards) together with the wild-card substitution list.
  std::pair<std::string, std::vector<int>> GetMaxStr(const std::string &str);

  // GetMaxStrHandPub returns the best 5-card hand of 2 hole cards and 3-5
  // community cards, together with the wild-card substitution list.
  std::pair<std::string, std::vector<int>> GetMaxStrHandPub(const std::string &hand, const std::string &pub);

  // GetMax returns the best 5-card hand of the given cards (2 hole cards
  // followed by 3-5 community cards) together with the wild-card substitution
  // list.
  std::pair<std::vector<int>, std::vector<int>> GetMax(const std::vector<int> &pokes);

  // GetMaxHandPub returns the best 5-card hand of 2 hole cards and 3-5
  // community cards, together with the wild-card substitution list. It maps
  // the stored best-hand key back onto the actual cards that were dealt.
  std::pair<std::vector<int>, std::vector<int>> GetMaxHandPub(const std::vector<int> &hand,
                                                              const std::vector<int> &pub);

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

  // GetHandProbabilityKey returns the stored win-rate statistics of a key with
  // 2-6 cards, or nullptr when out of range or not loaded.
  const ProbilityData *GetHandProbabilityKey(int64_t k);

  // GetHandProbabilityKeyPub is GetHandProbability with hand and public cards
  // given as encoded keys.
  float GetHandProbabilityKeyPub(int64_t hand, int64_t pub);

  // GetHandProbability estimates the 1v1 win probability of 2 hole cards plus
  // 0-4 community cards. Requires the probability tables to be loaded.
  float GetHandProbability(const std::vector<int> &hand, const std::vector<int> &pub);
  float GetHandProbabilityStr(const std::string &hand, const std::string &pub);

} // namespace texas_algorithm
