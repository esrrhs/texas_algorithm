#include "sorter.h"

#include <cmath>
#include <condition_variable>
#include <functional>
#include <iostream>
#include <mutex>
#include <queue>
#include <thread>

#include "gen_util.h"

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

    // Minimal fixed-size thread pool, mirroring Java's Executors.newFixedThreadPool.
    class ThreadPool
    {
    public:
      explicit ThreadPool(int n)
      {
        for (int i = 0; i < n; i++)
        {
          workers_.emplace_back([this]
                                {
          for (;;)
          {
            std::function<void()> task;
            {
              std::unique_lock<std::mutex> lock(mu_);
              cv_.wait(lock, [this]
                       { return stop_ || !tasks_.empty(); });
              if (stop_ && tasks_.empty())
              {
                return;
              }
              task = std::move(tasks_.front());
              tasks_.pop();
            }
            task();
          } });
        }
      }

      ~ThreadPool()
      {
        {
          std::lock_guard<std::mutex> lock(mu_);
          stop_ = true;
        }
        cv_.notify_all();
        for (std::thread &t : workers_)
        {
          t.join();
        }
      }

      void Submit(std::function<void()> task)
      {
        {
          std::lock_guard<std::mutex> lock(mu_);
          tasks_.push(std::move(task));
        }
        cv_.notify_one();
      }

    private:
      std::vector<std::thread> workers_;
      std::queue<std::function<void()>> tasks_;
      std::mutex mu_;
      std::condition_variable cv_;
      bool stop_ = false;
    };

    ThreadPool &SortPool()
    {
      static ThreadPool pool((int)std::thread::hardware_concurrency());
      return pool;
    }

  } // namespace

  void Quicksort(std::vector<int64_t> &input)
  {
    if (input.size() <= 1)
    {
      return;
    }

    int nThreads = (int)std::thread::hardware_concurrency();
    if (nThreads <= 0)
    {
      nThreads = 4;
    }
    int fallback = 2 * nThreads;
    int64_t total = (int64_t)input.size();

    ThreadPool &pool = SortPool();

    std::mutex mu;
    std::condition_variable cv;
    std::atomic<int> count{1}; // number of tasks not yet finished, including the root
    bool done = false;

    auto finishTask = [&]()
    {
      bool finished;
      {
        std::lock_guard<std::mutex> lock(mu);
        finished = count.fetch_sub(1) == 1;
        if (finished)
        {
          done = true;
        }
      }
      if (finished)
      {
        cv.notify_one();
      }
    };

    std::function<void(int, int, int)> quicksort = [&](int layer, int lowerIndex, int higherIndex)
    {
      if (higherIndex < lowerIndex)
      {
        return;
      }
      if (higherIndex == lowerIndex)
      {
        Progress.fetch_add(1);

        int64_t step = Progress.load();
        int64_t cur = step * 10000 / total;
        if (cur != LastPrint.load())
        {
          LastPrint.store(cur);

          int64_t now = NowMillis();
          double per = (double)(now - BeginPrint.load()) / (double)step;
          std::cout << cur << "%% 需要" << per * (double)(total - step) / 60 / 1000 << "分"
                    << " 用时" << (now - BeginPrint.load()) / 60 / 1000 << "分"
                    << " 速度" << (double)step / ((double)(now - BeginPrint.load()) / 1000) << "条/秒"
                    << std::endl;
        }
        return;
      }

      int i = lowerIndex;
      int j = higherIndex;
      // calculate pivot number, I am taking pivot as middle index number
      int64_t pivot = input[lowerIndex + (higherIndex - lowerIndex) / 2];

      int totalStep = j - i;
      if (totalStep == 0)
      {
        totalStep = 1;
      }
      int lastPrint = 0;
      int64_t beginPrint = NowMillis();

      while (i <= j)
      {
        // In each iteration, find a number from the left side greater than the
        // pivot, and a number from the right side less than the pivot, then
        // exchange them.
        while (GenCompare(input[i], pivot))
        {
          i++;
        }
        while (GenCompare(pivot, input[j]))
        {
          j--;
        }
        if (i <= j)
        {
          std::swap(input[i], input[j]);
          // move index to next position on both sides
          i++;
          j--;
        }

        if (i <= j && totalStep > 100000)
        {
          int step = totalStep - (j - i);
          if (step < 0)
          {
            step = 0;
          }
          int cur = step * 100 / totalStep;
          if (cur != lastPrint)
          {
            lastPrint = cur;

            int64_t now = NowMillis();
            double per = (double)(now - beginPrint) / (double)step;
            std::cout << layer << "/" << (int)std::log((double)total) << "层 " << cur << "% 需要"
                      << per * (double)(totalStep - step) / 60 / 1000 << "分"
                      << " 用时" << (now - beginPrint) / 60 / 1000 << "分"
                      << " 速度" << (double)step / ((double)(now - beginPrint) / 1000) << "条/秒"
                      << std::endl;
          }
        }
      }

      auto spawn = [&](int lo, int hi)
      {
        count.fetch_add(1);
        pool.Submit([&mu, &cv, &count, &quicksort, &finishTask, layer, lo, hi]()
                    { quicksort(layer + 1, lo, hi);
          finishTask(); });
      };

      // call quicksort recursively, inline or in pool tasks
      if (count.load() >= fallback)
      {
        if (i - j == 1)
        {
          quicksort(layer + 1, lowerIndex, j);
          quicksort(layer + 1, i, higherIndex);
        }
        else
        {
          quicksort(layer + 1, lowerIndex, j + 1);
          quicksort(layer + 1, i, higherIndex);
        }
      }
      else
      {
        if (i - j == 1)
        {
          spawn(lowerIndex, j);
          spawn(i, higherIndex);
        }
        else
        {
          spawn(lowerIndex, j + 1);
          spawn(i, higherIndex);
        }
      }
    };

    BeginPrint.store(NowMillis());
    LastPrint.store(0);

    pool.Submit([&]()
                { quicksort(0, 0, (int)input.size() - 1);
      finishTask(); });

    {
      std::unique_lock<std::mutex> lock(mu);
      cv.wait(lock, [&]
              { return done; });
    }
  }

} // namespace texas_algorithm
