package com.github.esrrhs.texas_algorithm;

import java.io.*;
import java.util.HashSet;

public class GenExtraOptUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static HashSet<Long> keys = new HashSet<>();
	public static final long genNum = GenUtil.genNum;
	public static long total;
	public static int N = 6;

	public static void optColorData()
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

			FileInputStream inputStream = new FileInputStream("texas_data_extra_" + N + ".txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_extra_color_" + N + ".txt");
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
					long colorKey = GenOptUtil.changeColor(key);

					if (!keys.contains(colorKey))
					{
						str = colorKey + " " + i + " " + index + " " + total + " " + GenUtil.toString(colorKey) + " "
								+ GenUtil.max(colorKey) + " " + GenUtil.toString(GenUtil.max(colorKey)) + " " + maxType
								+ "\n";
						out.write(str.getBytes("utf-8"));
						keys.add(colorKey);
					}
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / GenExtraOptUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println(cur + "% 需要" + per * (GenExtraOptUtil.total - totalKey) / 60 / 1000 + "分" + " 用时"
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

	public static void optNormalData()
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

			FileInputStream inputStream = new FileInputStream("texas_data_extra_" + N + ".txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_extra_normal_" + N + ".txt");
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
				long removeKey = GenOptUtil.removeColor(key);
				if (!(maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUA
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_TONGHUASHUN
						|| maxType == TexasCardUtil.TEXAS_CARD_TYPE_KINGTONGHUASHUN))
				{
					if (!keys.contains(removeKey))
					{
						str = removeKey + " " + i + " " + index + " " + total + " " + keystr + " " + max + " " + maxstr
								+ " " + maxType + "\n";
						out.write(str.getBytes("utf-8"));
						keys.add(removeKey);
					}
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / GenExtraOptUtil.total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println(cur + "% 需要" + per * (GenExtraOptUtil.total - totalKey) / 60 / 1000 + "分" + " 用时"
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

}
