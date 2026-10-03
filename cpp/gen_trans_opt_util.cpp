#include "gen_trans_opt_util.h"

#include <ctime>
#include <fstream>
#include <iostream>
#include <map>
#include <unordered_map>

#include "gen_opt_util.h"
#include "gen_util.h"

namespace texas_algorithm
{

  int GenTransOptN = 6;

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

    // GenTransOptKeyData groups the win-rate strings of all keys sharing the
    // same suit-stripped key, tracking the most frequent one.
    struct GenTransOptKeyData
    {
      std::string max;
      std::unordered_map<std::string, int> ps;
    };

    int64_t g_total = 1;
    int64_t g_totalKey = 0;
    int g_lastPrint = 0;
    int64_t g_beginPrint = 0;

    void PrintProgress(const char *step, int64_t processed, int64_t total)
    {
      int cur = (int)(processed * 100 / total);
      if (cur != g_lastPrint)
      {
        g_lastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_beginPrint) / (double)processed;
        std::cout << step << " N" << GenTransOptN << " " << cur << "% 需要"
                  << per * (double)(total - processed) / 60 / 1000 << "分"
                  << " 用时" << (now - g_beginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)processed / ((double)(now - g_beginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

  } // namespace

  void GenTransOptData()
  {
    g_total = 1;
    for (int i = 0; i < GenTransOptN; i++)
    {
      g_total *= GenNum - i;
    }
    for (int i = GenTransOptN; i >= 1; i--)
    {
      g_total /= i;
    }

    std::string inName = "texas_data_" + std::to_string(GenTransOptN) + ".txt";
    std::ifstream in(inName, std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open " << inName << std::endl;
      return;
    }

    std::string outName = "texas_data_opt_" + std::to_string(GenTransOptN) + ".txt";
    std::ofstream out(outName, std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open " << outName << std::endl;
      return;
    }

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();
    std::unordered_map<int64_t, std::string> keys;
    std::unordered_map<int64_t, GenTransOptKeyData> optkeys;

    std::string line;
    while (std::getline(in, line))
    {
      std::vector<std::string> params = SplitLine(line);
      int64_t key = ParseInt64(params[0]);
      const std::string &p = params[1];
      const std::string &min = params[2];
      const std::string &max = params[3];
      std::string value = p + " " + min + " " + max;
      keys[key] = value;
      int64_t removeKey = RemoveColor(key);
      GenTransOptKeyData &optKeyData = optkeys[removeKey];
      optKeyData.ps[value]++;

      g_totalKey++;

      PrintProgress("step1", g_totalKey, g_total);
    }

    in.close();

    for (auto &e : optkeys)
    {
      GenTransOptKeyData &optKeyData = e.second;
      int max = 0;
      for (const auto &e1 : optKeyData.ps)
      {
        if (e1.second > max)
        {
          optKeyData.max = e1.first;
          max = e1.second;
        }
      }
    }

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();

    // std::map for deterministic row order; Java's HashMap order is an
    // implementation artifact
    std::map<int64_t, std::string> orderedKeys(keys.begin(), keys.end());
    for (const auto &e : orderedKeys)
    {
      int64_t key = e.first;
      const std::string &p = e.second;
      int64_t removeKey = RemoveColor(key);
      auto it = optkeys.find(removeKey);
      if (it == optkeys.end() || it->second.max != p)
      {
        std::string tmp = std::to_string(key) + " 0 " + p + " " + ToString(key) + "\n";
        out << tmp;
      }

      g_totalKey++;

      PrintProgress("step2", g_totalKey, g_total);
    }

    std::map<int64_t, GenTransOptKeyData> orderedOpt(optkeys.begin(), optkeys.end());
    for (const auto &e : orderedOpt)
    {
      int64_t key = e.first;
      const GenTransOptKeyData &optKeyData = e.second;
      std::string tmp = std::to_string(key) + " 1 " + optKeyData.max + " " + ToString(key) + "\n";
      out << tmp;
    }

    out.close();

    std::cout << "optData finish " << g_totalKey << std::endl;
  }

} // namespace texas_algorithm
