#pragma once

#include <cstdint>
#include <vector>

namespace texas_algorithm
{

  // GenTransN selects the card count for the win-rate table generation (2-6).
  // Set it before generating, mirroring Java's static GenTransUtil.N.
  extern int GenTransN;

  // GenTransGenKey enumerates all C(54, GenTransN) combinations as the key set.
  void GenTransGenKey();

  // GenTransTransData reads texas_data.txt and, for every N-card subset of each
  // 7-card key, accumulates the average win rate into texas_data_N.txt.
  void GenTransTransData();

  // GenTransGetKeyList returns the distinct N-card subset keys contained in the
  // given 7-card key, mirroring Java's GenTransUtil.getKeyList.
  std::vector<int64_t> GenTransGetKeyList(int64_t key);

} // namespace texas_algorithm
