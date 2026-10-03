#pragma once

#include <atomic>
#include <cstdint>
#include <functional>
#include <string>
#include <vector>

#include "poke.h"

namespace texas_algorithm
{

  // Number of wild cards in the deck, aligned with Java's GenUtil.guiNum.
  constexpr int GuiNum = 2;

  // GenNum is the total number of cards in the deck (52 + 2 jokers).
  constexpr int64_t GenNum = 52 + GuiNum;

  // Total is the number of 7-card combinations: C(54, 7).
  extern int64_t Total;

  // UseOpt switches GenCompare to the lookup-table comparator once the
  // tables are loaded, mirroring Java's GenUtil.useOpt.
  extern bool UseOpt;

  // Keys accumulates the generated 7-card keys.
  extern std::vector<int64_t> Keys;

  // Progress / LastPrint / BeginPrint track quicksort progress, updated from
  // worker threads (atomics), mirroring Java's GenUtil fields.
  extern std::atomic<int64_t> Progress;
  extern std::atomic<int64_t> LastPrint;
  extern std::atomic<int64_t> BeginPrint;

  // GenKey enumerates all C(54, 7) combinations into Keys.
  void GenKey();

  // OutputData sorts Keys by poker strength and writes texas_data.txt.
  void OutputData();

  // GenAllCards returns all 54 cards as encoded bytes, sorted ascending.
  std::vector<int> GenAllCards();

  // GenAllPokes returns all 54 cards (52 regular cards + 2 jokers).
  std::vector<Poke> GenAllPokes();

  // AllCards is the full 54-card deck as encoded bytes, sorted ascending,
  // mirroring Java's GenUtil.allCards.
  extern std::vector<int> AllCards;

  // Permutation generates combinations of `except` elements from `a` with
  // strictly increasing indices, calling run for each combination,
  // mirroring Java's GenUtil.permutation.
  void Permutation(const std::vector<int> &a, int count, int count2, int except, std::vector<int> &tmp,
                   const std::function<void(std::vector<int> &)> &run);

  // GenCardBindInts / GenCardBindBytes pack card bytes into a key:
  // ret = ret * 100 + card.
  int64_t GenCardBindInts(const std::vector<int> &tmp);
  int64_t GenCardBindBytes(const std::vector<int> &tmp);

  // Max returns the encoded key of the best 5-card hand within key k.
  int64_t Max(int64_t k);

  // MaxType returns the card type of the best 5-card hand within key k.
  int MaxType(int64_t k);

  // ToArray decodes a key back into its card list, from the most significant
  // card down, mirroring Java's GenUtil.toArray.
  std::vector<Poke> ToArray(int64_t k);

  // ToString renders a key as its concatenated card strings, e.g. "方A黑K".
  std::string ToString(int64_t k);

  // GenCompare reports whether key k1 is weaker than key k2 by poker strength,
  // mirroring Java's GenUtil.compare(long, long). When UseOpt is set and the
  // tables are loaded, the lookup table is used as the comparator.
  bool GenCompare(int64_t k1, int64_t k2);

  // Equal reports whether the two 5-card best-hand keys are equal in strength,
  // mirroring Java's GenUtil.equal.
  bool Equal(int64_t k1, int64_t k2);

  // FileExists reports whether the path exists.
  bool FileExists(const std::string &path);

  // FormatDouble renders a double with round-trip precision, mirroring Java's
  // Double.toString role in the generated data files.
  std::string FormatDouble(double v);

} // namespace texas_algorithm
