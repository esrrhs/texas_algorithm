#include "texas_algorithm_util.h"

#include <algorithm>
#include <ctime>
#include <fstream>
#include <iostream>
#include <sstream>
#include <stdexcept>

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

  std::unordered_map<int64_t, ProbilityData> ProbilityMap[7];
  std::unordered_map<int64_t, ProbilityData> OptProbilityMap[7];

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

    // loadProbilityFile parses one probability table file into
    // ProbilityMap[n] / OptProbilityMap[n]. Rows with type 0 go to the plain
    // table, others to the optimized table.
    bool LoadProbilityFile(int n, const std::string &path)
    {
      std::ifstream in(path, std::ios::binary);
      if (!in.is_open())
      {
        std::cout << "cannot open " << path << std::endl;
        return false;
      }

      ProbilityMap[n].clear();
      OptProbilityMap[n].clear();

      std::string line;
      while (std::getline(in, line))
      {
        std::vector<std::string> params = SplitLine(line);
        int64_t key = ParseInt64(params[0]);
        int64_t type = ParseInt64(params[1]);
        float probility = std::strtof(params[2].c_str(), nullptr);
        float min = std::strtof(params[3].c_str(), nullptr);
        float max = std::strtof(params[4].c_str(), nullptr);

        ProbilityData data;
        data.avg = probility;
        data.min = min;
        data.max = max;
        if (type == 0)
        {
          ProbilityMap[n][key] = data;
        }
        else
        {
          OptProbilityMap[n][key] = data;
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

  bool IsProbabilityLoaded() { return !ProbilityMap[2].empty(); }

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

  void LoadProbilityDir(const std::string &dir)
  {
    int64_t begin = NowMillis();

    for (int i = 6; i >= 2; i--)
    {
      std::string name = "texas_data_opt_" + std::to_string(i) + ".txt";
      std::string path = dir.empty() ? name : dir + "/" + name;
      if (!LoadProbilityFile(i, path))
      {
        return;
      }
    }

    std::cout << "load time " << NowMillis() - begin << std::endl;
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
      throw std::invalid_argument("invalid poke value: " + str);
    }
    if (pos != str.size() || v < 0 || v > 127)
    {
      throw std::invalid_argument("invalid poke value: " + str);
    }
    return (int)v;
  }

  int StrToPoke(const std::string &str)
  {
    // every suit marker (and 鬼) is a 3-byte UTF-8 character
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
      if (head == "鬼")
      {
        return PokeToByte(GUI);
      }
    }
    return 0;
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
      return nullptr;
    }
    return GetKeyData(pokes);
  }

  const KeyData *GetKeyData(const std::vector<int> &pokes) { return GetKeyDataKey(GenCardBindBytes(pokes)); }

  const KeyData *GetKeyDataKey(int64_t key)
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

  std::pair<std::string, std::vector<int>> GetMaxStr(const std::string &str)
  {
    std::pair<std::vector<int>, std::vector<int>> res = GetMax(StrToPokes(str));
    return {PokesToStr(res.first), res.second};
  }

  std::pair<std::string, std::vector<int>> GetMaxStrHandPub(const std::string &hand, const std::string &pub)
  {
    std::pair<std::vector<int>, std::vector<int>> res = GetMaxHandPub(StrToPokes(hand), StrToPokes(pub));
    return {PokesToStr(res.first), res.second};
  }

  std::pair<std::vector<int>, std::vector<int>> GetMax(const std::vector<int> &pokes)
  {
    if (pokes.size() < 5 || pokes.size() > 7)
    {
      return {{}, {}};
    }
    std::vector<int> hand(pokes.begin(), pokes.begin() + 2);
    std::vector<int> pub(pokes.begin() + 2, pokes.end());
    return GetMaxHandPub(hand, pub);
  }

  std::pair<std::vector<int>, std::vector<int>> GetMaxHandPub(const std::vector<int> &hand,
                                                              const std::vector<int> &pub)
  {
    std::vector<int> ret;
    if (hand.size() != 2)
    {
      return {ret, {}};
    }
    if (pub.size() < 3 || pub.size() > 5)
    {
      return {ret, {}};
    }
    std::vector<int> tmp;
    tmp.reserve(hand.size() + pub.size());
    tmp.insert(tmp.end(), hand.begin(), hand.end());
    tmp.insert(tmp.end(), pub.begin(), pub.end());
    const KeyData *keyData = GetKeyData(tmp);
    if (keyData == nullptr)
    {
      return {ret, {}};
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
        if (!IsGuiByte(c))
        {
          srccolor[c >> 4]++;
        }
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
          if (pubtmp[j] == max[i] && max[i] != 0 && pubtmp[j] != 0 && !IsGuiByte(pubtmp[j]))
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
          if (handtmp[j] == max[i] && max[i] != 0 && handtmp[j] != 0 && !IsGuiByte(handtmp[j]))
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
          if (pubtmp[j] % 16 == max[i] % 16 && max[i] != 0 && pubtmp[j] != 0 && !IsGuiByte(pubtmp[j]))
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
            if (handtmp[j] % 16 == max[i] % 16 && max[i] != 0 && handtmp[j] != 0 && !IsGuiByte(handtmp[j]))
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

    while (ret.size() < 5)
    {
      ret.push_back(PokeToByte(GUI));
    }

    std::vector<int> guiTrans;
    for (int m : max)
    {
      if (m != 0)
      {
        guiTrans.push_back(m);
      }
    }

    std::sort(ret.begin(), ret.end());

    return {ret, guiTrans};
  }

  int GetWinPosition(const std::vector<int> &pokes)
  {
    const KeyData *keyData = GetKeyData(pokes);
    if (keyData == nullptr)
    {
      return 0;
    }
    return keyData->postion;
  }

  int GetWinPositionStr(const std::string &str) { return GetWinPosition(StrToPokes(str)); }

  double GetWinProbability(const std::vector<int> &pokes)
  {
    const KeyData *keyData = GetKeyData(pokes);
    if (keyData == nullptr)
    {
      return 0;
    }
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

  int64_t GetWinMax(const std::vector<int> &pokes)
  {
    const KeyData *keyData = GetKeyData(pokes);
    if (keyData == nullptr)
    {
      return 0;
    }
    return keyData->max;
  }

  int64_t GetWinMaxStr(const std::string &str) { return GetWinMax(StrToPokes(str)); }

  int GetWinType(const std::vector<int> &pokes)
  {
    const KeyData *keyData = GetKeyData(pokes);
    if (keyData == nullptr)
    {
      return 0;
    }
    return keyData->type;
  }

  int GetWinTypeStr(const std::string &str) { return GetWinType(StrToPokes(str)); }

  int CompareKey(int64_t k1, int64_t k2)
  {
    const KeyData *keyData1 = GetKeyDataKey(k1);
    const KeyData *keyData2 = GetKeyDataKey(k2);
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

  int CompareBytes(const std::vector<int> &bytes1, const std::vector<int> &bytes2)
  {
    return CompareKey(GenCardBindBytes(bytes1), GenCardBindBytes(bytes2));
  }

  int CompareStr(const std::string &str1, const std::string &str2)
  {
    return CompareBytes(StrToPokes(str1), StrToPokes(str2));
  }

  const ProbilityData *GetHandProbabilityKey(int64_t k)
  {
    int num = 0;
    if (k > 10000000000LL)
    {
      num = 6;
    }
    else if (k > 100000000LL)
    {
      num = 5;
    }
    else if (k > 1000000LL)
    {
      num = 4;
    }
    else if (k > 10000LL)
    {
      num = 3;
    }
    else if (k > 100LL)
    {
      num = 2;
    }
    if (num < 2 || num > 6)
    {
      return nullptr;
    }
    if (ProbilityMap[num].empty() && OptProbilityMap[num].empty())
    {
      return nullptr;
    }

    auto it = ProbilityMap[num].find(k);
    if (it != ProbilityMap[num].end())
    {
      return &it->second;
    }
    k = RemoveColor(k);
    auto optIt = OptProbilityMap[num].find(k);
    return optIt == OptProbilityMap[num].end() ? nullptr : &optIt->second;
  }

  float GetHandProbabilityKeyPub(int64_t hand, int64_t pub)
  {
    return GetHandProbability(KeyToPoke(hand), KeyToPoke(pub));
  }

  float GetHandProbability(const std::vector<int> &hand, const std::vector<int> &pub)
  {
    std::vector<int> allCards;
    allCards.reserve(hand.size() + pub.size());
    allCards.insert(allCards.end(), hand.begin(), hand.end());
    allCards.insert(allCards.end(), pub.begin(), pub.end());
    std::sort(allCards.begin(), allCards.end());
    std::vector<int> pubCopy(pub.begin(), pub.end());
    std::sort(pubCopy.begin(), pubCopy.end());
    int64_t pubkey = GenCardBindBytes(pubCopy);

    const ProbilityData *pubProbilityData = GetHandProbabilityKey(pubkey);

    float avg = 0;
    if (allCards.size() == 7)
    {
      avg = (float)GetWinProbability(allCards);
    }
    else
    {
      int64_t totalkey = GenCardBindBytes(allCards);
      const ProbilityData *totalProbilityData = GetHandProbabilityKey(totalkey);
      if (totalProbilityData == nullptr)
      {
        return 0;
      }
      avg = totalProbilityData->avg;
    }

    if (pubProbilityData == nullptr)
    {
      return avg;
    }

    float p = 0.5f;

    if (avg > pubProbilityData->avg)
    {
      p += 0.5f * (avg - pubProbilityData->avg) / (pubProbilityData->max - pubProbilityData->avg);
    }
    else
    {
      p += 0.5f * (avg - pubProbilityData->avg) / (pubProbilityData->avg - pubProbilityData->min);
    }

    if (p > 1)
    {
      p = 1;
    }
    if (p < 0)
    {
      p = 0;
    }

    return p;
  }

  float GetHandProbabilityStr(const std::string &hand, const std::string &pub)
  {
    return GetHandProbability(StrToPokes(hand), StrToPokes(pub));
  }

} // namespace texas_algorithm
