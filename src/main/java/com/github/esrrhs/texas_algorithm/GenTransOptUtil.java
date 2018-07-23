package com.github.esrrhs.texas_algorithm;

import java.io.*;
import java.util.HashMap;
import java.util.Map;

public class GenTransOptUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static final long genNum = GenUtil.genNum;
	public static int N = 6;
	public static long total = 1;
	public static HashMap<Long, String> keys = new HashMap<>();

	public static HashMap<Long, OptKeyData> optkeys = new HashMap<>();

	public static class OptKeyData
	{
		String max;
		HashMap<String, Integer> ps = new HashMap<>();
	}

	public static void optData()
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

			FileInputStream inputStream = new FileInputStream("texas_data_" + N + ".txt");
			BufferedReader bufferedReader = new BufferedReader(new InputStreamReader(inputStream));

			File file = new File("texas_data_opt_" + N + ".txt");
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
			optkeys.clear();

			String str = null;
			while ((str = bufferedReader.readLine()) != null)
			{
				String[] params = str.split(" ");
				long key = Long.parseLong(params[0]);
				String p = params[1];

				keys.put(key, p);
				long removeKey = GenOptUtil.removeColor(key);
				OptKeyData optKeyData = optkeys.get(removeKey);
				if (optKeyData == null)
				{
					optKeyData = new OptKeyData();
					optkeys.put(removeKey, optKeyData);
				}
				Integer num = optKeyData.ps.get(p);
				if (num != null)
				{
					optKeyData.ps.put(p, num + 1);
				}
				else
				{
					optKeyData.ps.put(p, 1);
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println("N" + N + " " + cur + "% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			for (Map.Entry<Long, OptKeyData> e : optkeys.entrySet())
			{
				OptKeyData optKeyData = e.getValue();
				int max = 0;
				for (Map.Entry<String, Integer> e1 : optKeyData.ps.entrySet())
				{
					if (e1.getValue() > max)
					{
						optKeyData.max = e1.getKey();
					}
				}
			}

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();

			for (Map.Entry<Long, String> e : keys.entrySet())
			{
				long key = e.getKey();
				String p = e.getValue();
				long removeKey = GenOptUtil.removeColor(key);
				OptKeyData optKeyData = optkeys.get(removeKey);
				if (optKeyData == null || !optKeyData.max.equals(p))
				{
					String tmp = key + " 0 " + p + "\n";
					out.write(tmp.getBytes("utf-8"));
				}

				totalKey++;

				int cur = (int) (totalKey * 100 / total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / totalKey;
					System.out.println("N" + N + " " + cur + "% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度"
							+ totalKey / ((float) (now - beginPrint) / 1000) + "条/秒");
				}
			}

			for (Map.Entry<Long, OptKeyData> e : optkeys.entrySet())
			{
				long key = e.getKey();
				OptKeyData optKeyData = e.getValue();
				String tmp = key + " 1 " + optKeyData.max + "\n";
				out.write(tmp.getBytes("utf-8"));
			}
			out.close();

			System.out.println("optData finish " + totalKey);
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}

	}
}
