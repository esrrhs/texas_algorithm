package com.github.esrrhs.texas_algorithm;

import java.io.*;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashSet;

public class GenOptUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static HashSet<Long> keys = new HashSet<>();

	public static void optColorData()
	{
		try
		{
			FileInputStream inputStream = new FileInputStream("texas_data.txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_color.txt");
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
				String[] params = str.split(" ");
				long key = Long.parseLong(params[0]);
				long i = Long.parseLong(params[1]);
				long index = Long.parseLong(params[2]);
				long total = Long.parseLong(params[3]);
				String keystr = params[4];
				long max = Long.parseLong(params[5]);
				String maxstr = params[6];
				long maxType = Long.parseLong(params[7]);
				if (maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN)
				{
					str = key + " " + i + " " + index + " " + keystr + " " + max + " " + maxstr + " " + maxType + "\n";
					out.write(str.getBytes("utf-8"));
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / GenUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println(cur + "% 需要" + per * (GenUtil.total - totalKey) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			out.close();

			System.out.println("optData finish " + totalKey);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static void optNormalData()
	{
		try
		{
			FileInputStream inputStream = new FileInputStream("texas_data.txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_normal.txt");
			if (file.exists())
			{
				file.delete();
			}
			file.createNewFile();
			out = new FileOutputStream(file, true);

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();
			keys.clear();

			String str = null;
			long lastKey = 0;
			while ((str = bufferedReader.readLine()) != null)
			{
				String[] params = str.split(" ");
				long key = Long.parseLong(params[0]);
				long i = Long.parseLong(params[1]);
				long index = Long.parseLong(params[2]);
				long total = Long.parseLong(params[3]);
				String keystr = params[4];
				long max = Long.parseLong(params[5]);
				String maxstr = params[6];
				long maxType = Long.parseLong(params[7]);
				long removeKey = removeColor(key);
				if (!(maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN))
				{
					if (lastKey != removeKey && !keys.contains(removeKey))
					{
						str = removeKey + " " + i + " " + index + " " + keystr + " " + max + " " + maxstr + " "
								+ maxType + "\n";
						out.write(str.getBytes("utf-8"));
						lastKey = removeKey;
						keys.add(removeKey);
					}
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / GenUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println(cur + "% 需要" + per * (GenUtil.total - totalKey) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			out.close();
			keys.clear();

			System.out.println("optData finish " + totalKey);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

	public static long removeColor(long k)
	{
		ArrayList<Integer> cs = new ArrayList<>();
		cs.add((int) (k % 100000000000000L / 1000000000000L) % 16);
		cs.add((int) (k % 1000000000000L / 10000000000L) % 16);
		cs.add((int) (k % 10000000000L / 100000000L) % 16);
		cs.add((int) (k % 100000000L / 1000000L) % 16);
		cs.add((int) (k % 1000000L / 10000L) % 16);
		cs.add((int) (k % 10000L / 100L) % 16);
		cs.add((int) (k % 100L / 1L) % 16);

		Collections.sort(cs);

		long ret = 0;
		for (Integer p : cs)
		{
			ret = ret * 100 + p;
		}
		return ret;
	}

}
