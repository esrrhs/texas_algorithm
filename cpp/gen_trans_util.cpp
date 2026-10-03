#include "gen_trans_util.h"

#include <algorithm>
#include <ctime>
#include <fstream>
#include <iostream>
#include <map>
#include <unordered_map>

#include "gen_util.h"

namespace texas_algorithm
{

  int GenTransN = 6;

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

    // GenTransKeyData accumulates the win statistics of one N-card key.
    struct GenTransKeyData
    {
      int64_t win = 0;
      int64_t num = 0;
      double min = 1;
      double max = 0;
    };

    int64_t g_total = 1;
    int64_t g_totalKey = 0;
    int g_lastPrint = 0;
    int64_t g_beginPrint = 0;
    std::unordered_map<int64_t, GenTransKeyData> g_keys;

    void PrintProgress(const char *step, int64_t processed, int64_t total)
    {
      int cur = (int)(processed * 100 / total);
      if (cur != g_lastPrint)
      {
        g_lastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_beginPrint) / (double)processed;
        std::cout << step << " N" << GenTransN << " " << cur << "% 需要"
                  << per * (double)(total - processed) / 60 / 1000 << "分"
                  << " 用时" << (now - g_beginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)processed / ((double)(now - g_beginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    void GenTransCardSave(const std::vector<int> &tmp)
    {
      int64_t c = GenCardBindInts(tmp);

      g_keys[c] = GenTransKeyData();
      g_totalKey++;

      PrintProgress("", g_totalKey, g_total);
    }

    void GenTransCard()
    {
      std::vector<int> list = GenAllCards();

      std::vector<int> tmp(GenTransN);
      Permutation(list, 0, 0, GenTransN, tmp, [](std::vector<int> &tmp)
                  { GenTransCardSave(tmp); });
    }

  } // namespace

  void GenTransGenKey()
  {
    g_total = 1;
    for (int i = 0; i < GenTransN; i++)
    {
      g_total *= GenNum - i;
    }
    for (int i = GenTransN; i >= 1; i--)
    {
      g_total /= i;
    }
    g_beginPrint = NowMillis();
    g_keys.clear();
    g_totalKey = 0;
    g_lastPrint = 0;

    GenTransCard();

    std::cout << "genKey finish " << g_total << std::endl;
  }

  void GenTransTransData()
  {
    std::ifstream in("texas_data.txt", std::ios::binary);
    if (!in.is_open())
    {
      std::cout << "cannot open texas_data.txt" << std::endl;
      return;
    }

    std::string outName = "texas_data_" + std::to_string(GenTransN) + ".txt";
    std::ofstream out(outName, std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open " << outName << std::endl;
      return;
    }

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();

    std::string line;
    while (std::getline(in, line))
    {
      std::vector<std::string> fields = SplitLine(line);
      int64_t key = std::strtoll(fields[0].c_str(), nullptr, 10);
      int64_t win = std::strtoll(fields[2].c_str(), nullptr, 10);
      std::vector<int64_t> tmp = GenTransGetKeyList(key);
      for (int64_t l : tmp)
      {
        auto it = g_keys.find(l);
        if (it == g_keys.end())
        {
          continue;
        }
        it->second.win += win;
        it->second.num++;
        double p = (double)win / (double)Total;
        if (p < it->second.min)
        {
          it->second.min = p;
        }
        if (p > it->second.max)
        {
          it->second.max = p;
        }
      }

      g_totalKey++;

      PrintProgress("step1", g_totalKey, Total);
    }

    in.close();

    g_totalKey = 0;
    g_lastPrint = 0;
    g_beginPrint = NowMillis();

    // use std::map for deterministic iteration; Java's HashMap order is an
    // implementation artifact and the output rows are independent
    std::map<int64_t, GenTransKeyData> ordered(g_keys.begin(), g_keys.end());
    for (const auto &e : ordered)
    {
      int64_t key = e.first;
      double win = (double)e.second.win / (double)e.second.num / (double)Total;

      std::string tmp = std::to_string(key) + " " + FormatDouble(win) + " " + FormatDouble(e.second.min) + " " +
                        FormatDouble(e.second.max) + " " + ToString(key) + "\n";
      out << tmp;
      g_totalKey++;

      PrintProgress("step2", g_totalKey, Total);
    }

    out.close();
    g_keys.clear();

    std::cout << "transData finish " << g_totalKey << std::endl;
  }

  std::vector<int64_t> GenTransGetKeyList(int64_t key)
  {
    std::vector<int64_t> ret;
    std::vector<int> list;
    while (key > 100)
    {
      list.push_back((int)(key % 100));
      key /= 100;
    }
    list.push_back((int)key);
    std::sort(list.begin(), list.end());

    std::vector<int> tmp(GenTransN);
    Permutation(list, 0, 0, GenTransN, tmp, [&](std::vector<int> &tmp)
                {
      int64_t c = GenCardBindInts(tmp);
      bool found = false;
      for (int64_t r : ret)
      {
        if (r == c)
        {
          found = true;
          break;
        }
      }
      if (!found)
      {
        ret.push_back(c);
      } });

    return ret;
  }

} // namespace texas_algorithm
