package com.github.esrrhs.texas_algorithm;

import java.io.File;

public class TexasAlgorithmUtil
{
	public static void main(String[] args)
	{
		gen();
		genOpt();
		genTrans();
	}

	private static void gen()
	{
		File file = new File("texas_data.txt");
		if (!file.exists())
		{
			GenUtil.genKey();
			GenUtil.outputData();
		}
	}

	private static void genOpt()
	{
		File file = new File("texas_data.txt");
		if (file.exists())
		{
			GenOptUtil.optNormalData();
			GenOptUtil.optColorData();
		}
	}

	private static void genTrans()
	{
		File file = new File("texas_data.txt");
		if (file.exists())
		{
			for (int i = 6; i >= 2; i--)
			{
				GenTransUtil.N = i;
				GenTransUtil.genKey();
				GenTransUtil.transData();
			}
		}
	}

}
