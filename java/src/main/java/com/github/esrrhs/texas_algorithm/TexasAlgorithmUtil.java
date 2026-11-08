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

	public static class WinRateData
	{
		private long win;
		private long tie;
		private long total;

		public WinRateData(long win, long tie, long total)
		{
			this.win = win;
			this.tie = tie;
			this.total = total;
		}

		public long getWin()
		{
			return win;
		}

		public long getTie()
		{
			return tie;
		}

		public long getTotal()
		{
			return total;
		}
	}

	public static ConcurrentHashMap<Long, KeyData> colorMap = new ConcurrentHashMap<>();
	public static ConcurrentHashMap<Long, KeyData> normalMap = new ConcurrentHashMap<>();
	// winRateMap[n] holds the exact counts for n = 0..2 public cards.
	@SuppressWarnings("unchecked")
	public static ConcurrentHashMap<Long, WinRateData>[] winRateMap = (ConcurrentHashMap<Long, WinRateData>[]) new ConcurrentHashMap[3];

	public static void main(String[] args)
	{
		gen();
		genExtra();
		genOpt();
		genExtraOpt();
		genTrans();
		genTransOpt();
		genHand();
		System.out.println("done!");
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

	private static void genExtra()
	{
		File file = new File("texas_data_extra_6.txt");
		if (!file.exists())
		{
			GenExtraUtil.N = 6;
			GenExtraUtil.genKey();
			GenExtraUtil.outputData();
		}
		file = new File("texas_data_extra_5.txt");
		if (!file.exists())
		{
			GenExtraUtil.N = 5;
			GenExtraUtil.genKey();
			GenExtraUtil.outputData();
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

	private static void genExtraOpt()
	{
		File file = new File("texas_data_extra_6.txt");
		File file1 = new File("texas_data_extra_color_6.txt");
		File file2 = new File("texas_data_extra_normal_6.txt");
		if (file.exists() && !file1.exists() && !file2.exists())
		{
			GenExtraOptUtil.N = 6;
			GenExtraOptUtil.optNormalData();
			GenExtraOptUtil.optColorData();
		}
		file = new File("texas_data_extra_5.txt");
		file1 = new File("texas_data_extra_color_5.txt");
		file2 = new File("texas_data_extra_normal_5.txt");
		if (file.exists() && !file1.exists() && !file2.exists())
		{
			GenExtraOptUtil.N = 5;
			GenExtraOptUtil.optNormalData();
			GenExtraOptUtil.optColorData();
		}
	}

	private static void genHand()
	{
		for (int i = 0; i <= 4; i++)
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

	public static boolean isLoaded()
	{
		return !normalMap.isEmpty() && !colorMap.isEmpty();
	}

	public static boolean isProbabilityLoaded()
	{
		for (ConcurrentHashMap<Long, WinRateData> m : winRateMap)
		{
			if (m == null || m.isEmpty())
			{
				return false;
			}
		}
		return true;
	}

	public static void load()
	{
		load(new File("."));
	}

	public static void load(String dirPath)
	{
		load(new File(dirPath));
	}

	public static void load(File dir)
	{
		try
		{
			long begin = System.currentTimeMillis();
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_color.txt")))
			{
				loadColor(inputStream);
			}
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_normal.txt")))
			{
				loadNormal(inputStream);
			}
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_extra_color_6.txt")))
			{
				loadColor(inputStream);
			}
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_extra_normal_6.txt")))
			{
				loadNormal(inputStream);
			}
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_extra_color_5.txt")))
			{
				loadColor(inputStream);
			}
			try (FileInputStream inputStream = new FileInputStream(new File(dir, "texas_data_extra_normal_5.txt")))
			{
				loadNormal(inputStream);
			}
			System.out.println("load time " + (System.currentTimeMillis() - begin));
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static void loadProbility()
	{
		loadProbility(new File("."));
	}

	public static void loadProbility(String dirPath)
	{
		loadProbility(new File(dirPath));
	}

	// loadProbility loads the exact win-rate tables. The three tables are
	// independent: a missing texas_data_winrate_N.txt only makes the matching
	// N-card lookup throw, everything else keeps working.
	public static void loadProbility(File dir)
	{
		long begin = System.currentTimeMillis();
		int loaded = 0;
		for (int i = 0; i <= 2; i++)
		{
			File file = new File(dir, "texas_data_winrate_" + i + ".txt");
			if (!file.exists())
			{
				System.out.println("missing " + file.getPath());
				continue;
			}
			try (FileInputStream inputStream = new FileInputStream(file))
			{
				loadWinRateFile(i, inputStream);
				loaded++;
			}
			catch (Exception e)
			{
				e.printStackTrace();
			}
		}
		System.out.println("load time " + (System.currentTimeMillis() - begin) + " tables " + loaded);
	}

	public static void loadWinRateFile(int i, InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

		winRateMap[i] = new ConcurrentHashMap<>();

		String str = null;
		while ((str = bufferedReader.readLine()) != null)
		{
			String[] params = str.split(" ");
			if (params.length < 4)
			{
				continue;
			}
			long key = Long.parseLong(params[0]);
			long win = Long.parseLong(params[1]);
			long tie = Long.parseLong(params[2]);
			long total = Long.parseLong(params[3]);
			winRateMap[i].put(key, new WinRateData(win, tie, total));
		}
		bufferedReader.close();
	}

	public static void loadNormal(InputStream inputStream) throws Exception
	{
		BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

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

	//校验单个牌字节：合法牌为花色 0..3、点数 2..14；旧版鬼牌字节 88（花色 5）非法。
	//非法输入抛 IllegalArgumentException，不再像旧版那样静默返回 0
	public static void checkPoke(byte poke)
	{
		if (!Poke.isLegalByte(poke))
		{
			throw new IllegalArgumentException("无法识别的牌: 牌字节 " + poke);
		}
	}

	public static void checkPokes(byte[] pokes)
	{
		for (byte poke : pokes)
		{
			checkPoke(poke);
		}
	}

	private static void checkPokes(List<Byte> pokes)
	{
		for (Byte poke : pokes)
		{
			checkPoke(poke);
		}
	}

	//校验 key 中编码的每张牌，key 含非法牌（如 88）时抛 IllegalArgumentException。
	//这里直接按十进制位解析，避免为每次比较分配列表
	private static void checkKey(long key)
	{
		long rest = key;
		while (rest > 0)
		{
			byte poke = (byte) (rest % 100L);
			if (!Poke.isLegalByte(poke))
			{
				throw new IllegalArgumentException("无法识别的牌: 牌 key " + key);
			}
			rest /= 100L;
		}
	}

	//解析点数部分（2-10/J/Q/K/A）。未知点数（如"鬼"、"1"、"15"）抛 IllegalArgumentException，
	//不再像旧版那样交给 NumberFormatException 或静默返回非法值
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
			byte value;
			try
			{
				value = Byte.parseByte(str);
			}
			catch (NumberFormatException e)
			{
				throw new IllegalArgumentException("无法识别的牌: 牌面 \"" + str + "\"", e);
			}
			if (value < Poke.PokeValue_2 || value > Poke.PokeValue_A)
			{
				throw new IllegalArgumentException("无法识别的牌: 牌面 \"" + str + "\"");
			}
			return value;
		}
	}

	//把牌面字符串（如"方A"）转成牌字节。只支持 52 张牌堆：空串、未知花色/点数以及已删除的"鬼"
	//都会抛 IllegalArgumentException，不再像旧版那样静默返回 0
	public static byte strToPoke(String str)
	{
		if (str == null || str.length() == 0)
		{
			throw new IllegalArgumentException("无法识别的牌: 空牌面");
		}
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
			throw new IllegalArgumentException("无法识别的牌: 牌面 \"" + str + "\"");
		}
	}

	public static String keyToStr(long key)
	{
		return GenUtil.toString(key);
	}

	public static List<Byte> keyToPoke(long k)
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

	//把逗号分隔的牌面字符串转成牌字节列表。任意一张牌无法识别（如"鬼"）都会抛 IllegalArgumentException
	public static List<Byte> strToPokes(String str)
	{
		List<Byte> ret = new ArrayList<>();
		if (str.length() == 0)
		{
			return ret;
		}
		String[] strs = str.split(",");
		for (String s : strs)
		{
			ret.add(strToPoke(s));
		}
		return ret;
	}

	public static String pokesToStr(List<Byte> pokes)
	{
		return keyToStr(GenUtil.genCardBind(pokes));
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

	//按 key 查询牌力数据。key 含非法牌（如鬼牌字节 88）或查不到表项时抛 IllegalArgumentException
	public static KeyData getKeyData(long key)
	{
		checkKey(key);
		KeyData keyData = lookupKeyData(key);
		if (keyData == null)
		{
			throw new IllegalArgumentException("无法识别的牌: 未找到 key " + key + " 的牌力数据");
		}
		return keyData;
	}

	//只做查表，不校验 key；供内部热路径（排序、生成）使用，未命中时返回 null
	static KeyData lookupKeyData(long key)
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
			return color;
		}
		if (color.getIndex() > normal.getIndex())
		{
			return color;
		}
		else
		{
			return normal;
		}
	}

	public static KeyData getKeyData(List<Byte> pokes)
	{
		checkPokes(pokes);
		long key = GenUtil.genCardBind(pokes);

		return getKeyData(key);
	}

	public static String getMax(String str)
	{
		return GenUtil.toString(GenUtil.genCardBind(getMax(strToPokes(str))));
	}

	public static List<Byte> getMax(List<Byte> hand, List<Byte> pub)
	{
		List<Byte> ret = new ArrayList<>();
		if (hand.size() != 2)
		{
			throw new IllegalArgumentException("无法识别的牌: 手牌需要 2 张，实际 " + hand.size() + " 张");
		}
		if (pub.size() < 3 || pub.size() > 5)
		{
			throw new IllegalArgumentException("无法识别的牌: 公牌需要 3-5 张，实际 " + pub.size() + " 张");
		}
		checkPokes(hand);
		checkPokes(pub);
		List<Byte> tmp = new ArrayList<>();
		tmp.addAll(hand);
		tmp.addAll(pub);
		KeyData keyData = getKeyData(tmp);

		List<Byte> max = keyToPoke(keyData.max);

		List<Byte> pubtmp = new ArrayList<>();
		pubtmp.addAll(pub);
		List<Byte> handtmp = new ArrayList<>();
		handtmp.addAll(hand);

		if (keyData.type == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA
				|| keyData.type == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN
				|| keyData.type == TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN)
		{
			int[] srccolor = new int[4];
			//入参已校验为 52 张合法牌（花色 0..3），可直接统计，不会数组越界
			for (int i = 0; i < tmp.size(); i++)
			{
				srccolor[tmp.get(i) >> 4]++;
			}

			int srcmaxColor = 0;
			int srcmaxColorNum = 0;
			for (int i = 0; i < srccolor.length; i++)
			{
				if (srccolor[i] >= srcmaxColorNum)
				{
					srcmaxColor = i;
					srcmaxColorNum = srccolor[i];
				}
			}

			for (int i = 0; i < max.size(); i++)
			{
				for (int j = 0; j < pubtmp.size(); j++)
				{
					if ((pubtmp.get(j) % 16) == (max.get(i) % 16) && (pubtmp.get(j) >> 4) == srcmaxColor
							&& max.get(i) != 0 && pubtmp.get(j) != 0)
					{
						ret.add(pubtmp.get(j));

						max.set(i, (byte) 0);
						pubtmp.set(j, (byte) 0);
						break;
					}
				}
			}

			if (ret.size() < 5)
			{
				for (int i = 0; i < max.size(); i++)
				{
					for (int j = 0; j < handtmp.size(); j++)
					{
						if ((handtmp.get(j) % 16) == (max.get(i) % 16) && (handtmp.get(j) >> 4) == srcmaxColor
								&& max.get(i) != 0 && handtmp.get(j) != 0)
						{
							ret.add(handtmp.get(j));

							max.set(i, (byte) 0);
							handtmp.set(j, (byte) 0);
							break;
						}
					}
				}
			}

			for (int i = 0; i < max.size(); i++)
			{
				if (max.get(i) != 0)
				{
					max.set(i, (byte) (srcmaxColor << 4 | (max.get(i) % 16)));
				}
			}
		}
		else
		{
			for (int j = 0; j < pubtmp.size(); j++)
			{
				for (int i = 0; i < max.size(); i++)
				{
					if (pubtmp.get(j) == max.get(i) && max.get(i) != 0 && pubtmp.get(j) != 0)
					{
						ret.add(pubtmp.get(j));

						max.set(i, (byte) 0);
						pubtmp.set(j, (byte) 0);
						break;
					}
				}
			}

			for (int j = 0; j < handtmp.size(); j++)
			{
				for (int i = 0; i < max.size(); i++)
				{
					if (handtmp.get(j) == max.get(i) && max.get(i) != 0 && handtmp.get(j) != 0)
					{
						ret.add(handtmp.get(j));

						max.set(i, (byte) 0);
						handtmp.set(j, (byte) 0);
						break;
					}
				}
			}

			for (int i = 0; i < max.size(); i++)
			{
				for (int j = 0; j < pubtmp.size(); j++)
				{
					if ((pubtmp.get(j) % 16) == (max.get(i) % 16) && max.get(i) != 0 && pubtmp.get(j) != 0)
					{
						ret.add(pubtmp.get(j));

						max.set(i, (byte) 0);
						pubtmp.set(j, (byte) 0);
						break;
					}
				}
			}

			if (ret.size() < 5)
			{
				for (int i = 0; i < max.size(); i++)
				{
					for (int j = 0; j < handtmp.size(); j++)
					{
						if ((handtmp.get(j) % 16) == (max.get(i) % 16) && max.get(i) != 0 && handtmp.get(j) != 0)
						{
							ret.add(handtmp.get(j));

							max.set(i, (byte) 0);
							handtmp.set(j, (byte) 0);
							break;
						}
					}
				}
			}

		}

		if (ret.isEmpty())
		{
			throw new IllegalArgumentException("无法识别的牌: 未找到这手牌的牌力数据");
		}

		Collections.sort(ret);

		return ret;
	}

	public static List<Byte> getMax(List<Byte> pokes)
	{
		if (pokes.size() < 5 || pokes.size() > 7)
		{
			throw new IllegalArgumentException("无法识别的牌: 需要 5-7 张牌，实际 " + pokes.size() + " 张");
		}
		checkPokes(pokes);
		List<Byte> hand = new ArrayList<>();
		hand.add(pokes.get(0));
		hand.add(pokes.get(1));

		List<Byte> pub = new ArrayList<>();
		for (int i = 2; i < pokes.size(); i++)
		{
			pub.add(pokes.get(i));
		}

		return getMax(hand, pub);
	}

	public static String getMax(String hand, String pub)
	{
		return GenUtil.toString(GenUtil.genCardBind(getMax(strToPokes(hand), strToPokes(pub))));
	}

	public static int getWinPosition(String str)
	{
		return getWinPosition(strToPokes(str));
	}

	public static int getWinPosition(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		return keyData.getPostion();
	}

	public static double getWinProbability(String str)
	{
		return getWinProbability(strToPokes(str));
	}

	public static double getWinProbability(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		long total = 1;
		for (int i = 0; i < pokes.size(); i++)
		{
			total = total * (GenUtil.genNum - i);
		}
		for (int i = pokes.size(); i >= 1; i--)
		{
			total = total / i;
		}
		return (double) keyData.getIndex() / total;
	}

	public static long getWinMax(String str)
	{
		return getWinMax(strToPokes(str));
	}

	public static long getWinMax(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		return keyData.getMax();
	}

	public static int getWinType(String str)
	{
		return getWinType(strToPokes(str));
	}

	public static int getWinType(List<Byte> pokes)
	{
		KeyData keyData = getKeyData(pokes);
		return keyData.getType();
	}

	public static int compare(String str1, String str2)
	{
		return compare(strToPokes(str1), strToPokes(str2));
	}

	//比较两手牌。任意一手含非法牌（如鬼牌字节 88）时抛 IllegalArgumentException
	public static int compare(List<Byte> bytes1, List<Byte> bytes2)
	{
		return compare(GenUtil.genCardBind(bytes1), GenUtil.genCardBind(bytes2));
	}

	public static int compare(long k1, long k2)
	{
		checkKey(k1);
		checkKey(k2);
		KeyData keyData1 = lookupKeyData(k1);
		KeyData keyData2 = lookupKeyData(k2);
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

	public static float getHandProbability(long hand, long pub)
	{
		checkKey(hand);
		checkKey(pub);
		return getHandProbability(keyToPoke(hand), keyToPoke(pub));
	}

	// canonKey returns the suit-canonical key of a situation made of the first
	// nHole cards plus the remaining public cards. Suits are labelled 0..3 by
	// sorting them on the ranks each suit holds (and which of those are hole
	// cards), which is invariant under the 24 suit permutations. Hole and
	// public cards occupy separate key positions because the roles differ.
	public static long canonKey(List<Byte> cards, int nHole)
	{
		int[] mask = new int[4];
		int[] hole = new int[4];
		for (int i = 0; i < cards.size(); i++)
		{
			byte c = cards.get(i);
			mask[c >> 4] |= 1 << (c & 0x0f);
			if (i < nHole)
			{
				hole[c >> 4] |= 1 << (c & 0x0f);
			}
		}
		int[] label = new int[4];
		boolean[] used = new boolean[4];
		for (int k = 0; k < 4; k++)
		{
			int best = -1;
			int bestSig = 0;
			for (int s = 0; s < 4; s++)
			{
				if (used[s])
				{
					continue;
				}
				int sig = (mask[s] << 16) | hole[s];
				if (best < 0 || sig < bestSig)
				{
					best = s;
					bestSig = sig;
				}
			}
			used[best] = true;
			label[best] = k;
		}

		byte[] holeCards = new byte[5];
		byte[] pubCards = new byte[5];
		int nh = 0;
		int np = 0;
		for (int i = 0; i < cards.size(); i++)
		{
			byte c = cards.get(i);
			byte b = (byte) ((label[c >> 4] << 4) | (c & 0x0f));
			if (i < nHole)
			{
				holeCards[nh] = b;
				nh++;
			}
			else
			{
				pubCards[np] = b;
				np++;
			}
		}
		java.util.Arrays.sort(holeCards, 0, nh);
		java.util.Arrays.sort(pubCards, 0, np);

		long k = 0;
		for (int i = 0; i < nh; i++)
		{
			k = k * 100 + holeCards[i];
		}
		for (int i = 0; i < np; i++)
		{
			k = k * 100 + pubCards[i];
		}
		return k;
	}

	private static long key7Board(byte[] b, byte x, byte y)
	{
		byte[] buf = new byte[7];
		int i = 0;
		int j = 0;
		while (j < 5)
		{
			if (i == 0 && x < b[j])
			{
				buf[i + j] = x;
				i++;
			}
			else if (i == 1 && y < b[j])
			{
				buf[i + j] = y;
				i++;
			}
			else
			{
				buf[i + j] = b[j];
				j++;
			}
		}
		if (i == 0)
		{
			buf[5] = x;
			buf[6] = y;
		}
		else if (i == 1)
		{
			buf[6] = y;
		}
		long k = 0;
		for (int n = 0; n < 7; n++)
		{
			k = k * 100 + buf[n];
		}
		return k;
	}

	// equityExact enumerates every completion of the community cards and every
	// opponent hole-card pair, returning {win, tie, loss}.
	public static long[] equityExact(List<Byte> hand, List<Byte> pub)
	{
		List<Byte> known = new ArrayList<>(hand);
		known.addAll(pub);
		List<Byte> rest = new ArrayList<>();
		for (int color = 0; color < 4; color++)
		{
			for (int v = 2; v <= 14; v++)
			{
				byte c = (byte) ((color << 4) | v);
				if (!known.contains(c))
				{
					rest.add(c);
				}
			}
		}

		int need = 5 - pub.size();
		byte[] board = new byte[5];
		for (int i = 0; i < pub.size(); i++)
		{
			board[i] = pub.get(i);
		}
		List<byte[]> boards = new ArrayList<>();
		genBoards(boards, board, pub.size(), rest, 0, 0, need);

		byte hx = hand.get(0);
		byte hy = hand.get(1);
		if (hx > hy)
		{
			byte t = hx;
			hx = hy;
			hy = t;
		}
		final byte heroA = hx;
		final byte heroB = hy;

		int workers = 1;
		if (boards.size() * 990L >= 100000L)
		{
			workers = Runtime.getRuntime().availableProcessors();
		}
		if (workers < 1)
		{
			workers = 1;
		}
		long[][] results = new long[workers][3];
		int chunk = (boards.size() + workers - 1) / workers;
		Thread[] threads = new Thread[workers];
		for (int w = 0; w < workers; w++)
		{
			final int slot = w;
			final int from = w * chunk;
			final int to = Math.min(boards.size(), from + chunk);
			threads[w] = new Thread(() -> {
				if (from >= to)
				{
					return;
				}
				long[] acc = results[slot];
				for (int idx = from; idx < to; idx++)
				{
					byte[] b = boards.get(idx);
					byte[] opp = new byte[45];
					int n = 0;
					for (byte c : rest)
					{
						if (c == b[0] || c == b[1] || c == b[2] || c == b[3] || c == b[4])
						{
							continue;
						}
						opp[n] = c;
						n++;
					}
					KeyData hero = lookupKeyData(key7Board(b, heroA, heroB));
					if (hero == null)
					{
						continue;
					}
					for (int i = 0; i < n; i++)
					{
						for (int j = i + 1; j < n; j++)
						{
							KeyData data = lookupKeyData(key7Board(b, opp[i], opp[j]));
							if (data == null)
							{
								continue;
							}
							if (hero.getIndex() > data.getIndex())
							{
								acc[0]++;
							}
							else if (hero.getIndex() == data.getIndex())
							{
								acc[1]++;
							}
							else
							{
								acc[2]++;
							}
						}
					}
				}
			});
			threads[w].start();
		}
		for (Thread t : threads)
		{
			try
			{
				t.join();
			}
			catch (InterruptedException e)
			{
				Thread.currentThread().interrupt();
			}
		}
		long[] counts = new long[3];
		for (long[] r : results)
		{
			counts[0] += r[0];
			counts[1] += r[1];
			counts[2] += r[2];
		}
		return counts;
	}

	private static void genBoards(List<byte[]> boards, byte[] board, int boardLen, List<Byte> rest,
			int start, int depth, int need)
	{
		if (depth == need)
		{
			byte[] b = board.clone();
			java.util.Arrays.sort(b);
			boards.add(b);
			return;
		}
		for (int i = start; i <= rest.size() - (need - depth); i++)
		{
			board[boardLen + depth] = rest.get(i);
			genBoards(boards, board, boardLen, rest, i + 1, depth + 1, need);
		}
	}

	public static float getHandProbability(List<Byte> hand, List<Byte> pub)
	{
		checkPokes(hand);
		checkPokes(pub);
		if (hand.size() != 2 || pub.size() > 5)
		{
			throw new IllegalArgumentException("无法识别的牌: 手牌必须为 2 张，公牌最多 5 张");
		}
		if (pub.size() >= 3)
		{
			long[] counts = equityExact(hand, pub);
			long total = counts[0] + counts[1] + counts[2];
			if (total == 0)
			{
				throw new IllegalArgumentException("缺少胜率表数据");
			}
			return (float) ((double) (2 * counts[0] + counts[1]) / (double) (2 * total));
		}
		ConcurrentHashMap<Long, WinRateData> m = winRateMap[pub.size()];
		if (m == null || m.isEmpty())
		{
			throw new IllegalArgumentException("缺少胜率表数据");
		}
		List<Byte> cards = new ArrayList<>(hand);
		cards.addAll(pub);
		WinRateData data = m.get(canonKey(cards, 2));
		if (data == null || data.getTotal() == 0)
		{
			throw new IllegalArgumentException("缺少胜率表数据");
		}
		return (float) ((double) (2 * data.getWin() + data.getTie()) / (double) (2 * data.getTotal()));
	}
}
