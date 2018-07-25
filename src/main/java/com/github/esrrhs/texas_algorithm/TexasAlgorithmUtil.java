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
		public KeyData(int index, int postion, long max, int type)
		{
			this.index = index;
			this.postion = postion;
			this.max = max;
			this.type = type;
		}

		private int index;
		private int postion;
		private long max;
		private int type;

		public int getIndex()
		{
			return index;
		}

		public int getPostion()
		{
			return postion;
		}

		public long getMax()
		{
			return max;
		}

		public int getType()
		{
			return type;
		}
	}

	public static class ProbilityData
	{
		private float avg;
		private float min;
		private float max;

		public ProbilityData(float avg, float min, float max)
		{
			this.avg = avg;
			this.min = min;
			this.max = max;
		}

		public float getAvg()
		{
			return avg;
		}

		public float getMin()
		{
			return min;
		}

		public float getMax()
		{
			return max;
		}
	}

	public static ConcurrentHashMap<Long, KeyData> colorMap = new ConcurrentHashMap<>();
	public static ConcurrentHashMap<Long, KeyData> normalMap = new ConcurrentHashMap<>();
	public static ConcurrentHashMap<Long, ProbilityData>[] probilityMap = new ConcurrentHashMap[7];
	public static ConcurrentHashMap<Long, ProbilityData>[] optprobilityMap = new ConcurrentHashMap[7];

	public static void main(String[] args)
	{
		gen();
		genOpt();
		genTrans();
		genTransOpt();
		genHand();
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
		File file1 = new File("texas_data_color.txt");
		File file2 = new File("texas_data_normal.txt");
		if (file.exists() && !file1.exists() && !file2.exists())
		{
			GenOptUtil.optNormalData();
			GenOptUtil.optColorData();
		}
	}

	private static void genHand()
	{
		for (int i = 4; i >= 0; i--)
		{
			GenHandUtil.N = i;
			GenHandUtil.genKey();
		}
	}

	private static void genTrans()
	{
		File file = new File("texas_data.txt");
		if (file.exists())
		{
			for (int i = 6; i >= 2; i--)
			{
				File file1 = new File("texas_data_" + i + ".txt");
				if (!file1.exists())
				{
					GenTransUtil.N = i;
					GenTransUtil.genKey();
					GenTransUtil.transData();
				}
			}
		}
	}

	private static void genTransOpt()
	{
		File file = new File("texas_data.txt");
		if (file.exists())
		{
			for (int i = 6; i >= 2; i--)
			{
				File file1 = new File("texas_data_" + i + ".txt");
				if (file1.exists())
				{
					File file2 = new File("texas_data_opt_" + i + ".txt");
					if (!file2.exists())
					{
						GenTransOptUtil.N = i;
						GenTransOptUtil.optData();
					}
				}
			}
		}
	}

	public static void load()
	{
		try
		{
			long begin = System.currentTimeMillis();
			FileInputStream inputStream = new FileInputStream("texas_data_color.txt");
			loadColor(inputStream);
			inputStream.close();
			FileInputStream inputStream1 = new FileInputStream("texas_data_normal.txt");
			loadNormal(inputStream1);
			inputStream1.close();
			System.out.println("load time " + (System.currentTimeMillis() - begin));
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static void loadProbility()
	{
		try
		{
			long begin = System.currentTimeMillis();
			for (int i = 6; i >= 2; i--)
			{
				FileInputStream inputStream2 = new FileInputStream("texas_data_opt_" + i + ".txt");
				loadProbility(i, inputStream2);
				inputStream2.close();
			}
			System.out.println("load time " + (System.currentTimeMillis() - begin));
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	private static void loadProbility(int i, FileInputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

		probilityMap[i] = new ConcurrentHashMap<>();
		optprobilityMap[i] = new ConcurrentHashMap<>();

		String str = null;
		while ((str = bufferedReader.readLine()) != null)
		{
			String[] params = str.split(" ");
			long key = Long.parseLong(params[0]);
			long type = Long.parseLong(params[1]);
			float probility = Float.parseFloat(params[2]);
			float min = Float.parseFloat(params[3]);
			float max = Float.parseFloat(params[4]);

			if (type == 0)
			{
				probilityMap[i].put(key, new ProbilityData(probility, min, max));
			}
			else
			{
				optprobilityMap[i].put(key, new ProbilityData(probility, min, max));
			}
		}
		bufferedReader.close();
	}

	public static void loadNormal(InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

		normalMap.clear();
		String str = null;
		while ((str = bufferedReader.readLine()) != null)
		{
			String[] params = str.split(" ");
			long key = Long.parseLong(params[0]);
			int i = Integer.parseInt(params[1]);
			int index = Integer.parseInt(params[2]);
			long max = Long.parseLong(params[5]);
			int type = Integer.parseInt(params[7]);

			KeyData keyData = new KeyData(index, i, max, type);
			normalMap.put(key, keyData);
		}
		bufferedReader.close();
	}

	public static void loadColor(InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));
		colorMap.clear();
		String str = null;
		while ((str = bufferedReader.readLine()) != null)
		{
			String[] params = str.split(" ");
			long key = Long.parseLong(params[0]);
			int i = Integer.parseInt(params[1]);
			int index = Integer.parseInt(params[2]);
			long max = Long.parseLong(params[5]);
			int type = Integer.parseInt(params[7]);

			KeyData keyData = new KeyData(index, i, max, type);
			colorMap.put(key, keyData);
		}
		bufferedReader.close();
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
		if (str.startsWith("方"))
		{
			return (new Poke(Poke.PokeColor_FANG, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("梅"))
		{
			return (new Poke(Poke.PokeColor_MEI, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("红"))
		{
			return (new Poke(Poke.PokeColor_HONG, strToPokeValue(str.substring(1)))).toByte();
		}
		else if (str.startsWith("黑"))
		{
			return (new Poke(Poke.PokeColor_HEI, strToPokeValue(str.substring(1)))).toByte();
		}
		else
		{
			return 0;
		}
	}

	public static String keyToStr(long key)
	{
		return GenUtil.toString(key);
	}

	public static List<Byte> keyToByte(long k)
	{
		ArrayList<Byte> cs = new ArrayList<>();
		if (k > 1000000000000L)
		{
			cs.add(((byte) (k % 100000000000000L / 1000000000000L)));
		}
		if (k > 10000000000L)
		{
			cs.add(((byte) (k % 1000000000000L / 10000000000L)));
		}
		if (k > 100000000L)
		{
			cs.add(((byte) (k % 10000000000L / 100000000L)));
		}
		if (k > 1000000L)
		{
			cs.add(((byte) (k % 100000000L / 1000000L)));
		}
		if (k > 10000L)
		{
			cs.add(((byte) (k % 1000000L / 10000L)));
		}
		if (k > 100L)
		{
			cs.add(((byte) (k % 10000L / 100L)));
		}
		if (k > 1L)
		{
			cs.add(((byte) (k % 100L / 1L)));
		}
		return cs;
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

	public static KeyData getKeyData(String str)
	{
		List<Byte> pokes = strToPokes(str);
		if (pokes.size() != 7)
		{
			return null;
		}

		return getKeyData(pokes);
	}

	public static KeyData getKeyData(long key)
	{
		long colorKey = GenOptUtil.changeColor(key);
		KeyData color = colorMap.get(colorKey);
		long normalKey = GenOptUtil.removeColor(key);
		KeyData normal = normalMap.get(normalKey);
		if (color == null)
		{
			return normal;
		}
		if (normal == null)
		{
			return normal;
		}
		if (color.getIndex() > normal.getIndex())
		{
			return color;
		}
		else
		{
			return color;
		}
	}

	public static KeyData getKeyData(List<Byte> pokes)
	{
		long key = GenUtil.genCardBind(pokes);

		return getKeyData(key);
	}

	public static int getWinPosition(String str)
	{
		KeyData keyData = getKeyData(str);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getPostion();
	}

	public static int getWinPosition(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getPostion();
	}

	public static double getWinProbability(String str)
	{
		KeyData keyData = getKeyData(str);
		if (keyData == null)
		{
			return 0;
		}
		return (double) keyData.getIndex() / GenUtil.total;
	}

	public static double getWinProbability(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		if (keyData == null)
		{
			return 0;
		}
		return (double) keyData.getIndex() / GenUtil.total;
	}

	public static long getWinMax(String str)
	{
		KeyData keyData = getKeyData(str);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getMax();
	}

	public static long getWinMax(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getMax();
	}

	public static int getWinType(String str)
	{
		KeyData keyData = getKeyData(str);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getType();
	}

	public static int getWinType(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		if (keyData == null)
		{
			return 0;
		}
		return keyData.getType();
	}

	public static int compare(String str1, String str2)
	{
		return compare(strToPokes(str1), strToPokes(str2));
	}

	public static int compare(List<Byte> bytes1, List<Byte> bytes2)
	{
		return compare(GenUtil.genCardBind(bytes1), GenUtil.genCardBind(bytes2));
	}

	public static int compare(long k1, long k2)
	{
		KeyData keyData1 = getKeyData(k1);
		KeyData keyData2 = getKeyData(k2);
		if (keyData1 == null && keyData2 == null)
		{
			return 0;
		}
		if (keyData1 == null)
		{
			return -1;
		}
		if (keyData2 == null)
		{
			return 1;
		}
		return keyData1.getPostion() - keyData2.getPostion();
	}

	public static float getHandProbability(String hand, String pub)
	{
		return getHandProbability(strToPokes(hand), strToPokes(pub));
	}

	public static ProbilityData getHandProbability(long k)
	{
		int num = 0;
		if (k > 10000000000L)
		{
			num = 6;
		}
		else if (k > 100000000L)
		{
			num = 5;
		}
		else if (k > 1000000L)
		{
			num = 4;
		}
		else if (k > 10000L)
		{
			num = 3;
		}
		else if (k > 100L)
		{
			num = 2;
		}
		if (num < 2 || num > 6)
		{
			return null;
		}

		ProbilityData probilityData = probilityMap[num].get(k);
		if (probilityData == null)
		{
			k = GenOptUtil.removeColor(k);
			probilityData = optprobilityMap[num].get(k);
		}
		return probilityData;
	}

	public static float getHandProbability(long hand, long pub)
	{
		return getHandProbability(keyToByte(hand), keyToByte(pub));
	}

	public static float getHandProbability(List<Byte> hand, List<Byte> pub)
	{
		hand.addAll(pub);
		Collections.sort(hand);
		long totalkey = GenUtil.genCardBind(hand);
		Collections.sort(pub);
		long pubkey = GenUtil.genCardBind(pub);

		ProbilityData totalProbilityData = getHandProbability(totalkey);
		ProbilityData pubProbilityData = getHandProbability(pubkey);

		if (totalProbilityData == null || pubProbilityData == null)
		{
			return 0;
		}

		float p = 0.5f;

		if (totalProbilityData.avg > pubProbilityData.avg)
		{
			p += 0.5f * (totalProbilityData.avg - pubProbilityData.avg) / (pubProbilityData.max - pubProbilityData.avg);
		}
		else
		{
			p += 0.5f * (totalProbilityData.avg - pubProbilityData.avg) / (pubProbilityData.avg - pubProbilityData.min);
		}

		return p;
	}
}
