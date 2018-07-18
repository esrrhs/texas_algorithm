package com.github.esrrhs.texas_algorithm;

import java.io.File;
import java.io.FileOutputStream;
import java.util.ArrayList;
import java.util.Collections;

public class GenUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static final long genNum = 24;
	public static final long total = (genNum * (genNum - 1) * (genNum - 2) * (genNum - 3) * (genNum - 4) * (genNum - 5)
			* (genNum - 6)) / (7 * 6 * 5 * 4 * 3 * 2);
	public static ArrayList<Long> keys = new ArrayList<>((int) total);
	public static ArrayList<Poke> tmp = new ArrayList<>();
	public static ArrayList<Poke> tmp1 = new ArrayList<>();
	public static ArrayList<Poke> tmp2 = new ArrayList<>();

	public static void genKey()
	{
		try
		{
			beginPrint = System.currentTimeMillis();

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

		int[] tmp = new int[7];
		permutation(list, 0, 0, 7, tmp);
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
		final long c = genCardBind(tmp);

		keys.add(c);
		totalKey++;

		int cur = (int) (totalKey * 100 / total);
		if (cur != lastPrint)
		{
			lastPrint = cur;

			long now = System.currentTimeMillis();
			float per = (float) (now - beginPrint) / totalKey;
			System.out.println(
					cur + "% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时" + (now - beginPrint) / 60 / 1000
							+ "分" + " 速度" + totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
		}
	}

	private static long genCardBind(int[] tmp)
	{
		long ret = 0;
		for (Integer i : tmp)
		{
			ret = ret * 100 + i;
		}
		return ret;
	}

	public static void outputData()
	{
		long begin = System.currentTimeMillis();
		try
		{
			File file = new File("texas_data.txt");
			if (file.exists())
			{
				file.delete();
			}
			file.createNewFile();
			out = new FileOutputStream(file, true);

			quickSort(0, 0, keys.size() - 1);

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();
			int i = 0;
			for (Long k : keys)
			{
				String str = k + " " + i + " " + keys.size() + " " + toString(keys.get(i)) + " " + max(keys.get(i))
						+ " " + toString(max(keys.get(i))) + " " + maxType(keys.get(i)) + "\n";

				out.write(str.getBytes("utf-8"));
				i++;
				totalKey++;

				int cur = (int) (totalKey * 100 / total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println(cur + "% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			out.close();

			System.out.println(
					"outputData finish " + total + " time:" + (System.currentTimeMillis() - begin) / 1000 / 60 + "分");
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	private static void quickSort(int layer, int lowerIndex, int higherIndex)
	{
		int i = lowerIndex;
		int j = higherIndex;
		// calculate pivot number, I am taking pivot as middle index number
		long pivot = keys.get(lowerIndex + (higherIndex - lowerIndex) / 2);
		// Divide into two arrays

		int totalStep = j - i;
		totalStep = totalStep == 0 ? 1 : totalStep;
		lastPrint = 0;
		beginPrint = System.currentTimeMillis();

		while (i <= j)
		{
			/**
			 * In each iteration, we will identify a number from left side which
			 * is greater then the pivot value, and also we will identify a number
			 * from right side which is less then the pivot value. Once the search
			 * is done, then we exchange both numbers.
			 */
			while (compare(keys.get(i), pivot))
			{
				i++;
			}
			while (compare(pivot, keys.get(j)))
			{
				j--;
			}
			if (i <= j)
			{
				long temp = keys.get(i);
				keys.set(i, keys.get(j));
				keys.set(j, temp);
				//move index to next position on both sides
				i++;
				j--;
			}

			if (i <= j)
			{
				int step = totalStep - (j - i);
				step = step > 0 ? step : 0;
				int cur = (int) (step * 100 / totalStep);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / step;
					System.out.println(layer + "/" + (int) Math.log(total) + "层 " + cur + "% 需要"
							+ per * (totalStep - step) / 60 / 1000 + "分" + " 用时" + (now - beginPrint) / 60 / 1000 + "分"
							+ " 速度" + step / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}
		}
		// call quickSort() method recursively
		if (lowerIndex < j)
			quickSort(layer + 1, lowerIndex, j);
		if (i < higherIndex)
			quickSort(layer + 1, i, higherIndex);
	}

	public static long max(long k)
	{
		tmp.clear();
		ArrayList<Poke> cs = tmp;
		cs.add(new Poke((byte) (k % 100000000000000L / 1000000000000L)));
		cs.add(new Poke((byte) (k % 1000000000000L / 10000000000L)));
		cs.add(new Poke((byte) (k % 10000000000L / 100000000L)));
		cs.add(new Poke((byte) (k % 100000000L / 1000000L)));
		cs.add(new Poke((byte) (k % 1000000L / 10000L)));
		cs.add(new Poke((byte) (k % 10000L / 100L)));
		cs.add(new Poke((byte) (k % 100L / 1L)));
		tmp1.clear();
		ArrayList<Poke> pickedCards1 = tmp1;
		TexasCardUtil.fiveFromSeven(cs, pickedCards1);

		long ret = 0;
		for (Poke p : pickedCards1)
		{
			ret = ret * 100 + p.toByte();
		}
		return ret;
	}

	public static int maxType(long k)
	{
		tmp.clear();
		ArrayList<Poke> cs = tmp;
		cs.add(new Poke((byte) (k % 100000000000000L / 1000000000000L)));
		cs.add(new Poke((byte) (k % 1000000000000L / 10000000000L)));
		cs.add(new Poke((byte) (k % 10000000000L / 100000000L)));
		cs.add(new Poke((byte) (k % 100000000L / 1000000L)));
		cs.add(new Poke((byte) (k % 1000000L / 10000L)));
		cs.add(new Poke((byte) (k % 10000L / 100L)));
		cs.add(new Poke((byte) (k % 100L / 1L)));
		tmp1.clear();
		ArrayList<Poke> pickedCards1 = tmp1;
		TexasCardUtil.fiveFromSeven(cs, pickedCards1);

		return TexasCardUtil.getCardTypeUnordered(pickedCards1);
	}

	public static String toString(long k)
	{
		tmp.clear();
		ArrayList<Poke> cs = tmp;
		if (k > 1000000000000L)
		{
			cs.add(new Poke((byte) (k % 100000000000000L / 1000000000000L)));
		}
		if (k > 10000000000L)
		{
			cs.add(new Poke((byte) (k % 1000000000000L / 10000000000L)));
		}
		cs.add(new Poke((byte) (k % 10000000000L / 100000000L)));
		cs.add(new Poke((byte) (k % 100000000L / 1000000L)));
		cs.add(new Poke((byte) (k % 1000000L / 10000L)));
		cs.add(new Poke((byte) (k % 10000L / 100L)));
		cs.add(new Poke((byte) (k % 100L / 1L)));
		String ret = "";
		for (Poke poke : cs)
		{
			ret += poke;
		}
		return ret;
	}

	public static boolean compare(long k1, long k2)
	{
		tmp.clear();
		ArrayList<Poke> cs1 = tmp;
		cs1.add(new Poke((byte) (k1 % 100000000000000L / 1000000000000L)));
		cs1.add(new Poke((byte) (k1 % 1000000000000L / 10000000000L)));
		cs1.add(new Poke((byte) (k1 % 10000000000L / 100000000L)));
		cs1.add(new Poke((byte) (k1 % 100000000L / 1000000L)));
		cs1.add(new Poke((byte) (k1 % 1000000L / 10000L)));
		cs1.add(new Poke((byte) (k1 % 10000L / 100L)));
		cs1.add(new Poke((byte) (k1 % 100L / 1L)));
		tmp1.clear();
		ArrayList<Poke> pickedCards1 = tmp1;
		TexasCardUtil.fiveFromSeven(cs1, pickedCards1);

		tmp.clear();
		ArrayList<Poke> cs2 = tmp;
		cs2.add(new Poke((byte) (k2 % 100000000000000L / 1000000000000L)));
		cs2.add(new Poke((byte) (k2 % 1000000000000L / 10000000000L)));
		cs2.add(new Poke((byte) (k2 % 10000000000L / 100000000L)));
		cs2.add(new Poke((byte) (k2 % 100000000L / 1000000L)));
		cs2.add(new Poke((byte) (k2 % 1000000L / 10000L)));
		cs2.add(new Poke((byte) (k2 % 10000L / 100L)));
		cs2.add(new Poke((byte) (k2 % 100L / 1L)));
		tmp2.clear();
		ArrayList<Poke> pickedCards2 = tmp2;
		TexasCardUtil.fiveFromSeven(cs2, pickedCards2);

		return TexasCardUtil.compareCards(pickedCards1, pickedCards2) < 0;
	}

}
