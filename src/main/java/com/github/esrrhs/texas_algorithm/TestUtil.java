package com.github.esrrhs.texas_algorithm;

public class TestUtil
{
	public static void main(String[] args)
	{
		TexasAlgorithmUtil.load();
		String cards = "方4,方A,黑2,黑A,黑3,黑5,黑6";
		String cards1 = "红8,方A,方2,黑8,黑3,黑5,黑7";
		System.out.println(TexasAlgorithmUtil.getWinPosition(cards));
		System.out.println(TexasAlgorithmUtil.getWinProbability(cards));
		System.out.println(TexasAlgorithmUtil.keyToStr(TexasAlgorithmUtil.getWinMax(cards)));
		System.out.println(TexasAlgorithmUtil.getWinType(cards));

		System.out.println(TexasAlgorithmUtil.getWinPosition(cards1));
		System.out.println(TexasAlgorithmUtil.getWinProbability(cards1));
		System.out.println(TexasAlgorithmUtil.keyToStr(TexasAlgorithmUtil.getWinMax(cards1)));
		System.out.println(TexasAlgorithmUtil.getWinType(cards1));

		System.out.println(TexasAlgorithmUtil.compare(cards, cards1));
	}

}
