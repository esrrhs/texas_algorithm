package com.github.esrrhs.texas_algorithm;

public class TestUtil
{
	public static void main(String[] args)
	{
		TexasAlgorithmUtil.load();
		System.out.println(TexasAlgorithmUtil.getWinPosition("♦3,♣3,♥2,♥3,♥4,♠3,♠4"));
		System.out.println(TexasAlgorithmUtil.getWinProbability("♦3,♣3,♥2,♥3,♥4,♠3,♠4"));
	}

}
