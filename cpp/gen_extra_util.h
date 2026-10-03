#pragma once

#include <cstdint>
#include <vector>

namespace texas_algorithm
{

  // GenExtraN selects the card count for the extra table generation (5 or 6).
  // Set it before generating, mirroring Java's static GenExtraUtil.N.
  extern int GenExtraN;

  // GenExtraGenKey enumerates all C(54, GenExtraN) combinations.
  void GenExtraGenKey();

  // GenExtraOutputData sorts the keys by poker strength and writes
  // texas_data_extra_N.txt.
  void GenExtraOutputData();

} // namespace texas_algorithm
