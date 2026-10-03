#include "texas_card_util.h"

#include <algorithm>

#include "gen_util.h"

namespace texas_algorithm
{

  namespace
  {

    struct AnalyseResult
    {
      int fourCount = 0;
      int threeCount = 0;
      int twoCount = 0;
      int oneCount = 0;
      std::vector<int> fourLogicValue;
      std::vector<int> threeLogicValue;
      std::vector<int> twoLogicValue;
      std::vector<int> oneLogicValue;
    };

    // analyseCards groups cards by value; cards must be sorted descending by value.
    AnalyseResult AnalyseCards(const std::vector<Poke> &cards)
    {
      AnalyseResult res;
      for (int i = 0; i < (int)cards.size(); i++)
      {
        int sameCount = 1;
        int cardValue = cards[i].value;

        for (int j = i + 1; j < (int)cards.size(); j++)
        {
          if (cards[j].value != cardValue)
          {
            break;
          }
          sameCount++;
        }

        switch (sameCount)
        {
        case 1:
          res.oneCount++;
          res.oneLogicValue.push_back(cardValue);
          break;
        case 2:
          res.twoCount++;
          res.twoLogicValue.push_back(cardValue);
          break;
        case 3:
          res.threeCount++;
          res.threeLogicValue.push_back(cardValue);
          break;
        case 4:
          res.fourCount++;
          res.fourLogicValue.push_back(cardValue);
          break;
        default:
          break;
        }
        i += sameCount - 1;
      }

      return res;
    }

    // getCardType returns the type of 5 sorted cards, mirroring Java's private
    // getCardType. cards must be sorted descending by value and have length 5.
    int GetCardType(const std::vector<Poke> &cards)
    {
      if (cards.size() != 5)
      {
        return 0;
      }
      bool sameColor = true;
      bool lineCard = true;
      int firstColor = cards[0].color;
      int firstValue = cards[0].value;

      // 牌型分析
      for (int i = 1; i < 5; i++)
      {
        if (cards[i].color != firstColor)
        {
          sameColor = false;
        }
        if (cards[i].value + i != firstValue)
        {
          lineCard = false;
        }
        if (!sameColor && !lineCard)
        {
          break;
        }
      }

      // 最小顺子
      if (!lineCard && firstValue == 14)
      {
        int i = 1;
        for (; i < 5; i++)
        {
          int value = cards[i].value;
          if (value + i + 8 != firstValue)
          {
            break;
          }
        }
        if (i == 5)
        {
          lineCard = true;
        }
      }

      // 皇家同花顺
      if (sameColor && lineCard && cards[1].value == 13)
      {
        return TexasCardTypeKingTongHuaShun;
      }

      // 同花顺
      if (sameColor && lineCard)
      {
        return TexasCardTypeTongHuaShun;
      }

      // 顺子
      if (!sameColor && lineCard)
      {
        return TexasCardTypeShunZi;
      }

      // 同花
      if (sameColor && !lineCard)
      {
        return TexasCardTypeTongHua;
      }

      // 分析相同牌有几个
      AnalyseResult analyseRes = AnalyseCards(cards);
      if (analyseRes.fourCount == 1)
      {
        return TexasCardTypeSiTiao; // 四条
      }
      if (analyseRes.threeCount == 1 && analyseRes.twoCount == 1)
      {
        return TexasCardTypeHuLu; // 葫芦
      }
      if (analyseRes.threeCount == 1 && analyseRes.twoCount == 0)
      {
        return TexasCardTypeSanTiao; // 三条
      }
      if (analyseRes.twoCount == 2)
      {
        return TexasCardTypeLiangDui; // 两对
      }
      if (analyseRes.twoCount == 1 && analyseRes.oneCount == 3)
      {
        return TexasCardTypeDuiZi; // 对子
      }

      return TexasCardTypeGaoPai; // 高牌
    }

    // pickBestFromGui enumerates wild-card replacements over AllCards and keeps
    // the strongest hand, mirroring the gui path of Java's fiveFromFive/Six/Seven.
    std::vector<Poke> FiveFromWithGui(std::vector<Poke> cards,
                                      std::vector<Poke> (*withoutGui)(std::vector<Poke>))
    {
      int gui = 0;
      for (const Poke &poke : cards)
      {
        if (IsGui(poke))
        {
          gui++;
        }
      }
      if (gui == 0)
      {
        return withoutGui(cards);
      }

      std::vector<Poke> left;
      std::vector<int> leftBytes; // leftMap as set of card bytes
      for (const Poke &poke : cards)
      {
        if (!IsGui(poke))
        {
          left.push_back(poke);
          leftBytes.push_back(PokeToByte(poke));
        }
      }
      auto inLeft = [&leftBytes](int t)
      {
        for (int b : leftBytes)
        {
          if (b == t)
          {
            return true;
          }
        }
        return false;
      };

      std::vector<Poke> maxCards;
      std::vector<int> tmp(gui);
      Permutation(AllCards, 0, 0, gui, tmp, [&](std::vector<int> &tmp)
                  {
        for (int t : tmp)
        {
          if (inLeft(t) || t == PokeToByte(GUI))
          {
            return;
          }
        }

        std::vector<Poke> last;
        last.reserve(left.size() + tmp.size());
        last.insert(last.end(), left.begin(), left.end());
        for (int t : tmp)
        {
          last.push_back(PokeFromByte(t));
        }
        std::vector<Poke> picked = withoutGui(last);
        if (maxCards.empty() || CompareCardsWithoutGui(picked, maxCards) > 0)
        {
          maxCards = picked;
        } });
      return maxCards;
    }

  } // namespace

  void SortPokesDesc(std::vector<Poke> &cards)
  {
    std::stable_sort(cards.begin(), cards.end(),
                     [](const Poke &a, const Poke &b)
                     { return a.value > b.value; });
  }

  int GetCardTypeUnorderedWithoutGui(std::vector<Poke> &cards)
  {
    SortPokesDesc(cards);
    return GetCardType(cards);
  }

  std::vector<Poke> FiveFromSeven(std::vector<Poke> cards)
  {
    return FiveFromWithGui(std::move(cards), FiveFromSevenWithoutGui);
  }

  std::vector<Poke> FiveFromSevenWithoutGui(std::vector<Poke> cards)
  {
    // 对七张牌排序
    SortPokesDesc(cards);
    const std::vector<Poke> &tmpSevenCards = cards;

    std::vector<Poke> pickedCards(tmpSevenCards.begin(), tmpSevenCards.begin() + 5);

    for (int i = 0; i < 3; i++)
    {
      for (int j = i + 1; j < 4; j++)
      {
        for (int k = j + 1; k < 5; k++)
        {
          for (int l = k + 1; l < 6; l++)
          {
            for (int m = l + 1; m < 7; m++)
            {
              std::vector<Poke> tmpPickedCards{tmpSevenCards[i], tmpSevenCards[j], tmpSevenCards[k],
                                               tmpSevenCards[l], tmpSevenCards[m]};

              if (CompareCardsWithoutGui(tmpPickedCards, pickedCards) == 1)
              {
                // 找到更大的牌 进行替换
                pickedCards = tmpPickedCards;
              }
            }
          }
        }
      }
    }
    return pickedCards;
  }

  std::vector<Poke> FiveFromFive(std::vector<Poke> cards)
  {
    return FiveFromWithGui(std::move(cards), FiveFromFiveWithoutGui);
  }

  std::vector<Poke> FiveFromFiveWithoutGui(std::vector<Poke> cards)
  {
    SortPokesDesc(cards);
    return cards;
  }

  std::vector<Poke> FiveFromSix(std::vector<Poke> cards)
  {
    return FiveFromWithGui(std::move(cards), FiveFromSixWithoutGui);
  }

  std::vector<Poke> FiveFromSixWithoutGui(std::vector<Poke> cards)
  {
    // 对六张牌排序
    SortPokesDesc(cards);
    const std::vector<Poke> &tmpSixCards = cards;

    std::vector<Poke> pickedCards(tmpSixCards.begin(), tmpSixCards.begin() + 5);

    for (int i = 0; i < 3; i++)
    {
      for (int j = i + 1; j < 4; j++)
      {
        for (int k = j + 1; k < 5; k++)
        {
          for (int l = k + 1; l < 6; l++)
          {
            for (int m = l + 1; m < 6; m++)
            {
              std::vector<Poke> tmpPickedCards{tmpSixCards[i], tmpSixCards[j], tmpSixCards[k],
                                               tmpSixCards[l], tmpSixCards[m]};

              if (CompareCardsWithoutGui(tmpPickedCards, pickedCards) == 1)
              {
                // 找到更大的牌 进行替换
                pickedCards = tmpPickedCards;
              }
            }
          }
        }
      }
    }
    return pickedCards;
  }

  int CompareCardsWithoutGui(const std::vector<Poke> &firstCards, const std::vector<Poke> &secondCards)
  {
    int firstType = GetCardType(firstCards);
    int secondType = GetCardType(secondCards);

    if (firstType > secondType)
    {
      return 1;
    }
    if (firstType < secondType)
    {
      return -1;
    }
    switch (firstType)
    {
    case TexasCardTypeGaoPai: // 单牌
      for (int i = 0; i < 5; i++)
      {
        int secondValue = secondCards[i].value;
        int firstValue = firstCards[i].value;
        if (firstValue > secondValue)
        {
          return 1;
        }
        else if (firstValue < secondValue)
        {
          return -1;
        }
      }
      return 0;
    case TexasCardTypeDuiZi:
    case TexasCardTypeLiangDui:
    case TexasCardTypeSanTiao:
    case TexasCardTypeSiTiao:
    case TexasCardTypeHuLu:
    {
      AnalyseResult secondAnalyseRes = AnalyseCards(secondCards);
      AnalyseResult firstAnalyseRes = AnalyseCards(firstCards);
      // 四条比较
      if (firstAnalyseRes.fourCount > 0)
      {
        int secondValue = secondAnalyseRes.fourLogicValue[0];
        int firesValue = firstAnalyseRes.fourLogicValue[0];
        // 比较四条的值
        if (firesValue != secondValue)
        {
          return firesValue > secondValue ? 1 : -1;
        }
        // 比较单牌的值
        secondValue = secondAnalyseRes.oneLogicValue[0];
        firesValue = firstAnalyseRes.oneLogicValue[0];
        if (firesValue != secondValue)
        {
          return firesValue > secondValue ? 1 : -1;
        }
        return 0;
      }

      // 三条比较
      if (firstAnalyseRes.threeCount > 0)
      {
        int secondValue = secondAnalyseRes.threeLogicValue[0];
        int firstValue = firstAnalyseRes.threeLogicValue[0];
        // 比较三条
        if (firstValue != secondValue)
        {
          return firstValue > secondValue ? 1 : -1;
        }

        // 葫芦牌型
        if (firstType == TexasCardTypeHuLu)
        {
          secondValue = secondAnalyseRes.twoLogicValue[0];
          firstValue = firstAnalyseRes.twoLogicValue[0];
          if (firstValue != secondValue)
          {
            return firstValue > secondValue ? 1 : -1;
          }
          return 0;
        }
        // 三条带单
        for (int i = 0; i < firstAnalyseRes.oneCount; i++)
        {
          secondValue = secondAnalyseRes.oneLogicValue[i];
          firstValue = firstAnalyseRes.oneLogicValue[i];
          if (firstValue > secondValue)
          {
            return 1;
          }
          else if (firstValue < secondValue)
          {
            return -1;
          }
        }
        return 0;
      }

      // 对子
      for (int i = 0; i < firstAnalyseRes.twoCount; i++)
      {
        int secondValue = secondAnalyseRes.twoLogicValue[i];
        int firstValue = firstAnalyseRes.twoLogicValue[i];
        if (firstValue > secondValue)
        {
          return 1;
        }
        else if (firstValue < secondValue)
        {
          return -1;
        }
      }
      // 比较单牌
      for (int i = 0; i < firstAnalyseRes.oneCount; i++)
      {
        int secondValue = secondAnalyseRes.oneLogicValue[i];
        int firstValue = firstAnalyseRes.oneLogicValue[i];
        if (firstValue > secondValue)
        {
          return 1;
        }
        else if (firstValue < secondValue)
        {
          return -1;
        }
      }
      return 0;
    }
    case TexasCardTypeShunZi:
    case TexasCardTypeTongHuaShun:
    {
      int secondValue = secondCards[0].value;
      int firstValue = firstCards[0].value;

      // 是最小顺子吗 第一张5 第五张14
      bool firstMin = (firstCards[4].value == (firstCards[0].value - 12));
      bool secondMin = (secondCards[4].value == (secondCards[0].value - 12));

      if (firstMin && !secondMin)
      {
        // 第一个是最小顺子 第一个小于第二个
        return -1;
      }
      if (!firstMin && secondMin)
      {
        return 1;
      }
      if (firstValue == secondValue)
      {
        return 0;
      }
      return firstValue > secondValue ? 1 : -1;
    }
    case TexasCardTypeTongHua:
      for (int i = 0; i < 5; i++)
      {
        int secondValue = secondCards[i].value;
        int firstValue = firstCards[i].value;

        if (firstValue == secondValue)
        {
          continue;
        }
        return firstValue > secondValue ? 1 : -1;
      }
      return 0;
    default:
      return 0;
    }
  }

} // namespace texas_algorithm
