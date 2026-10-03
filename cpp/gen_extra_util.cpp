#include "gen_extra_util.h"

#include <atomic>
#include <ctime>
#include <fstream>
#include <iostream>

#include "gen_util.h"
#include "sorter.h"
#include "texas_algorithm_util.h"
#include "texas_card_util.h"

namespace texas_algorithm
{

  int GenExtraN = 6;

  namespace
  {

    int64_t NowMillis()
    {
      struct timespec ts;
      clock_gettime(CLOCK_REALTIME, &ts);
      return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
    }

    // Package-level generation state mirroring Java's GenExtraUtil static fields.
    int64_t g_extraTotal = 0;
    int64_t g_extraTotalKey = 0;
    int g_extraLastPrint = 0;
    int64_t g_extraBeginPrint = 0;
    std::vector<int64_t> g_extraKeys;
    std::atomic<int64_t> g_extraProgress{0};

    void GenExtraCardSave(const std::vector<int> &tmp)
    {
      int64_t c = GenCardBindInts(tmp);

      g_extraKeys.push_back(c);
      g_extraTotalKey++;

      int cur = (int)(g_extraTotalKey * 100 / g_extraTotal);
      if (cur != g_extraLastPrint)
      {
        g_extraLastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_extraBeginPrint) / (double)g_extraTotalKey;
        std::cout << cur << "% 需要" << per * (double)(g_extraTotal - g_extraTotalKey) / 60 / 1000 << "分"
                  << " 用时" << (now - g_extraBeginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)g_extraTotalKey / ((double)(now - g_extraBeginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    void GenExtraCard()
    {
      std::vector<int> list = GenAllCards();

      std::vector<int> tmp(GenExtraN);
      Permutation(list, 0, 0, GenExtraN, tmp, [](std::vector<int> &tmp)
                  { GenExtraCardSave(tmp); });
    }

  } // namespace

  void GenExtraGenKey()
  {
    g_extraTotal = 1;
    for (int i = 0; i < GenExtraN; i++)
    {
      g_extraTotal *= GenNum - i;
    }
    for (int i = GenExtraN; i >= 1; i--)
    {
      g_extraTotal /= i;
    }
    g_extraBeginPrint = NowMillis();

    GenExtraCard();

    std::cout << "genKey finish " << g_extraTotal << std::endl;
  }

  void GenExtraOutputData()
  {
    int64_t begin = NowMillis();

    if (UseOpt)
    {
      Load();
    }

    std::string name = "texas_data_extra_" + std::to_string(GenExtraN) + ".txt";
    std::ofstream out(name, std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open " << name << std::endl;
      return;
    }

    g_extraBeginPrint = NowMillis();
    g_extraLastPrint = 0;

    Quicksort(g_extraKeys);

    g_extraTotalKey = 0;
    g_extraLastPrint = 0;
    g_extraBeginPrint = NowMillis();
    int i = 0;
    int64_t iindex = 0;
    int64_t index = 0;
    int64_t lastMax = 0;
    int64_t size = (int64_t)g_extraKeys.size();
    for (int64_t k : g_extraKeys)
    {
      int64_t curMax = Max(g_extraKeys[index]);
      std::string str;
      auto row = [&]()
      {
        return std::to_string(k) + " " + std::to_string(i) + " " + std::to_string(iindex) + " " +
               std::to_string(size) + " " + ToString(g_extraKeys[index]) + " " + std::to_string(curMax) + " " +
               ToString(curMax) + " " + std::to_string(MaxType(g_extraKeys[index])) + "\n";
      };
      if (lastMax == 0)
      {
        str = row();
        lastMax = curMax;
        iindex = index;
      }
      else
      {
        if (Equal(lastMax, curMax))
        {
          str = row();
          lastMax = curMax;
        }
        else
        {
          i++;
          iindex = index;
          str = row();
          lastMax = curMax;
        }
      }

      out << str;
      index++;

      int cur = (int)(index * 100 / g_extraTotal);
      if (cur != g_extraLastPrint)
      {
        g_extraLastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_extraBeginPrint) / (double)index;
        std::cout << cur << "% 需要" << per * (double)(g_extraTotal - index) / 60 / 1000 << "分"
                  << " 用时" << (now - g_extraBeginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)index / ((double)(now - g_extraBeginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    out.close();

    std::cout << "outputData finish " << g_extraTotal << " time:" << (NowMillis() - begin) / 1000 / 60 << "分 "
              << g_extraProgress.load() << std::endl;

    g_extraKeys.clear();
  }

} // namespace texas_algorithm
