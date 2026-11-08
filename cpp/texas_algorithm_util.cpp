#include "texas_algorithm_util.h"

#include <algorithm>
#include <array>
#include <ctime>
#include <fstream>
#include <functional>
#include <iostream>
#include <sstream>
#include <stdexcept>
#include <thread>

#include "gen_extra_opt_util.h"
#include "gen_extra_util.h"
#include "gen_hand_util.h"
#include "gen_opt_util.h"
#include "gen_trans_opt_util.h"
#include "gen_trans_util.h"
#include "gen_util.h"
#include "texas_card_util.h"

namespace texas_algorithm
{

  std::unordered_map<int64_t, KeyData> ColorMap;
  std::unordered_map<int64_t, KeyData> NormalMap;



  namespace
  {

    int64_t NowMillis()
    {
      struct timespec ts;
      clock_gettime(CLOCK_REALTIME, &ts);
      return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
    }

    std::vector<std::string> SplitLine(const std::string &line)
    {
      std::vector<std::string> params;
      size_t start = 0;
      while (true)
      {
        size_t pos = line.find(' ', start);
        if (pos == std::string::npos)
        {
          params.push_back(line.substr(start));
          break;
        }
        params.push_back(line.substr(start, pos - start));
        start = pos + 1;
      }
      return params;
    }

    int64_t ParseInt64(const std::string &s) { return std::strtoll(s.c_str(), nullptr, 10); }

    // ThrowUnknownPoke rejects input that is not part of the standard 52-card
    // deck. Before the joker was removed these paths returned a silent zero /
    // nullptr / false; they now fail loudly.
    [[noreturn]] void ThrowUnknownPoke(const std::string &detail)
    {
      throw std::invalid_argument("无法识别的牌: " + detail);
    }

    // loadKeyFile parses one lookup table file into ColorMap (color=true) or
    // NormalMap (color=false).
    bool LoadKeyFile(const std::string &path, bool color)
    {
      std::ifstream in(path, std::ios::binary);
      if (!in.is_open())
      {
        std::cout << "cannot open " << path << std::endl;
        return false;
      }

      std::string line;
      while (std::getline(in, line))
      {
        std::vector<std::string> params = SplitLine(line);
        int64_t key = ParseInt64(params[0]);
        int i = (int)std::strtol(params[1].c_str(), nullptr, 10);
        int index = (int)std::strtol(params[2].c_str(), nullptr, 10);
        int64_t max = ParseInt64(params[5]);
        int type = (int)std::strtol(params[7].c_str(), nullptr, 10);

        KeyData keyData;
        keyData.index = index;
        keyData.postion = i;
        keyData.max = max;
        keyData.type = type;
        if (color)
        {
          ColorMap[key] = keyData;
        }
        else
        {
          NormalMap[key] = keyData;
        }
      }
      return true;
    }

  } // namespace

  void Gen()
  {
    if (!FileExists("texas_data.txt"))
    {
      GenKey();
      OutputData();
    }
  }

  void GenExtra()
  {
    if (!FileExists("texas_data_extra_6.txt"))
    {
      GenExtraN = 6;
      GenExtraGenKey();
      GenExtraOutputData();
    }
    if (!FileExists("texas_data_extra_5.txt"))
    {
      GenExtraN = 5;
      GenExtraGenKey();
      GenExtraOutputData();
    }
  }

  void GenOpt()
  {
    if (FileExists("texas_data.txt") && !FileExists("texas_data_color.txt") && !FileExists("texas_data_normal.txt"))
    {
      OptNormalData();
      OptColorData();
    }
  }

  void GenExtraOpt()
  {
    if (FileExists("texas_data_extra_6.txt") && !FileExists("texas_data_extra_color_6.txt") &&
        !FileExists("texas_data_extra_normal_6.txt"))
    {
      GenExtraOptN = 6;
      GenExtraOptNormalData();
      GenExtraOptColorData();
    }
    if (FileExists("texas_data_extra_5.txt") && !FileExists("texas_data_extra_color_5.txt") &&
        !FileExists("texas_data_extra_normal_5.txt"))
    {
      GenExtraOptN = 5;
      GenExtraOptNormalData();
      GenExtraOptColorData();
    }
  }

  void GenHand()
  {
    for (int i = 0; i <= 4; i++)
    {
      GenHandN = i;
      GenHandGenKey();
    }
  }

  void GenTrans()
  {
    if (!FileExists("texas_data.txt"))
    {
      return;
    }
    for (int i = 6; i >= 2; i--)
    {
      if (!FileExists("texas_data_" + std::to_string(i) + ".txt"))
      {
        GenTransN = i;
        GenTransGenKey();
        GenTransTransData();
      }
    }
  }

  void GenTransOpt()
  {
    if (!FileExists("texas_data.txt"))
    {
      return;
    }
    for (int i = 6; i >= 2; i--)
    {
      if (FileExists("texas_data_" + std::to_string(i) + ".txt") &&
          !FileExists("texas_data_opt_" + std::to_string(i) + ".txt"))
      {
        GenTransOptN = i;
        GenTransOptData();
      }
    }
  }

  bool IsLoaded() { return !NormalMap.empty() && !ColorMap.empty(); }

  bool IsProbabilityLoaded()
  {
    for (const auto &m : WinRateMap)
    {
      if (m.empty())
      {
        return false;
      }
    }
    return true;
  }

  void Load() { LoadDir("."); }

  void LoadDir(const std::string &dir)
  {
    int64_t begin = NowMillis();

    const char *files[] = {
        "texas_data_color.txt",      "texas_data_normal.txt",     "texas_data_extra_color_6.txt",
        "texas_data_extra_normal_6.txt", "texas_data_extra_color_5.txt", "texas_data_extra_normal_5.txt",
    };
    for (const char *name : files)
    {
      std::string path = dir.empty() ? name : dir + "/" + name;
      bool color = std::string(name).find("color") != std::string::npos;
      if (!LoadKeyFile(path, color))
      {
        return;
      }
    }

    std::cout << "load time " << NowMillis() - begin << std::endl;
  }

  void LoadProbility() { LoadProbilityDir("."); }

  // LoadProbilityDir loads the exact win-rate tables from dir. The three
  // tables are independent: a missing texas_data_winrate_N.txt only makes the
  // matching N-card lookup throw, everything else keeps working.
  void LoadProbilityDir(const std::string &dir)
  {
    int64_t begin = NowMillis();

    int loaded = 0;
    for (int i = 0; i <= 2; i++)
    {
      std::string name = "texas_data_winrate_" + std::to_string(i) + ".txt";
      std::string path = dir.empty() ? name : dir + "/" + name;
      if (LoadWinRateFile(i, path))
      {
        loaded++;
      }
    }

    std::cout << "load time " << NowMillis() - begin << " tables " << loaded << std::endl;
  }

  std::unordered_map<int64_t, WinRateData> WinRateMap[3];

  bool LoadWinRateFile(int n, const std::string &path)
  {
    std::ifstream in(path, std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "open " << path << ": no such file or directory" << std::endl;
      return false;
    }
    std::unordered_map<int64_t, WinRateData> m;
    std::string line;
    while (std::getline(in, line))
    {
      std::vector<std::string> params = SplitLine(line);
      if (params.size() < 4)
      {
        continue;
      }
      WinRateData data;
      int64_t key = ParseInt64(params[0]);
      data.win = ParseInt64(params[1]);
      data.tie = ParseInt64(params[2]);
      data.total = ParseInt64(params[3]);
      m[key] = data;
    }
    WinRateMap[n] = std::move(m);
    return true;
  }

  bool IsWinRateLoaded() { return !WinRateMap[0].empty(); }

  void CheckPoke(int poke)
  {
    if (!IsLegalCardByte(poke))
    {
      ThrowUnknownPoke("牌字节 " + std::to_string(poke));
    }
  }

  void CheckPokes(const std::vector<int> &pokes)
  {
    for (int poke : pokes)
    {
      CheckPoke(poke);
    }
  }

  void CheckKey(int64_t key)
  {
    for (int poke : KeyToPoke(key))
    {
      if (!IsLegalCardByte(poke))
      {
        ThrowUnknownPoke("牌 key " + std::to_string(key));
      }
    }
  }

  int StrToPokeValue(const std::string &str)
  {
    if (str == "A")
    {
      return PokeValueA;
    }
    if (str == "K")
    {
      return PokeValueK;
    }
    if (str == "Q")
    {
      return PokeValueQ;
    }
    if (str == "J")
    {
      return PokeValueJ;
    }
    size_t pos = 0;
    long v = 0;
    try
    {
      v = std::stol(str, &pos);
    }
    catch (const std::exception &)
    {
      ThrowUnknownPoke("牌面 \"" + str + "\"");
    }
    if (pos != str.size() || v < PokeValue2 || v > PokeValueA)
    {
      ThrowUnknownPoke("牌面 \"" + str + "\"");
    }
    return (int)v;
  }

  int StrToPoke(const std::string &str)
  {
    // every suit marker is a 3-byte UTF-8 character
    if (str.size() >= 3)
    {
      std::string head = str.substr(0, 3);
      std::string rest = str.substr(3);
      if (head == "方")
      {
        return PokeToByte(MakePoke(PokeColorFang, StrToPokeValue(rest)));
      }
      if (head == "梅")
      {
        return PokeToByte(MakePoke(PokeColorMei, StrToPokeValue(rest)));
      }
      if (head == "红")
      {
        return PokeToByte(MakePoke(PokeColorHong, StrToPokeValue(rest)));
      }
      if (head == "黑")
      {
        return PokeToByte(MakePoke(PokeColorHei, StrToPokeValue(rest)));
      }
    }
    ThrowUnknownPoke("牌面 \"" + str + "\"");
  }

  std::string KeyToStr(int64_t key) { return ToString(key); }

  std::vector<int> KeyToPoke(int64_t k)
  {
    std::vector<int> cs;
    if (k > 1000000000000LL)
    {
      cs.push_back((int)(k % 100000000000000LL / 1000000000000LL));
    }
    if (k > 10000000000LL)
    {
      cs.push_back((int)(k % 1000000000000LL / 10000000000LL));
    }
    if (k > 100000000LL)
    {
      cs.push_back((int)(k % 10000000000LL / 100000000LL));
    }
    if (k > 1000000LL)
    {
      cs.push_back((int)(k % 100000000LL / 1000000LL));
    }
    if (k > 10000LL)
    {
      cs.push_back((int)(k % 1000000LL / 10000LL));
    }
    if (k > 100LL)
    {
      cs.push_back((int)(k % 10000LL / 100LL));
    }
    if (k > 1LL)
    {
      cs.push_back((int)(k % 100LL / 1LL));
    }
    return cs;
  }

  std::vector<int> StrToPokes(const std::string &str)
  {
    std::vector<int> ret;
    if (str.empty())
    {
      return ret;
    }
    size_t start = 0;
    while (true)
    {
      size_t pos = str.find(',', start);
      if (pos == std::string::npos)
      {
        ret.push_back(StrToPoke(str.substr(start)));
        break;
      }
      ret.push_back(StrToPoke(str.substr(start, pos - start)));
      start = pos + 1;
    }
    return ret;
  }

  std::string PokesToStr(const std::vector<int> &pokes) { return KeyToStr(GenCardBindBytes(pokes)); }

  const KeyData *GetKeyDataStr(const std::string &str)
  {
    std::vector<int> pokes = StrToPokes(str);
    if (pokes.size() != 7)
    {
      ThrowUnknownPoke("需要 7 张牌，实际 " + std::to_string(pokes.size()) + " 张");
    }
    return GetKeyData(pokes);
  }

  const KeyData *GetKeyData(const std::vector<int> &pokes)
  {
    CheckPokes(pokes);
    int64_t key = GenCardBindBytes(pokes);
    const KeyData *data = LookupKeyData(key);
    if (data == nullptr)
    {
      ThrowUnknownPoke("未找到 key " + std::to_string(key) + " 的牌力数据");
    }
    return data;
  }

  const KeyData *GetKeyDataKey(int64_t key)
  {
    CheckKey(key);
    const KeyData *data = LookupKeyData(key);
    if (data == nullptr)
    {
      ThrowUnknownPoke("未找到 key " + std::to_string(key) + " 的牌力数据");
    }
    return data;
  }

  const KeyData *LookupKeyData(int64_t key)
  {
    int64_t colorKey = ChangeColor(key);
    auto colorIt = ColorMap.find(colorKey);
    const KeyData *color = colorIt == ColorMap.end() ? nullptr : &colorIt->second;
    int64_t normalKey = RemoveColor(key);
    auto normalIt = NormalMap.find(normalKey);
    const KeyData *normal = normalIt == NormalMap.end() ? nullptr : &normalIt->second;
    if (color == nullptr)
    {
      return normal;
    }
    if (normal == nullptr)
    {
      return color;
    }
    if (color->index > normal->index)
    {
      return color;
    }
    return normal;
  }

  std::string GetMaxStr(const std::string &str) { return PokesToStr(GetMax(StrToPokes(str))); }

  std::string GetMaxStrHandPub(const std::string &hand, const std::string &pub)
  {
    return PokesToStr(GetMaxHandPub(StrToPokes(hand), StrToPokes(pub)));
  }

  std::vector<int> GetMax(const std::vector<int> &pokes)
  {
    if (pokes.size() < 5 || pokes.size() > 7)
    {
      ThrowUnknownPoke("需要 5-7 张牌，实际 " + std::to_string(pokes.size()) + " 张");
    }
    CheckPokes(pokes);
    std::vector<int> hand(pokes.begin(), pokes.begin() + 2);
    std::vector<int> pub(pokes.begin() + 2, pokes.end());
    return GetMaxHandPub(hand, pub);
  }

  std::vector<int> GetMaxHandPub(const std::vector<int> &hand, const std::vector<int> &pub)
  {
    if (hand.size() != 2)
    {
      ThrowUnknownPoke("手牌需要 2 张，实际 " + std::to_string(hand.size()) + " 张");
    }
    if (pub.size() < 3 || pub.size() > 5)
    {
      ThrowUnknownPoke("公牌需要 3-5 张，实际 " + std::to_string(pub.size()) + " 张");
    }
    CheckPokes(hand);
    CheckPokes(pub);

    std::vector<int> ret;
    std::vector<int> tmp;
    tmp.reserve(hand.size() + pub.size());
    tmp.insert(tmp.end(), hand.begin(), hand.end());
    tmp.insert(tmp.end(), pub.begin(), pub.end());
    const KeyData *keyData = LookupKeyData(GenCardBindBytes(tmp));
    if (keyData == nullptr)
    {
      ThrowUnknownPoke("未找到这手牌的牌力数据");
    }

    std::vector<int> max = KeyToPoke(keyData->max);

    std::vector<int> pubtmp(pub.begin(), pub.end());
    std::vector<int> handtmp(hand.begin(), hand.end());

    if (keyData->type == TexasCardTypeTongHua || keyData->type == TexasCardTypeTongHuaShun ||
        keyData->type == TexasCardTypeKingTongHuaShun)
    {
      int srccolor[4] = {0, 0, 0, 0};
      for (int c : tmp)
      {
        srccolor[c >> 4]++;
      }

      int srcmaxColor = 0;
      int srcmaxColorNum = 0;
      for (int i = 0; i < 4; i++)
      {
        if (srccolor[i] >= srcmaxColorNum)
        {
          srcmaxColor = i;
          srcmaxColorNum = srccolor[i];
        }
      }

      for (size_t i = 0; i < max.size(); i++)
      {
        for (size_t j = 0; j < pubtmp.size(); j++)
        {
          if (pubtmp[j] % 16 == max[i] % 16 && (pubtmp[j] >> 4) == srcmaxColor && max[i] != 0 && pubtmp[j] != 0)
          {
            ret.push_back(pubtmp[j]);

            max[i] = 0;
            pubtmp[j] = 0;
            break;
          }
        }
      }

      if (ret.size() < 5)
      {
        for (size_t i = 0; i < max.size(); i++)
        {
          for (size_t j = 0; j < handtmp.size(); j++)
          {
            if (handtmp[j] % 16 == max[i] % 16 && (handtmp[j] >> 4) == srcmaxColor && max[i] != 0 && handtmp[j] != 0)
            {
              ret.push_back(handtmp[j]);

              max[i] = 0;
              handtmp[j] = 0;
              break;
            }
          }
        }
      }

      for (size_t i = 0; i < max.size(); i++)
      {
        if (max[i] != 0)
        {
          max[i] = (srcmaxColor << 4) | (max[i] % 16);
        }
      }
    }
    else
    {
      for (size_t j = 0; j < pubtmp.size(); j++)
      {
        for (size_t i = 0; i < max.size(); i++)
        {
          if (pubtmp[j] == max[i] && max[i] != 0 && pubtmp[j] != 0)
          {
            ret.push_back(pubtmp[j]);

            max[i] = 0;
            pubtmp[j] = 0;
            break;
          }
        }
      }

      for (size_t j = 0; j < handtmp.size(); j++)
      {
        for (size_t i = 0; i < max.size(); i++)
        {
          if (handtmp[j] == max[i] && max[i] != 0 && handtmp[j] != 0)
          {
            ret.push_back(handtmp[j]);

            max[i] = 0;
            handtmp[j] = 0;
            break;
          }
        }
      }

      for (size_t i = 0; i < max.size(); i++)
      {
        for (size_t j = 0; j < pubtmp.size(); j++)
        {
          if (pubtmp[j] % 16 == max[i] % 16 && max[i] != 0 && pubtmp[j] != 0)
          {
            ret.push_back(pubtmp[j]);

            max[i] = 0;
            pubtmp[j] = 0;
            break;
          }
        }
      }

      if (ret.size() < 5)
      {
        for (size_t i = 0; i < max.size(); i++)
        {
          for (size_t j = 0; j < handtmp.size(); j++)
          {
            if (handtmp[j] % 16 == max[i] % 16 && max[i] != 0 && handtmp[j] != 0)
            {
              ret.push_back(handtmp[j]);

              max[i] = 0;
              handtmp[j] = 0;
              break;
            }
          }
        }
      }
    }

    std::sort(ret.begin(), ret.end());

    return ret;
  }

  int GetWinPosition(const std::vector<int> &pokes) { return GetKeyData(pokes)->postion; }

  int GetWinPositionStr(const std::string &str) { return GetWinPosition(StrToPokes(str)); }

  double GetWinProbability(const std::vector<int> &pokes)
  {
    const KeyData *keyData = GetKeyData(pokes);
    int64_t total = 1;
    for (size_t i = 0; i < pokes.size(); i++)
    {
      total = total * (GenNum - (int64_t)i);
    }
    for (int i = (int)pokes.size(); i >= 1; i--)
    {
      total = total / i;
    }
    return (double)keyData->index / (double)total;
  }

  double GetWinProbabilityStr(const std::string &str) { return GetWinProbability(StrToPokes(str)); }

  int64_t GetWinMax(const std::vector<int> &pokes) { return GetKeyData(pokes)->max; }

  int64_t GetWinMaxStr(const std::string &str) { return GetWinMax(StrToPokes(str)); }

  int GetWinType(const std::vector<int> &pokes) { return GetKeyData(pokes)->type; }

  int GetWinTypeStr(const std::string &str) { return GetWinType(StrToPokes(str)); }

  int CompareKeyLookup(int64_t k1, int64_t k2)
  {
    const KeyData *keyData1 = LookupKeyData(k1);
    const KeyData *keyData2 = LookupKeyData(k2);
    if (keyData1 == nullptr && keyData2 == nullptr)
    {
      return 0;
    }
    if (keyData1 == nullptr)
    {
      return -1;
    }
    if (keyData2 == nullptr)
    {
      return 1;
    }
    return keyData1->postion - keyData2->postion;
  }

  int CompareKey(int64_t k1, int64_t k2)
  {
    CheckKey(k1);
    CheckKey(k2);
    return CompareKeyLookup(k1, k2);
  }

  int CompareBytes(const std::vector<int> &bytes1, const std::vector<int> &bytes2)
  {
    CheckPokes(bytes1);
    CheckPokes(bytes2);
    return CompareKeyLookup(GenCardBindBytes(bytes1), GenCardBindBytes(bytes2));
  }

  int CompareStr(const std::string &str1, const std::string &str2)
  {
    return CompareBytes(StrToPokes(str1), StrToPokes(str2));
  }

  float GetHandProbabilityKeyPub(int64_t hand, int64_t pub)
  {
    CheckKey(hand);
    CheckKey(pub);
    return GetHandProbability(KeyToPoke(hand), KeyToPoke(pub));
  }

  int64_t CanonKey(const std::vector<int> &cards, int nHole)
  {
    int mask[4] = {0, 0, 0, 0};
    int hole[4] = {0, 0, 0, 0};
    for (int i = 0; i < (int)cards.size() && i < 5; i++)
    {
      int c = cards[i];
      mask[c >> 4] |= 1 << (c & 0x0f);
      if (i < nHole)
      {
        hole[c >> 4] |= 1 << (c & 0x0f);
      }
    }
    int8_t label[4] = {0, 0, 0, 0};
    bool used[4] = {false, false, false, false};
    for (int8_t k = 0; k < 4; k++)
    {
      int best = -1;
      int bestSig = 0;
      for (int s = 0; s < 4; s++)
      {
        if (used[s])
        {
          continue;
        }
        int sig = (mask[s] << 16) | hole[s];
        if (best < 0 || sig < bestSig)
        {
          best = s;
          bestSig = sig;
        }
      }
      used[best] = true;
      label[best] = k;
    }

    uint8_t holeCards[5] = {0};
    uint8_t pubCards[5] = {0};
    int nh = 0, np = 0;
    for (int i = 0; i < (int)cards.size() && i < 5; i++)
    {
      int c = cards[i];
      uint8_t b = (uint8_t)((label[c >> 4] << 4) | (c & 0x0f));
      if (i < nHole)
      {
        holeCards[nh++] = b;
      }
      else
      {
        pubCards[np++] = b;
      }
    }
    std::sort(holeCards, holeCards + nh);
    std::sort(pubCards, pubCards + np);

    int64_t k = 0;
    for (int i = 0; i < nh; i++)
    {
      k = k * 100 + (int64_t)holeCards[i];
    }
    for (int i = 0; i < np; i++)
    {
      k = k * 100 + (int64_t)pubCards[i];
    }
    return k;
  }

  namespace
  {
    // key7Board merges two ascending cards x<y into the ascending 5-card board
    // and returns the 7-card key.
    int64_t Key7Board(const uint8_t *b, uint8_t x, uint8_t y)
    {
      uint8_t buf[7];
      int i = 0, j = 0;
      while (j < 5)
      {
        if (i == 0 && x < b[j])
        {
          buf[i + j] = x;
          i++;
          continue;
        }
        if (i == 1 && y < b[j])
        {
          buf[i + j] = y;
          i++;
          continue;
        }
        buf[i + j] = b[j];
        j++;
      }
      if (i == 0)
      {
        buf[5] = x;
        buf[6] = y;
      }
      else if (i == 1)
      {
        buf[6] = y;
      }
      int64_t k = 0;
      for (int n = 0; n < 7; n++)
      {
        k = k * 100 + (int64_t)buf[n];
      }
      return k;
    }

    struct EquityCount
    {
      int64_t win = 0;
      int64_t tie = 0;
      int64_t loss = 0;
    };

    void CountBoards(const std::vector<std::array<uint8_t, 5>> &boards, int from, int to,
                     const std::vector<uint8_t> &rest, uint8_t hx, uint8_t hy, EquityCount &out)
    {
      for (int idx = from; idx < to; idx++)
      {
        const std::array<uint8_t, 5> &b = boards[idx];
        uint8_t opp[45];
        int n = 0;
        for (uint8_t c : rest)
        {
          if (c == b[0] || c == b[1] || c == b[2] || c == b[3] || c == b[4])
          {
            continue;
          }
          opp[n] = c;
          n++;
        }
        const KeyData *hero = LookupKeyData(Key7Board(b.data(), hx, hy));
        if (hero == nullptr)
        {
          continue;
        }
        for (int i = 0; i < n; i++)
        {
          for (int j = i + 1; j < n; j++)
          {
            const KeyData *data = LookupKeyData(Key7Board(b.data(), opp[i], opp[j]));
            if (data == nullptr)
            {
              continue;
            }
            if (hero->index > data->index)
            {
              out.win++;
            }
            else if (hero->index == data->index)
            {
              out.tie++;
            }
            else
            {
              out.loss++;
            }
          }
        }
      }
    }
  } // namespace

  void EquityExact(const std::vector<int> &hand, const std::vector<int> &pub,
                   int64_t &win, int64_t &tie, int64_t &loss)
  {
    win = tie = loss = 0;
    std::vector<uint8_t> known;
    for (int c : hand)
    {
      known.push_back((uint8_t)c);
    }
    for (int c : pub)
    {
      known.push_back((uint8_t)c);
    }
    std::vector<uint8_t> rest;
    for (int color = 0; color < 4; color++)
    {
      for (int v = PokeValue2; v <= PokeValueA; v++)
      {
        uint8_t c = (uint8_t)((color << 4) | v);
        bool used = false;
        for (uint8_t u : known)
        {
          if (u == c)
          {
            used = true;
            break;
          }
        }
        if (!used)
        {
          rest.push_back(c);
        }
      }
    }

    int need = 5 - (int)pub.size();
    std::array<uint8_t, 5> board = {0, 0, 0, 0, 0};
    for (int i = 0; i < (int)pub.size(); i++)
    {
      board[i] = (uint8_t)pub[i];
    }
    std::vector<std::array<uint8_t, 5>> boards;
    std::function<void(int, int)> gen = [&](int start, int depth)
    {
      if (depth == need)
      {
        std::array<uint8_t, 5> b = board;
        std::sort(b.begin(), b.end());
        boards.push_back(b);
        return;
      }
      for (int i = start; i <= (int)rest.size() - (need - depth); i++)
      {
        board[pub.size() + depth] = rest[i];
        gen(i + 1, depth + 1);
      }
    };
    gen(0, 0);

    uint8_t hx = (uint8_t)hand[0], hy = (uint8_t)hand[1];
    if (hx > hy)
    {
      std::swap(hx, hy);
    }

    int workers = 1;
    // 20k evaluations is roughly where the thread hand-off pays off: N>=4
    // (45,540 pairs) benefits, N=5 (990) does not.
    if ((int64_t)boards.size() * 990 >= 20000)
    {
      workers = (int)std::thread::hardware_concurrency();
      if (workers < 1)
      {
        workers = 1;
      }
    }
    std::vector<EquityCount> results(workers);
    int chunk = (int)((boards.size() + workers - 1) / workers);
    std::vector<std::thread> threads;
    for (int w = 0; w < workers; w++)
    {
      int from = w * chunk;
      int to = from + chunk;
      if (to > (int)boards.size())
      {
        to = (int)boards.size();
      }
      if (from >= to)
      {
        continue;
      }
      threads.emplace_back(CountBoards, std::cref(boards), from, to, std::cref(rest), hx, hy,
                           std::ref(results[w]));
    }
    for (std::thread &t : threads)
    {
      t.join();
    }
    for (const EquityCount &r : results)
    {
      win += r.win;
      tie += r.tie;
      loss += r.loss;
    }
  }

  float GetHandProbability(const std::vector<int> &hand, const std::vector<int> &pub)
  {
    CheckPokes(hand);
    CheckPokes(pub);
    if (hand.size() != 2 || pub.size() > 5)
    {
      throw std::invalid_argument("无法识别的牌: 手牌必须为 2 张，公牌最多 5 张");
    }
    if (pub.size() >= 3)
    {
      int64_t win = 0, tie = 0, loss = 0;
      EquityExact(hand, pub, win, tie, loss);
      int64_t total = win + tie + loss;
      if (total == 0)
      {
        throw std::invalid_argument("缺少胜率表数据");
      }
      return (float)((double)(2 * win + tie) / (double)(2 * total));
    }
    const auto &m = WinRateMap[pub.size()];
    if (m.empty())
    {
      throw std::invalid_argument("缺少胜率表数据");
    }
    std::vector<int> cards;
    cards.insert(cards.end(), hand.begin(), hand.end());
    cards.insert(cards.end(), pub.begin(), pub.end());
    auto it = m.find(CanonKey(cards, 2));
    if (it == m.end() || it->second.total == 0)
    {
      throw std::invalid_argument("缺少胜率表数据");
    }
    return (float)((double)(2 * it->second.win + it->second.tie) / (double)(2 * it->second.total));
  }

  float GetHandProbabilityStr(const std::string &hand, const std::string &pub)
  {
    return GetHandProbability(StrToPokes(hand), StrToPokes(pub));
  }

} // namespace texas_algorithm
