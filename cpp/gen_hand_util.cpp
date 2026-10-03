#include "gen_hand_util.h"

#include <algorithm>
#include <atomic>
#include <condition_variable>
#include <ctime>
#include <filesystem>
#include <fstream>
#include <iostream>
#include <memory>
#include <mutex>
#include <thread>

#include "gen_util.h"
#include "texas_algorithm_util.h"

namespace texas_algorithm
{

  int GenHandN = 6;

  namespace
  {

    int64_t NowMillis()
    {
      struct timespec ts;
      clock_gettime(CLOCK_REALTIME, &ts);
      return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
    }

    int64_t g_totalHand = 0;
    int64_t g_totalPub = 1;
    int64_t g_total = 1;
    int64_t g_totalCalc = 1;
    int64_t g_totalKey = 0;
    int g_lastPrint = 0;
    int64_t g_beginPrint = 0;

    std::mutex g_ioMu; // protects the per-hand output streams and progress counters

    struct CalcData
    {
      int64_t win = 0;
      int64_t lose = 0;
      int64_t tie = 0;
      int64_t total = 0;
    };

    // removeCounts removes one occurrence of each value in remove from list,
    // matching Java's repeated list.remove((Integer)p) semantics without
    // mutating the input.
    std::vector<int> RemoveCounts(const std::vector<int> &list, const std::vector<int> &remove)
    {
      std::vector<int> cnt(256, 0);
      for (int p : remove)
      {
        cnt[p]++;
      }
      std::vector<int> ret;
      ret.reserve(list.size());
      for (int x : list)
      {
        if (cnt[x] > 0)
        {
          cnt[x]--;
          continue;
        }
        ret.push_back(x);
      }
      return ret;
    }

    void GenHandCalcCompare(CalcData &data, const std::vector<int> &hand, const std::vector<int> &pub,
                            const std::vector<int> &leftpub, const std::vector<int> &otherhand)
    {
      std::vector<int> my;
      std::vector<int> other;
      my.reserve(7);
      other.reserve(7);
      for (int i : hand)
      {
        my.push_back(i);
      }
      for (int i : pub)
      {
        my.push_back(i);
      }
      for (int i : leftpub)
      {
        my.push_back(i);
      }

      for (int i : otherhand)
      {
        other.push_back(i);
      }
      for (int i : pub)
      {
        other.push_back(i);
      }
      for (int i : leftpub)
      {
        other.push_back(i);
      }

      int ret = CompareBytes(my, other);
      if (ret == 0)
      {
        data.tie++;
      }
      else if (ret < 0)
      {
        data.lose++;
      }
      else if (ret > 0)
      {
        data.win++;
      }
      data.total++;
    }

    // GenHandCalc exhaustively enumerates every opponent hand and the remaining
    // public cards, comparing each against the player's hand.
    double GenHandCalc(const std::vector<int> &hand, const std::vector<int> &pub)
    {
      std::vector<int> list = RemoveCounts(GenAllCards(), hand);
      list = RemoveCounts(list, pub);
      std::sort(list.begin(), list.end());

      CalcData data;

      std::vector<int> otherhand(2);
      Permutation(list, 0, 0, 2, otherhand, [&](std::vector<int> &otherhand)
                  {
        std::vector<int> remaining = RemoveCounts(list, otherhand);

        std::vector<int> leftpub(5 - GenHandN);
        Permutation(remaining, 0, 0, 5 - GenHandN, leftpub, [&](std::vector<int> &leftpub)
                    { GenHandCalcCompare(data, hand, pub, leftpub, otherhand); }); });

      return ((double)data.win + (double)data.tie * 0.5) / (double)data.total;
    }

    void GenHandCardSave(std::ostream &out, const std::vector<int> &hand, const std::vector<int> &pub)
    {
      int64_t h = GenCardBindInts(hand);
      int64_t p = GenCardBindInts(pub);
      for (int i = 0; i < GenHandN; i++)
      {
        h *= 100;
      }
      h += p;

      double pp = GenHandCalc(hand, pub);

      std::string tmp = std::to_string(h) + " " + FormatDouble(pp) + " " + ToString(h) + "\n";

      std::lock_guard<std::mutex> lock(g_ioMu);
      out << tmp;
      g_totalKey++;
      int cur = (int)(g_totalKey * 10000 / g_total);
      if (cur != g_lastPrint)
      {
        g_lastPrint = cur;

        int64_t now = NowMillis();
        double per = (double)(now - g_beginPrint) / (double)g_totalKey;
        std::cout << "N" << GenHandN << " " << cur << "%% 需要" << per * (double)(g_total - g_totalKey) / 60 / 1000
                  << "分"
                  << " 用时" << (now - g_beginPrint) / 60 / 1000 << "分"
                  << " 速度" << (double)g_totalKey / ((double)(now - g_beginPrint) / 1000) << "条/秒" << std::endl;
      }
    }

    struct Task
    {
      std::shared_ptr<std::ofstream> out;
      std::vector<int> hand;
      std::vector<int> pub;
    };

  } // namespace

  void GenHandGenKey()
  {
    g_totalPub = 1;
    for (int i = 0; i < GenHandN; i++)
    {
      g_totalPub *= GenNum - 2 - i;
    }
    for (int i = GenHandN; i >= 1; i--)
    {
      g_totalPub /= i;
    }
    g_totalHand = (GenNum * (GenNum - 1)) / 2;
    g_total = g_totalPub * g_totalHand;

    g_totalCalc = 1;
    for (int i = 0; i < 5 - GenHandN; i++)
    {
      g_totalCalc *= GenNum - 2 - GenHandN - i;
    }
    for (int i = 5 - GenHandN; i >= 1; i--)
    {
      g_totalCalc /= i;
    }
    for (int i = 0; i < 2; i++)
    {
      g_totalCalc *= GenNum - 7 - i;
    }
    for (int i = 2; i >= 1; i--)
    {
      g_totalCalc /= i;
    }

    Load();

    std::string dirName = "hand" + std::to_string(GenHandN);
    std::filesystem::create_directories(dirName);

    g_beginPrint = NowMillis();
    g_totalKey = 0;
    g_lastPrint = 0;

    int nThreads = (int)std::thread::hardware_concurrency();
    if (nThreads <= 0)
    {
      nThreads = 4;
    }
    int fallback = 2 * nThreads;

    // Scheduler mirroring Java's GenHandUtil: the enumerating thread runs a pub
    // task inline when enough tasks are already in flight, otherwise it queues
    // it for the worker pool.
    std::mutex schedMu;
    std::condition_variable schedCv;
    std::vector<Task> tasks;
    int inFlight = 0;
    bool enumerationDone = false;

    auto processTask = [](const Task &task)
    { GenHandCardSave(*task.out, task.hand, task.pub); };

    std::vector<std::thread> workers;
    for (int w = 0; w < nThreads; w++)
    {
      workers.emplace_back([&]
                           {
        for (;;)
        {
          Task task;
          {
            std::unique_lock<std::mutex> lock(schedMu);
            schedCv.wait(lock, [&]
                         { return !tasks.empty() || enumerationDone; });
            if (tasks.empty())
            {
              return;
            }
            task = std::move(tasks.front());
            tasks.erase(tasks.begin());
          }
          processTask(task);
          {
            std::lock_guard<std::mutex> lock(schedMu);
            inFlight--;
          }
          schedCv.notify_all();
        } });
    }

    auto submitOrInline = [&](std::shared_ptr<std::ofstream> out, const std::vector<int> &hand,
                              const std::vector<int> &pub)
    {
      bool spawn;
      {
        std::lock_guard<std::mutex> lock(schedMu);
        spawn = inFlight < fallback;
        if (spawn)
        {
          Task task;
          task.out = out;
          task.hand = hand;
          task.pub = pub;
          tasks.push_back(std::move(task));
          inFlight++;
        }
      }
      if (spawn)
      {
        schedCv.notify_one();
      }
      else
      {
        Task task;
        task.out = out;
        task.hand = hand;
        task.pub = pub;
        processTask(task);
      }
    };

    std::vector<int> list = GenAllCards();

    std::vector<int> hand(2);
    Permutation(list, 0, 0, 2, hand, [&](std::vector<int> &hand)
                {
      std::string path = dirName + "/texas_hand_" + ToString((int64_t)hand[0] * 100 + hand[1]) + ".txt";
      if (FileExists(path))
      {
        return;
      }

      // create an empty file; tasks append to it
      {
        std::ofstream create(path, std::ios::binary | std::ios::trunc);
      }
      auto out = std::make_shared<std::ofstream>(path, std::ios::binary | std::ios::app);

      std::vector<int> list1 = RemoveCounts(GenAllCards(), hand);

      std::vector<int> pub(GenHandN);
      Permutation(list1, 0, 0, GenHandN, pub, [&](std::vector<int> &pub)
                  { submitOrInline(out, hand, pub); }); });

    {
      std::lock_guard<std::mutex> lock(schedMu);
      enumerationDone = true;
    }
    schedCv.notify_all();

    for (std::thread &t : workers)
    {
      t.join();
    }

    std::cout << "genKey finish " << g_total << std::endl;
  }

} // namespace texas_algorithm
