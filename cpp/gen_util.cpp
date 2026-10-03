#include "gen_util.h"

#include <algorithm>
#include <cstdio>
#include <ctime>
#include <fstream>
#include <iostream>
#include <sys/stat.h>

#include "texas_algorithm_util.h"
#include "texas_card_util.h"
#include "sorter.h"

namespace texas_algorithm
{

  int64_t Total = (GenNum * (GenNum - 1) * (GenNum - 2) * (GenNum - 3) * (GenNum - 4) * (GenNum - 5) * (GenNum - 6)) /
                  (7 * 6 * 5 * 4 * 3 * 2);

  bool UseOpt = false;
  std::vector<int64_t> Keys;
  std::atomic<int64_t> Progress{0};
  std::atomic<int64_t> LastPrint{0};
  std::atomic<int64_t> BeginPrint{0};

  namespace
  {

    int64_t NowMillis()
    {
      struct timespec ts;
      clock_gettime(CLOCK_REALTIME, &ts);
      return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
    }

    // Package-level generation state mirroring Java's GenUtil static fields.
    int64_t g_genTotalKey = 0;
    int g_genLastPrint = 0;
    int64_t g_genBeginPrint = 0;

    void GenCardSave(const std::vector<int> &tmp)
    {
      int64_t c = GenCardBindInts(tmp);

      Keys.push_back(c);
      g_genTotalKey++;

      int cur = (int)(g_genTotalKey * 100 / Total);
      if (cur != g_genLastPrint)
      {
        g_genLastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_genBeginPrint) / (double)g_genTotalKey;
        std::cout << cur << "% 需要" << per * (double)(Total - g_genTotalKey) / 60 / 1000 << "分"
                  << " 用时" << (now - g_genBeginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)g_genTotalKey / ((double)(now - g_genBeginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    void GenCard()
    {
      std::vector<int> list = GenAllCards();

      std::vector<int> tmp(7);
      Permutation(list, 0, 0, 7, tmp, [](std::vector<int> &tmp)
                  { GenCardSave(tmp); });
    }

  } // namespace

  std::vector<int> GenAllCards()
  {
    std::vector<int> list;
    for (int i = 0; i < 4; i++)
    {
      for (int j = 0; j < (int)(GenNum / 4); j++)
      {
        list.push_back(PokeToByte(MakePoke(i, j + 2)));
      }
    }
    for (int i = 0; i < GuiNum; i++)
    {
      list.push_back(PokeToByte(GUI));
    }
    std::sort(list.begin(), list.end());
    return list;
  }

  // AllCards is initialized after GenAllCards is defined (static init order is
  // safe here because GenAllCards has no dependencies).
  std::vector<int> AllCards = GenAllCards();

  std::vector<Poke> GenAllPokes()
  {
    std::vector<Poke> list;
    for (int t : GenAllCards())
    {
      list.push_back(PokeFromByte(t));
    }
    return list;
  }

  void GenKey()
  {
    g_genBeginPrint = NowMillis();

    GenCard();

    std::cout << "genKey finish " << Total << std::endl;
  }

  void Permutation(const std::vector<int> &a, int count, int count2, int except, std::vector<int> &tmp,
                   const std::function<void(std::vector<int> &)> &run)
  {
    if (count2 == except)
    {
      run(tmp);
    }
    else
    {
      for (int i = count; i < (int)a.size(); i++)
      {
        tmp[count2] = a[i];
        Permutation(a, i + 1, count2 + 1, except, tmp, run);
      }
    }
  }

  int64_t GenCardBindInts(const std::vector<int> &tmp)
  {
    int64_t ret = 0;
    for (int i : tmp)
    {
      ret = ret * 100 + i;
    }
    return ret;
  }

  int64_t GenCardBindBytes(const std::vector<int> &tmp)
  {
    int64_t ret = 0;
    for (int i : tmp)
    {
      ret = ret * 100 + i;
    }
    return ret;
  }

  void OutputData()
  {
    int64_t begin = NowMillis();

    if (UseOpt)
    {
      Load();
    }

    std::ofstream out("texas_data.txt", std::ios::binary | std::ios::trunc);
    if (!out.is_open())
    {
      std::cout << "cannot open texas_data.txt" << std::endl;
      return;
    }

    g_genBeginPrint = NowMillis();
    g_genLastPrint = 0;

    Quicksort(Keys);

    g_genTotalKey = 0;
    g_genLastPrint = 0;
    g_genBeginPrint = NowMillis();
    int i = 0;
    int64_t iindex = 0;
    int64_t index = 0;
    int64_t lastMax = 0;
    int64_t size = (int64_t)Keys.size();
    for (int64_t k : Keys)
    {
      int64_t curMax = Max(Keys[index]);
      std::string str;
      auto row = [&]()
      {
        return std::to_string(k) + " " + std::to_string(i) + " " + std::to_string(iindex) + " " +
               std::to_string(size) + " " + ToString(Keys[index]) + " " + std::to_string(curMax) + " " +
               ToString(curMax) + " " + std::to_string(MaxType(Keys[index])) + "\n";
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

      int cur = (int)(index * 100 / Total);
      if (cur != g_genLastPrint)
      {
        g_genLastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_genBeginPrint) / (double)index;
        std::cout << cur << "% 需要" << per * (double)(Total - index) / 60 / 1000 << "分"
                  << " 用时" << (now - g_genBeginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)index / ((double)(now - g_genBeginPrint) / 1000) << "条/秒"
                  << std::endl;
      }
    }

    out.close();

    std::cout << "outputData finish " << Total << " time:" << (NowMillis() - begin) / 1000 / 60 << "分" << std::endl;

    Keys.clear();
  }

  namespace
  {

    std::vector<Poke> PickBest(const std::vector<Poke> &cs)
    {
      std::vector<Poke> pickedCards;
      if (cs.size() == 7)
      {
        pickedCards = FiveFromSeven(cs);
      }
      else if (cs.size() == 6)
      {
        pickedCards = FiveFromSix(cs);
      }
      else if (cs.size() == 5)
      {
        pickedCards = FiveFromFive(cs);
      }
      return pickedCards;
    }

  } // namespace

  int64_t Max(int64_t k)
  {
    std::vector<Poke> cs = ToArray(k);
    std::vector<Poke> pickedCards = PickBest(cs);

    int64_t ret = 0;
    for (const Poke &p : pickedCards)
    {
      ret = ret * 100 + PokeToByte(p);
    }
    return ret;
  }

  int MaxType(int64_t k)
  {
    std::vector<Poke> cs = ToArray(k);
    std::vector<Poke> pickedCards = PickBest(cs);

    return GetCardTypeUnorderedWithoutGui(pickedCards);
  }

  std::vector<Poke> ToArray(int64_t k)
  {
    std::vector<Poke> cs;
    if (k > 1000000000000LL)
    {
      cs.push_back(PokeFromByte((int)(k % 100000000000000LL / 1000000000000LL)));
    }
    if (k > 10000000000LL)
    {
      cs.push_back(PokeFromByte((int)(k % 1000000000000LL / 10000000000LL)));
    }
    if (k > 100000000LL)
    {
      cs.push_back(PokeFromByte((int)(k % 10000000000LL / 100000000LL)));
    }
    if (k > 1000000LL)
    {
      cs.push_back(PokeFromByte((int)(k % 100000000LL / 1000000LL)));
    }
    if (k > 10000LL)
    {
      cs.push_back(PokeFromByte((int)(k % 1000000LL / 10000LL)));
    }
    if (k > 100LL)
    {
      cs.push_back(PokeFromByte((int)(k % 10000LL / 100LL)));
    }
    if (k > 1LL)
    {
      cs.push_back(PokeFromByte((int)(k % 100LL / 1LL)));
    }
    return cs;
  }

  std::string ToString(int64_t k)
  {
    std::vector<Poke> cs = ToArray(k);
    std::string ret;
    for (const Poke &poke : cs)
    {
      ret += PokeToString(poke);
    }
    return ret;
  }

  bool GenCompare(int64_t k1, int64_t k2)
  {
    if (UseOpt)
    {
      return CompareKey(k1, k2) < 0;
    }

    std::vector<Poke> pickedCards1 = PickBest(ToArray(k1));
    std::vector<Poke> pickedCards2 = PickBest(ToArray(k2));

    return CompareCardsWithoutGui(pickedCards1, pickedCards2) < 0;
  }

  bool Equal(int64_t k1, int64_t k2)
  {
    std::vector<Poke> cs1{
        PokeFromByte((int)(k1 % 10000000000LL / 100000000LL)),
        PokeFromByte((int)(k1 % 100000000LL / 1000000LL)),
        PokeFromByte((int)(k1 % 1000000LL / 10000LL)),
        PokeFromByte((int)(k1 % 10000LL / 100LL)),
        PokeFromByte((int)(k1 % 100LL / 1LL)),
    };
    SortPokesDesc(cs1);

    std::vector<Poke> cs2{
        PokeFromByte((int)(k2 % 10000000000LL / 100000000LL)),
        PokeFromByte((int)(k2 % 100000000LL / 1000000LL)),
        PokeFromByte((int)(k2 % 1000000LL / 10000LL)),
        PokeFromByte((int)(k2 % 10000LL / 100LL)),
        PokeFromByte((int)(k2 % 100LL / 1LL)),
    };
    SortPokesDesc(cs2);

    return CompareCardsWithoutGui(cs1, cs2) == 0;
  }

  bool FileExists(const std::string &path)
  {
    struct stat st;
    return ::stat(path.c_str(), &st) == 0;
  }

  std::string FormatDouble(double v)
  {
    char buf[64];
    std::snprintf(buf, sizeof(buf), "%.17g", v);
    return std::string(buf);
  }

} // namespace texas_algorithm
