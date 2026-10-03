#include "poke.h"

namespace texas_algorithm
{

  const int PokeValues[13] = {
      PokeValue2, PokeValue3, PokeValue4, PokeValue5, PokeValue6, PokeValue7,
      PokeValue8, PokeValue9, PokeValue10, PokeValueJ, PokeValueQ, PokeValueK, PokeValueA};

  const char *const HuaseName[4] = {"方", "梅", "红", "黑"};

  const char *const ValueName[15] = {"", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"};

  std::string PokeToString(const Poke &p)
  {
    if (IsGui(p))
    {
      return "鬼";
    }
    return std::string(HuaseName[p.color]) + ValueName[p.value];
  }

} // namespace texas_algorithm
