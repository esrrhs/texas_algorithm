#pragma once

#include <cstdint>

namespace texas_algorithm
{

  // GenTransOptN selects the card count for the win-rate table optimization.
  // Set it before optimizing, mirroring Java's static GenTransOptUtil.N.
  extern int GenTransOptN;

  // GenTransOptData reads texas_data_N.txt and writes texas_data_opt_N.txt:
  // rows whose win-rate differs from the most frequent one of their
  // suit-stripped group are written with flag 0, group winners with flag 1.
  void GenTransOptData();

} // namespace texas_algorithm
