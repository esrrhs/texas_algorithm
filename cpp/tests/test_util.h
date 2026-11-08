#pragma once

// Minimal assertion helpers shared by the test executables, mirroring the
// structure of Java's JUnit tests: each test prints its own PASS / SKIP /
// FAIL line, and the process exits non-zero when any check fails.

#include <cstdio>
#include <stdexcept>
#include <string>

namespace txtest
{

  inline int g_failures = 0;
  inline int g_checks = 0;

  inline void Check(bool cond, const std::string &msg)
  {
    g_checks++;
    if (!cond)
    {
      g_failures++;
      std::printf("    CHECK FAILED: %s\n", msg.c_str());
    }
  }

  inline void CheckEq(long long actual, long long expected, const std::string &msg)
  {
    g_checks++;
    if (actual != expected)
    {
      g_failures++;
      std::printf("    CHECK FAILED: %s (actual=%lld, expected=%lld)\n", msg.c_str(), actual, expected);
    }
  }

  inline void CheckEqStr(const std::string &actual, const std::string &expected, const std::string &msg)
  {
    g_checks++;
    if (actual != expected)
    {
      g_failures++;
      std::printf("    CHECK FAILED: %s (actual=%s, expected=%s)\n", msg.c_str(), actual.c_str(), expected.c_str());
    }
  }

  // CheckThrowsInvalidArgument checks that fn throws std::invalid_argument
  // whose message starts with "无法识别的牌" and contains `contains`. Any other
  // outcome (including no exception) is a failure.
  template <typename Fn>
  inline void CheckThrowsInvalidArgument(Fn fn, const std::string &contains, const std::string &msg)
  {
    g_checks++;
    try
    {
      fn();
      g_failures++;
      std::printf("    CHECK FAILED: %s (no exception thrown)\n", msg.c_str());
    }
    catch (const std::invalid_argument &e)
    {
      std::string what = e.what();
      if (what.rfind("无法识别的牌", 0) != 0 || what.find(contains) == std::string::npos)
      {
        g_failures++;
        std::printf("    CHECK FAILED: %s (unexpected message: %s)\n", msg.c_str(), what.c_str());
      }
    }
    catch (const std::exception &e)
    {
      g_failures++;
      std::printf("    CHECK FAILED: %s (wrong exception: %s)\n", msg.c_str(), e.what());
    }
    catch (...)
    {
      g_failures++;
      std::printf("    CHECK FAILED: %s (wrong exception type)\n", msg.c_str());
    }
  }

  // RunTest wraps one test body, printing its result like JUnit's @DisplayName.
  inline void RunTest(const char *name, bool runnable, void (*fn)())
  {
    if (!runnable)
    {
      std::printf("[ SKIPPED ] %s (prerequisite not met)\n", name);
      return;
    }
    std::printf("[ RUN     ] %s\n", name);
    int before = g_failures;
    fn();
    if (g_failures == before)
    {
      std::printf("[     OK  ] %s\n", name);
    }
    else
    {
      std::printf("[    FAIL ] %s\n", name);
    }
  }

  inline int ExitCode()
  {
    if (g_failures > 0)
    {
      std::printf("[  FAILED ] %d check(s) failed\n", g_failures);
      return 1;
    }
    std::printf("[  PASSED ] %d check(s) passed\n", g_checks);
    return 0;
  }

} // namespace txtest
