package com.github.esrrhs.texas_algorithm;

public class TestUtil
{
	public static void main(String[] args)
	{
		TexasAlgorithmUtil.load();
		String cards = "♦6,♣3,♠2,♠3,♠4,♠5,♠6";
		System.out.println(TexasAlgorithmUtil.getWinPosition(cards));
		System.out.println(TexasAlgorithmUtil.getWinProbability(cards));
		System.out.println(TexasAlgorithmUtil.keyToStr(TexasAlgorithmUtil.getWinMax(cards)));
		System.out.println(TexasAlgorithmUtil.getWinType(cards));
	}

}
