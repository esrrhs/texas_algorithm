#pragma once

#include <cstdint>
#include <vector>

namespace texas_algorithm
{

  // Quicksort sorts the keys ascending by poker strength (GenCompare) using a
  // bounded parallel quicksort, mirroring Java's Sorter. Ties (equal strength,
  // different keys) may end up in any order, which does not change the rank
  // each key maps to in the generated table.
  void Quicksort(std::vector<int64_t> &input);

} // namespace texas_algorithm
