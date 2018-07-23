package com.github.esrrhs.texas_algorithm;

import java.io.*;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

public class GenTransUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static final long genNum = GenUtil.genNum;
	public static int N = 6;
	public static long total = 1;
	public static HashMap<Long, KeyData> keys = new HashMap<>();

	public static class KeyData
	{
		public long win;
		public long num;
	}

	public static void genKey()
	{
		try
		{
			total = 1;
			for (int i = 0; i < N; i++)
			{
				total = total * (genNum - i);
			}
			for (int i = N; i >= 1; i--)
			{
				total = total / i;
			}
			beginPrint = System.currentTimeMillis();
			keys.clear();
			totalKey = 0;
			lastPrint = 0;

			genCard();

			System.out.println("genKey finish " + total);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	private static void genCard() throws Exception
	{
		ArrayList<Integer> list = new ArrayList<>();
		for (byte i = 0; i < 4; ++i)
		{
			for (byte j = 0; j < genNum / 4; ++j)
			{
				list.add((Integer) (int) (new Poke(i, (byte) (j + 2))).toByte());
			}
		}
		Collections.sort(list);

		int[] tmp = new int[N];
		permutation(list, 0, 0, N, tmp);
	}

	public static void permutation(ArrayList<Integer> a, int count, int count2, int except, int[] tmp) throws Exception
	{
		if (count2 == except)
		{
			genCardSave(tmp);
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				tmp[count2] = a.get(i);
				permutation(a, i + 1, count2 + 1, except, tmp);
			}
		}
	}

	private static void genCardSave(int[] tmp) throws Exception
	{
		final long c = GenUtil.genCardBind(tmp);

		keys.put(c, new KeyData());
		totalKey++;

		int cur = (int) (totalKey * 100 / total);
		if (cur != lastPrint)
		{
			lastPrint = cur;

			long now = System.currentTimeMillis();
			float per = (float) (now - beginPrint) / totalKey;
			System.out.println("N" + N + " " + cur + "% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时"
					+ (now - beginPrint) / 60 / 1000 + "分" + " 速度" + totalKey / ((float) (now - beginPrint) / 1000)
					+ "条/秒");
		}
	}

	public static void transData()
	{
		try
		{
			FileInputStream inputStream = new FileInputStream("texas_data.txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_" + N + ".txt");
			if (file.exists())
			{
				file.delete();
			}
			file.createNewFile();
			out = new FileOutputStream(file, true);

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();

			String str = null;
			while ((str = bufferedReader.readLine()) != null)
			{
				long key = Long.parseLong(str.split(" ")[0]);
				long win = Long.parseLong(str.split(" ")[2]);
				ArrayList<Long> tmp = getKeyList(key);
				for (Long l : tmp)
				{
					KeyData keyData = keys.get(l);
					keyData.win += win;
					keyData.num++;
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / GenUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println("N" + N + " " + cur + "% 需要" + per * (GenUtil.total - totalKey) / 60 / 1000 + "分"
							+ " 用时" + (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();

			for (Map.Entry<Long, KeyData> e : keys.entrySet())
			{
				long key = e.getKey();
				double win = (double) e.getValue().win / e.getValue().num / GenUtil.total;

				String tmp = key + " " + win + "\n";
				out.write(tmp.getBytes("utf-8"));
				totalKey++;

				int cur = (int) (totalKey * 100 / GenUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println("N" + N + " " + cur + "% 需要" + per * (GenUtil.total - totalKey) / 60 / 1000 + "分"
							+ " 用时" + (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			out.close();
			keys.clear();

			System.out.println("transData finish " + totalKey);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static ArrayList<Long> getKeyList(long key) throws Exception
	{
		ArrayList<Long> ret = new ArrayList<>();
		ArrayList<Integer> list = new ArrayList<>();
		while (key > 100)
		{
			list.add((int) (key % 100));
			key = key / 100;
		}
		list.add((int) (key));
		Collections.sort(list);

		int[] tmp = new int[N];
		permutationKey(ret, list, 0, 0, N, tmp);

		return ret;
	}

	public static void permutationKey(ArrayList<Long> ret, ArrayList<Integer> a, int count, int count2, int except,
			int[] tmp) throws Exception
	{
		if (count2 == except)
		{
			long c = GenUtil.genCardBind(tmp);
			if (!ret.contains(c))
			{
				ret.add(c);
			}
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				tmp[count2] = a.get(i);
				permutationKey(ret, a, i + 1, count2 + 1, except, tmp);
			}
		}
	}

}
