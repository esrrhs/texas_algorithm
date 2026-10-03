#pragma once

#include <vector>

#include "poke.h"

namespace texas_algorithm
{

  // Card type constants, aligned with Java's TexasCardUtil.
  constexpr int TexasCardTypeGaoPai = 1;          // 高牌 high card
  constexpr int TexasCardTypeDuiZi = 2;           // 对子 one pair
  constexpr int TexasCardTypeLiangDui = 3;        // 两对 two pairs
  constexpr int TexasCardTypeSanTiao = 4;         // 三条 three of a kind
  constexpr int TexasCardTypeShunZi = 5;          // 顺子 straight
  constexpr int TexasCardTypeTongHua = 6;         // 同花 flush
  constexpr int TexasCardTypeHuLu = 7;            // 葫芦 full house
  constexpr int TexasCardTypeSiTiao = 8;          // 四条 four of a kind
  constexpr int TexasCardTypeTongHuaShun = 9;     // 同花顺 straight flush
  constexpr int TexasCardTypeKingTongHuaShun = 10; // 皇家同花顺 royal flush

  // SortPokesDesc sorts cards by logic value descending, the same order as
  // Java's TexasPokeLogicValueComparator.
  void SortPokesDesc(std::vector<Poke> &cards);

  // GetCardTypeUnorderedWithoutGui sorts the cards by value descending in
  // place and returns their type, mirroring Java's getCardTypeUnorderedWithoutGui.
  int GetCardTypeUnorderedWithoutGui(std::vector<Poke> &cards);

  // FiveFromSeven picks the best 5 cards out of 7, handling wild cards.
  std::vector<Poke> FiveFromSeven(std::vector<Poke> cards);

  // FiveFromSevenWithoutGui picks the best 5 cards out of 7 without wild cards.
  std::vector<Poke> FiveFromSevenWithoutGui(std::vector<Poke> cards);

  // FiveFromSix picks the best 5 cards out of 6, handling wild cards.
  std::vector<Poke> FiveFromSix(std::vector<Poke> cards);

  // FiveFromSixWithoutGui picks the best 5 cards out of 6 without wild cards.
  std::vector<Poke> FiveFromSixWithoutGui(std::vector<Poke> cards);

  // FiveFromFive evaluates 5 cards, using wild cards if present.
  std::vector<Poke> FiveFromFive(std::vector<Poke> cards);

  // FiveFromFiveWithoutGui sorts the 5 cards by value descending.
  std::vector<Poke> FiveFromFiveWithoutGui(std::vector<Poke> cards);

  // CompareCardsWithoutGui compares two 5-card hands of the same card count.
  // Both hands must be sorted descending by value. Returns 1 if firstCards is
  // stronger, -1 if weaker, 0 if equal.
  int CompareCardsWithoutGui(const std::vector<Poke> &firstCards, const std::vector<Poke> &secondCards);

} // namespace texas_algorithm
