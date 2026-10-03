#include "gen_extra_opt_util.h"

#include <algorithm>
#include <ctime>
#include <fstream>
#include <iostream>
#include <unordered_set>

#include "gen_opt_util.h"
#include "gen_util.h"
#include "texas_card_util.h"

namespace texas_algorithm
{

  int GenExtraOptN = 6;

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

    int64_t g_total = 0;
    int64_t g_totalKey = 0;
    int g_lastPrint = 0;
    int64_t g_beginPrint = 0;

    void PrintProgress(int64_t processed, int64_t total)
    {
      int cur = (int)(processed * 100 / total);
      if (cur != g_lastPrint)
      {
        g_lastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_beginPrint) / (double)processed;
        std::cout << cur << "% 需要" << per * (double)(total - processed) / 60 / 1000 << "分"
                  << " 用时" << (now - g_beginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)processed / ((double)(now - g_beginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    bool IsFlushType(int64_t maxType)
    {
      return maxType == TexasCardTypeTongHua || maxType == TexasCardTypeTongHuaShun ||
             maxType == TexasCardTypeKingTongHuaShun;
    }

  } // namespace

  void GenExtraOptColorData()
  {
    g_total = 1;
    for (int i = 0; i < GenExtraOptN; i++)
    {
      g_total *= GenNum - i;
    }
    for (int i = GenExtraOptN; i >= 1; i--)
    {
      g_total /= i;
    }

    std::string inName = "texas_data_extra_" + std::to_string(GenExtraOptN) + ".txt";
    std::ifstream in(inName, std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open " << inName << std::endl;
      return;
    }

    std::string outName = "texas_data_extra_color_" + std::to_string(GenExtraOptN) + ".txt";
    std::ofstream out(outName, std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open " << outName << std::endl;
      return;
    }

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();
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

      g_totalKey++;

      PrintProgress(g_totalKey, g_total);
    }

    out.close();

    std::cout << "optData finish " << g_totalKey << std::endl;
  }

  void GenExtraOptNormalData()
  {
    g_total = 1;
    for (int i = 0; i < GenExtraOptN; i++)
    {
      g_total *= GenNum - i;
    }
    for (int i = GenExtraOptN; i >= 1; i--)
    {
      g_total /= i;
    }

    std::string inName = "texas_data_extra_" + std::to_string(GenExtraOptN) + ".txt";
    std::ifstream in(inName, std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open " << inName << std::endl;
      return;
    }

    std::string outName = "texas_data_extra_normal_" + std::to_string(GenExtraOptN) + ".txt";
    std::ofstream out(outName, std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open " << outName << std::endl;
      return;
    }

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();
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
      if (!(maxType == TexasCardTypeTongHua || maxType == TexasCardTypeTongHuaShun ||
            maxType == TexasCardTypeKingTongHuaShun))
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

      g_totalKey++;

      PrintProgress(g_totalKey, g_total);
    }

    out.close();

    std::cout << "optData finish " << g_totalKey << std::endl;
  }

} // namespace texas_algorithm
