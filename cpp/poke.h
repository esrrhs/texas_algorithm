#pragma once

#include <string>

namespace texas_algorithm
{

  // Poke value constants (card ranks), aligned with Java's Poke.
  constexpr int PokeValue2 = 2;
  constexpr int PokeValue3 = 3;
  constexpr int PokeValue4 = 4;
  constexpr int PokeValue5 = 5;
  constexpr int PokeValue6 = 6;
  constexpr int PokeValue7 = 7;
  constexpr int PokeValue8 = 8;
  constexpr int PokeValue9 = 9;
  constexpr int PokeValue10 = 10;
  constexpr int PokeValueJ = 11;
  constexpr int PokeValueQ = 12;
  constexpr int PokeValueK = 13;
  constexpr int PokeValueA = 14;

  // Poke color constants, aligned with Java's Poke.
  constexpr int PokeColorFang = 0; // 方 diamonds
  constexpr int PokeColorMei = 1;  // 梅 clubs
  constexpr int PokeColorHong = 2; // 红 hearts
  constexpr int PokeColorHei = 3;  // 黑 spades

  // Poke is a single card: a color (suit) plus a value (rank).
  struct Poke
  {
    int color;
    int value;
  };

  constexpr Poke MakePoke(int color, int value) { return Poke{color, value}; }

  // PokeFromByte decodes a card byte (high nibble color, low nibble value),
  // the same encoding as Java's new Poke(byteValue).
  constexpr Poke PokeFromByte(int byteValue) { return Poke{byteValue >> 4, byteValue % 16}; }

  // IsLegalCardByte reports whether the byte belongs to the standard 52-card
  // deck: color (high nibble) 0..3 and value (low nibble) 2..14. The removed
  // joker byte 0x58 (88, color 5) is not legal.
  inline bool IsLegalCardByte(int i)
  {
    int color = i >> 4;
    int value = i % 16;
    return color >= PokeColorFang && color <= PokeColorHei && value >= PokeValue2 && value <= PokeValueA;
  }

  // PokeToByte encodes the card into a byte (color<<4 | value).
  inline int PokeToByte(const Poke &p) { return (p.color << 4) | p.value; }

  // PokeToString renders the card the same way as Java's Poke.toString:
  // a Chinese suit name followed by the rank, e.g. "方A".
  std::string PokeToString(const Poke &p);

  extern const int PokeValues[13];
  extern const char *const HuaseName[4];
  extern const char *const ValueName[15];

} // namespace texas_algorithm
