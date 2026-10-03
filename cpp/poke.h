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

  // GUI is the wild card (ghost card), aligned with Java's Poke.GUI.
  inline constexpr Poke GUI{5, 8};

  constexpr Poke MakePoke(int color, int value) { return Poke{color, value}; }

  // PokeFromByte decodes a card byte (high nibble color, low nibble value),
  // the same encoding as Java's new Poke(byteValue).
  constexpr Poke PokeFromByte(int byteValue) { return Poke{byteValue >> 4, byteValue % 16}; }

  // IsGuiByte reports whether the card byte is a wild card,
  // mirroring Java's static Poke.isGui(int).
  inline bool IsGuiByte(int i)
  {
    int color = i >> 4;
    int value = i % 16;
    return value == GUI.value && color == GUI.color;
  }

  inline bool IsGui(const Poke &p) { return p.value == GUI.value && p.color == GUI.color; }

  // PokeToByte encodes the card into a byte (color<<4 | value).
  inline int PokeToByte(const Poke &p) { return (p.color << 4) | p.value; }

  // PokeToString renders the card the same way as Java's Poke.toString:
  // a Chinese suit name followed by the rank, e.g. "方A", "鬼".
  std::string PokeToString(const Poke &p);

  extern const int PokeValues[13];
  extern const char *const HuaseName[4];
  extern const char *const ValueName[15];

} // namespace texas_algorithm
