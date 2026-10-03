#include "gen_opt_util.h"

#include <algorithm>
#include <fstream>
#include <iostream>
#include <sstream>
#include <ctime>
#include <unordered_set>

#include "gen_util.h"
#include "texas_card_util.h"

namespace texas_algorithm
{

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

    // Package-level generation state mirroring Java's GenOptUtil static fields.
    int64_t g_optTotalKey = 0;
    int g_optLastPrint = 0;
    int64_t g_optBeginPrint = 0;

    void PrintProgress(int64_t processed, int64_t total)
    {
      int cur = (int)(processed * 100 / total);
      if (cur != g_optLastPrint)
      {
        g_optLastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_optBeginPrint) / (double)processed;
        std::cout << cur << "% 需要" << per * (double)(total - processed) / 60 / 1000 << "分"
                  << " 用时" << (now - g_optBeginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)processed / ((double)(now - g_optBeginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    bool IsFlushType(int64_t maxType)
    {
      return maxType == TexasCardTypeTongHua || maxType == TexasCardTypeTongHuaShun ||
             maxType == TexasCardTypeKingTongHuaShun;
    }

  } // namespace

  void OptColorData()
  {
    std::ifstream in("texas_data.txt", std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open texas_data.txt" << std::endl;
      return;
    }

    std::ofstream out("texas_data_color.txt", std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open texas_data_color.txt" << std::endl;
      return;
    }

    g_optTotalKey = 0;
    g_optLastPrint = 0;
    g_optBeginPrint = NowMillis();
    std::unordered_set<int64_t> keys;

    std::string line;
    while (std::getline(in, line))
    {
      std::vector<std::string> params = SplitLine(line);
      int64_t key = ParseInt64(params[0]);
      int64_t i = ParseInt64(params[1]);
      int64_t index = ParseInt64(params[2]);
      int64_t total = ParseInt64(params[3]);
      int64_t maxType = ParseInt64(params[7]);

      if (IsFlushType(maxType))
      {
        int64_t colorKey = ChangeColor(key);

        if (keys.find(colorKey) == keys.end())
        {
          std::string str = std::to_string(colorKey) + " " + std::to_string(i) + " " + std::to_string(index) + " " +
                            std::to_string(total) + " " + ToString(colorKey) + " " + std::to_string(Max(colorKey)) +
                            " " + ToString(Max(colorKey)) + " " + std::to_string(maxType) + "\n";
          out << str;
          keys.insert(colorKey);
        }
      }

      g_optTotalKey++;

      PrintProgress(g_optTotalKey, Total);
    }

    out.close();

    std::cout << "optData finish " << g_optTotalKey << std::endl;
  }

  void OptNormalData()
  {
    std::ifstream in("texas_data.txt", std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open texas_data.txt" << std::endl;
      return;
    }

    std::ofstream out("texas_data_normal.txt", std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open texas_data_normal.txt" << std::endl;
      return;
    }

    g_optTotalKey = 0;
    g_optLastPrint = 0;
    g_optBeginPrint = NowMillis();
    std::unordered_set<int64_t> keys;

    std::string line;
    while (std::getline(in, line))
    {
      std::vector<std::string> params = SplitLine(line);
      int64_t key = ParseInt64(params[0]);
      int64_t i = ParseInt64(params[1]);
      int64_t index = ParseInt64(params[2]);
      int64_t total = ParseInt64(params[3]);
      const std::string &keystr = params[4];
      int64_t max = ParseInt64(params[5]);
      const std::string &maxstr = params[6];
      int64_t maxType = ParseInt64(params[7]);
      int64_t removeKey = RemoveColor(key);
      if (!IsFlushType(maxType))
      {
        if (keys.find(removeKey) == keys.end())
        {
          std::string str = std::to_string(removeKey) + " " + std::to_string(i) + " " + std::to_string(index) + " " +
                            std::to_string(total) + " " + keystr + " " + std::to_string(max) + " " + maxstr + " " +
                            std::to_string(maxType) + "\n";
          out << str;
          keys.insert(removeKey);
        }
      }

      g_optTotalKey++;

      PrintProgress(g_optTotalKey, Total);
    }

    out.close();

    std::cout << "optData finish " << g_optTotalKey << std::endl;
  }

  namespace
  {

    std::vector<int> KeyToCards(int64_t k)
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

  } // namespace

  int64_t RemoveColor(int64_t k)
  {
    std::vector<int> cs = KeyToCards(k);

    for (int &c : cs)
    {
      if (!IsGuiByte(c))
      {
        c = (PokeColorFang << 4) | (c % 16);
      }
    }

    std::sort(cs.begin(), cs.end());

    return GenCardBindInts(cs);
  }

  int64_t ChangeColor(int64_t k)
  {
    std::vector<int> cs = KeyToCards(k);

    int color[4] = {0, 0, 0, 0};
    for (int i : cs)
    {
      if (!IsGuiByte(i))
      {
        color[i >> 4]++;
      }
    }

    int maxColor = 0;
    int maxColorNum = 0;
    for (int i = 0; i < 4; i++)
    {
      if (color[i] > maxColorNum)
      {
        maxColor = i;
        maxColorNum = color[i];
      }
    }

    for (int &c : cs)
    {
      if (!IsGuiByte(c))
      {
        if ((c >> 4) == maxColor)
        {
          c = (PokeColorHei << 4) | (c % 16);
        }
        else
        {
          c = (PokeColorFang << 4) | (c % 16);
        }
      }
    }

    std::sort(cs.begin(), cs.end());

    return GenCardBindInts(cs);
  }

} // namespace texas_algorithm
