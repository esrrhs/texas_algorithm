package com.github.esrrhs.texas_algorithm;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;

public class TexasCardUtil
{
	public static final int TEXAS_CARD_TYPE_GAOPAI = 1; //高牌
	public static final int TEXAS_CARD_TYPE_DUIZI = 2; //对子
	public static final int TEXAS_CARD_TYPE_LIANGDUI = 3; //两对
	public static final int TEXAS_CARD_TYPE_SANTIAO = 4; //三条
	public static final int TEXAS_CARD_TYPE_SHUNZI = 5; //顺子
	public static final int TEXAS_CARD_TYPE_TONGHUA = 6; //同花
	public static final int TEXAS_CARD_TYPE_HULU = 7; //葫芦
	public static final int TEXAS_CARD_TYPE_SITIAO = 8; //四条
	public static final int TEXAS_CARD_TYPE_TONGHUASHUN = 9; //同花顺
	public static final int TEXAS_CARD_TYPE_KINGTONGHUASHUN = 10; //皇家同花顺

	private static class TexasPokeLogicValueComparator implements Comparator
	{
		public int compare(Object o1, Object o2)
		{
			Poke p1 = (Poke) o1;
			Poke p2 = (Poke) o2;
			if (p1.getValue() < p2.getValue())
				return 1;
			else if (p1.getValue() > p2.getValue())
				return -1;
			else
				return 0;
		}
	}

	//
	public static int getCardTypeUnordered(ArrayList<Poke> cards)
	{
		Collections.sort(cards, new TexasPokeLogicValueComparator());
		return getCardType(cards);
	}

	//cards里的牌必须是排好序的
	private static int getCardType(ArrayList<Poke> cards)
	{
		if (cards.size() != 5)
		{
			return 0;
		}
		boolean sameColor = true, lineCard = true;
		int firstColor = cards.get(0).getColor();
		int firstValue = cards.get(0).getValue();

		//牌型分析
		for (int i = 1; i < 5; ++i)
		{
			if (cards.get(i).getColor() != firstColor)
			{
				sameColor = false;
			}
			if (cards.get(i).getValue() + i != firstValue)
			{
				lineCard = false;
			}
			if (!sameColor && !lineCard)
				break;
		}

		//最小顺子
		if (lineCard == false && firstValue == 14)
		{
			int value = 1;
			int i = 1;
			for (; i < 5; ++i)
			{
				value = cards.get(i).getValue();
				if ((value + i + 8) != firstValue)
					break;
			}
			if (i == 5)
			{
				lineCard = true;
			}
		}

		//皇家同花顺
		if (sameColor && lineCard && cards.get(1).getValue() == 13)
		{
			return TEXAS_CARD_TYPE_KINGTONGHUASHUN;
		}

		//同花顺
		if (sameColor && lineCard)
		{
			return TEXAS_CARD_TYPE_TONGHUASHUN;
		}

		//顺子
		if (!sameColor && lineCard)
		{
			return TEXAS_CARD_TYPE_SHUNZI;
		}

		//同花
		if (sameColor && !lineCard)
		{
			return TEXAS_CARD_TYPE_TONGHUA;
		}

		//分析相同牌有几个
		AnalyseResult analyseResult = AnalyseCards(cards);
		if (analyseResult.fourCount == 1)
			return TEXAS_CARD_TYPE_SITIAO; //四条
		if (analyseResult.threeCount == 1 && analyseResult.twoCount == 1)
			return TEXAS_CARD_TYPE_HULU; //葫芦
		if (analyseResult.threeCount == 1 && analyseResult.twoCount == 0)
			return TEXAS_CARD_TYPE_SANTIAO; //三条
		if (analyseResult.twoCount == 2)
		{
			return TEXAS_CARD_TYPE_LIANGDUI; //两对
		}
		if (analyseResult.twoCount == 1 && analyseResult.oneCount == 3)
			return TEXAS_CARD_TYPE_DUIZI; //对子

		return TEXAS_CARD_TYPE_GAOPAI; //高牌
	}

	private static class AnalyseResult
	{
		int fourCount = 0;
		int threeCount = 0;
		int twoCount = 0;
		int oneCount = 0;
		ArrayList<Integer> fourLogicValue = new ArrayList<>();
		ArrayList<Integer> threeLogicValue = new ArrayList<>();
		ArrayList<Integer> twoLogicValue = new ArrayList<>();
		ArrayList<Integer> oneLoggicValue = new ArrayList<>();
	}

	private static AnalyseResult AnalyseCards(ArrayList<Poke> cards)
	{
		AnalyseResult analyseResult = new AnalyseResult();
		for (int i = 0; i < cards.size(); ++i)
		{
			int sameCount = 1;
			int[] sameCardData = new int[4];
			int cardValue = cards.get(i).getValue();
			//			log.info("i:{} value:{}", i, cardValue);

			for (int j = i + 1; j < cards.size(); ++j)
			{
				if (cards.get(j).getValue() != cardValue)
				{
					break;
				}
				else
				{
					//					log.info("find same card. j:{} value:{}", j, cardValue);
					sameCount++;
				}
			}

			switch (sameCount)
			{
				case 1:
					analyseResult.oneCount += 1;
					analyseResult.oneLoggicValue.add(cardValue);
					break;
				case 2:
					analyseResult.twoCount += 1;
					analyseResult.twoLogicValue.add(cardValue);
					break;
				case 3:
					analyseResult.threeCount += 1;
					analyseResult.threeLogicValue.add(cardValue);
					break;
				case 4:
					analyseResult.fourCount += 1;
					analyseResult.fourLogicValue.add(cardValue);
					break;
				default:
			}
			i += sameCount - 1;
		}

		return analyseResult;
	}

	public static int fiveFromSeven(ArrayList<Poke> cards, ArrayList<Poke> pickedCards)
	{
		if (pickedCards == null)
		{
			pickedCards = new ArrayList<>();
		}
		pickedCards.clear();
		ArrayList<Poke> tmpSevenCards = cards;
		//对七张牌排序
		Collections.sort(tmpSevenCards, new TexasPokeLogicValueComparator());
		//		log.info("after sort. tmpSevenCards:{}", tmpSevenCards);

		for (int i = 0; i < 5; ++i)
		{
			pickedCards.add(tmpSevenCards.get(i));
		}
		int cardType = getCardType(pickedCards);
		int tmpCardType = 0;
		ArrayList<Poke> tmpPickedCards = new ArrayList<>();

		for (int i = 0; i < 3; ++i)
		{
			for (int j = i + 1; j < 4; ++j)
			{
				for (int k = j + 1; k < 5; ++k)
				{
					for (int l = k + 1; l < 6; ++l)
					{
						for (int m = l + 1; m < 7; ++m)
						{
							tmpPickedCards.clear();
							tmpPickedCards.add(tmpSevenCards.get(i));
							tmpPickedCards.add(tmpSevenCards.get(j));
							tmpPickedCards.add(tmpSevenCards.get(k));
							tmpPickedCards.add(tmpSevenCards.get(l));
							tmpPickedCards.add(tmpSevenCards.get(m));

							tmpCardType = getCardType(tmpPickedCards);

							if (compareCards(tmpPickedCards, pickedCards) == 1)
							{//找到更大的牌 进行替换
								cardType = tmpCardType;
								pickedCards.clear();
								for (Poke poke : tmpPickedCards)
								{
									pickedCards.add(poke);
								}
							}

						}
					}
				}
			}
		}
		return cardType;
	}

	//需要是拍好序的牌
	public static int compareCards(ArrayList<Poke> firstCards, ArrayList<Poke> secondCards)
	{
		int firstType = getCardType(firstCards);
		int secondType = getCardType(secondCards);

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
			case TEXAS_CARD_TYPE_GAOPAI: //单牌
			{
				int i = 0;
				for (; i < 5; ++i)
				{
					int secondValue = secondCards.get(i).getValue();
					int firstValue = firstCards.get(i).getValue();
					if (firstValue > secondValue)
						return 1;
					else if (firstValue < secondValue)
						return -1;
					else
						continue;
				}
				if (i == 5)
				{
					return 0;
				}
				break;
			}
			case TEXAS_CARD_TYPE_DUIZI:
			case TEXAS_CARD_TYPE_LIANGDUI:
			case TEXAS_CARD_TYPE_SANTIAO:
			case TEXAS_CARD_TYPE_SITIAO:
			case TEXAS_CARD_TYPE_HULU:
			{
				AnalyseResult secondAnalyseRes = AnalyseCards(secondCards);
				AnalyseResult firstAnalyseRes = AnalyseCards(firstCards);
				//四条比较
				if (firstAnalyseRes.fourCount > 0)
				{
					int secondValue = secondAnalyseRes.fourLogicValue.get(0);
					int firesValue = firstAnalyseRes.fourLogicValue.get(0);
					//比较四条的值
					if (firesValue != secondValue)
						return (firesValue > secondValue) ? 1 : -1;
					//比较单牌的值
					secondValue = secondAnalyseRes.oneLoggicValue.get(0);
					firesValue = firstAnalyseRes.oneLoggicValue.get(0);
					if (firesValue != secondValue)
						return (firesValue > secondValue) ? 1 : -1;
					else
						return 0;
				}

				//三条比较
				if (firstAnalyseRes.threeCount > 0)
				{
					int secondValue = secondAnalyseRes.threeLogicValue.get(0);
					int firstValue = firstAnalyseRes.threeLogicValue.get(0);
					//比较三条
					if (firstValue != secondValue)
						return (firstValue > secondValue) ? 1 : -1;

					//葫芦牌型
					if (firstType == TEXAS_CARD_TYPE_HULU)
					{
						secondValue = secondAnalyseRes.twoLogicValue.get(0);
						firstValue = firstAnalyseRes.twoLogicValue.get(0);
						//
						if (firstValue != secondValue)
							return (firstValue > secondValue) ? 1 : -1;
						else
							return 0;
					}
					else //三条带单
					{
						int i = 0;
						for (; i < firstAnalyseRes.oneCount; ++i)
						{
							secondValue = secondAnalyseRes.oneLoggicValue.get(i);
							firstValue = firstAnalyseRes.oneLoggicValue.get(i);
							//
							if (firstValue > secondValue)
								return 1;
							else if (firstValue < secondValue)
								return -1;
							else
								continue;
						}
						if (i == firstAnalyseRes.oneCount)
							return 0;
					}
				}

				//对子
				int i = 0;
				for (; i < firstAnalyseRes.twoCount; ++i)
				{
					int secondValue = secondAnalyseRes.twoLogicValue.get(i);
					int firstValue = firstAnalyseRes.twoLogicValue.get(i);
					if (firstValue > secondValue)
						return 1;
					else if (firstValue < secondValue)
						return -1;
					else
						continue;
				}
				//比较单牌
				for (i = 0; i < firstAnalyseRes.oneCount; ++i)
				{
					int secondValue = secondAnalyseRes.oneLoggicValue.get(i);
					int firstValue = firstAnalyseRes.oneLoggicValue.get(i);
					if (firstValue > secondValue)
					{
						return 1;
					}
					else if (firstValue < secondValue)
					{
						return -1;
					}
					else
					{
						continue;
					}
				}
				if (i == firstAnalyseRes.oneCount)
				{
					return 0;
				}
				break;
			}
			case TEXAS_CARD_TYPE_SHUNZI:
			case TEXAS_CARD_TYPE_TONGHUASHUN:
			{
				int secondValue = secondCards.get(0).getValue();
				int firstValue = firstCards.get(0).getValue();

				boolean firstMin = (firstValue == (firstCards.get(0).getValue() + 9)); //是最小顺子吗 第一张5 第五张14
				boolean secondMin = (secondValue == (secondCards.get(0).getValue() + 9));

				if (firstMin == true && secondMin == false)
				{//第一个是最小顺子 第一个小于第二个
					return -1;
				}
				if (firstMin == false && secondMin == true)
				{
					return 1;
				}
				//
				else
				{
					if (firstValue == secondValue)
					{
						return 0;
					}
					return (firstValue > secondValue) ? 1 : -1;
				}
			}
			case TEXAS_CARD_TYPE_TONGHUA:
			{
				//
				int i = 0;
				for (; i < 5; ++i)
				{
					int secondValue = secondCards.get(i).getValue();
					int firstValue = firstCards.get(i).getValue();

					if (firstValue == secondValue)
						continue;
					return (firstValue > secondValue) ? 1 : -1;
				}
				if (i == 5)
				{
					return 0;
				}
				break;
			}
			default:
				return 0;
		}
		return 0;
	}
}
