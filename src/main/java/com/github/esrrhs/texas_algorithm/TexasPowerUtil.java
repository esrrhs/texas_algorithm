package com.github.esrrhs.texas_algorithm;

import java.io.File;

public class TexasPowerUtil
{
	public static void main(String[] args)
	{
		gen();
		genOpt();
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
			for (int i = 2; i <= 7; i++)
			{
				GenOptUtil.N = i;
				GenOptUtil.genKey();
				GenOptUtil.transData();
			}
		}
	}
}
