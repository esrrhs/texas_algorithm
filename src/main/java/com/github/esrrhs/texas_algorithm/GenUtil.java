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

	public static void genKey()
	{
		try
		{
			File file = new File("texas_key.txt");
			if (file.exists())
			{
				file.delete();
			}
			file.createNewFile();
			out = new FileOutputStream(file, true);

			beginPrint = System.currentTimeMillis();

			genCard();

			out.close();
			System.out.println("totalKey=" + totalKey);
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
			for (byte j = 2; j < 15; ++j)
			{
				list.add((Integer) (int) (new Poke(i, j)).toByte());
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

		String tmpString = c + "\n";
		out.write(tmpString.getBytes("utf-8"));

		totalKey++;

		final long total = (52L * 51 * 50 * 49 * 48 * 47 * 46) / (7 * 6 * 5 * 4 * 3 * 2);
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
}
