#pragma once

#include <cstdint>

namespace texas_algorithm
{

  // GenExtraOptN selects the card count for the extra table optimization.
  // Set it before optimizing, mirroring Java's static GenExtraOptUtil.N.
  extern int GenExtraOptN;

  // GenExtraOptColorData writes texas_data_extra_color_N.txt from
  // texas_data_extra_N.txt, deduplicating the suit-normalized flush keys.
  void GenExtraOptColorData();

  // GenExtraOptNormalData writes texas_data_extra_normal_N.txt from
  // texas_data_extra_N.txt, deduplicating the suit-stripped non-flush keys.
  void GenExtraOptNormalData();

} // namespace texas_algorithm
