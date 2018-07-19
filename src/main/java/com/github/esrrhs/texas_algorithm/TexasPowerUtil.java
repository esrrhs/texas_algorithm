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
		GenOptUtil.genKey();
		//GenOptUtil.transData();
	}
}
