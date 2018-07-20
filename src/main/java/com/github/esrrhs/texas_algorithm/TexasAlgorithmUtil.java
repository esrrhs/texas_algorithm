package com.github.esrrhs.texas_algorithm;

import java.io.*;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

public class TexasAlgorithmUtil
{
	public static class KeyData
	{
		private int index;
		private int postion;

		public int getIndex()
		{
			return index;
		}

		public int getPostion()
		{
			return postion;
		}
	}

	public static ConcurrentHashMap<Long, KeyData> colorMap = new ConcurrentHashMap<>();
	public static ConcurrentHashMap<Long, KeyData> normalMap = new ConcurrentHashMap<>();

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

	public static byte strToPokeValue(String str)
	{
		if (str.equals("A"))
		{
			return Poke.PokeValue_A;
		}
		else if (str.equals("K"))
		{
			return Poke.PokeValue_K;
		}
		else if (str.equals("Q"))
		{
			return Poke.PokeValue_Q;
		}
		else if (str.equals("J"))
		{
			return Poke.PokeValue_J;
		}
		else
		{
			return Byte.parseByte(str);
		}
	}

	public static byte strToPoke(String str)
	{
		if (str.startsWith("♦"))
		{
			return (new Poke(Poke.PokeColor_FANG, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("♣"))
		{
			return (new Poke(Poke.PokeColor_MEI, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("♥"))
		{
			return (new Poke(Poke.PokeColor_HONG, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("♠"))
		{
			return (new Poke(Poke.PokeColor_HEI, strToPokeValue(str.substring(1)))).toByte();
		}
		else
		{
			return 0;
		}
	}

	public static List<Byte> strToPokes(String str)
	{
		List<Byte> ret = new ArrayList<>();
		String[] strs = str.split(",");
		for (String s : strs)
		{
			ret.add(strToPoke(s));
		}
		return ret;
	}

	public static int getWinPosition(String str)
	{
		List<Byte> pokes = strToPokes(str);
		if (pokes.size() != 7)
		{
			return 0;
		}

		Collections.sort(pokes);

		long key = GenUtil.genCardBind(pokes);
		KeyData color = colorMap.get(key);
		long normalKey = GenOptUtil.removeColor(key);
		KeyData normal = normalMap.get(normalKey);
		if (color == null)
		{
			return normal.getPostion();
		}
		if (normal == null)
		{
			return normal.getPostion();
		}
		if (color.getIndex() > normal.getIndex())
		{
			return color.getPostion();
		}
		else
		{
			return normal.getPostion();
		}
	}

	public static double getWinProbability(String str)
	{
		List<Byte> pokes = strToPokes(str);
		if (pokes.size() != 7)
		{
			return 0;
		}

		Collections.sort(pokes);

		long key = GenUtil.genCardBind(pokes);
		KeyData color = colorMap.get(key);
		long normalKey = GenOptUtil.removeColor(key);
		KeyData normal = normalMap.get(normalKey);
		if (color == null)
		{
			return (double)normal.getIndex() / GenUtil.total;
		}
		if (normal == null)
		{
			return (double)normal.getIndex() / GenUtil.total;
		}
		if (color.getIndex() > normal.getIndex())
		{
			return (double)color.getIndex() / GenUtil.total;
		}
		else
		{
			return (double)normal.getIndex() / GenUtil.total;
		}
	}
}
