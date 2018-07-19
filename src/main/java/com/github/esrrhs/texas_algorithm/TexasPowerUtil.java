package com.github.esrrhs.texas_algorithm;

public class TexasPowerUtil
{
	public static void main(String[] args)
	{
		genOpt();
	}

	private static void gen()
	{
		GenUtil.genKey();
		GenUtil.outputData();
	}

	private static void genOpt()
	{
		for (int i = 2; i <= 7; i++)
		{
			GenOptUtil.N = i;
			GenOptUtil.genKey();
			GenOptUtil.transData();
		}
	}
}
