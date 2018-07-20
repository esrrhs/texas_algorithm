package com.github.esrrhs.texas_algorithm;

import java.io.*;

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

	public static void load()
	{
		try
		{
			FileInputStream inputStream = new FileInputStream("texas_data_color.txt");
			loadColor(inputStream);
			FileInputStream inputStream1 = new FileInputStream("texas_data_color.txt");
			loadNormal(inputStream1);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static void loadNormal(InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

		String str = null;
		while ((str = bufferedReader.readLine()) != null)
		{
			String[] params = str.split(" ");
			long key = Long.parseLong(params[0]);
			long i = Long.parseLong(params[1]);
			long index = Long.parseLong(params[2]);
			long total = Long.parseLong(params[3]);
			long max = Long.parseLong(params[5]);

		}
	}

	public static void loadColor(InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

	}
}
