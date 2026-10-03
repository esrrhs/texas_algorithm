#pragma once

#include <cstdint>

namespace texas_algorithm
{

  // OptColorData writes texas_data_color.txt: for every flush-type row of
  // texas_data.txt, the suit-normalized key, deduplicated.
  void OptColorData();

  // OptNormalData writes texas_data_normal.txt: for every non-flush-type row of
  // texas_data.txt, the suit-stripped key, deduplicated.
  void OptNormalData();

  // RemoveColor collapses all suit colors of the non-wild cards to 方 (diamonds),
  // sorts and re-binds the key, mirroring Java's GenOptUtil.removeColor.
  int64_t RemoveColor(int64_t k);

  // ChangeColor normalizes the dominant suit of the key to 黑 (spades) and every
  // other suit to 方 (diamonds), keeping wild cards, then sorts and re-binds,
  // mirroring Java's GenOptUtil.changeColor.
  int64_t ChangeColor(int64_t k);

} // namespace texas_algorithm
