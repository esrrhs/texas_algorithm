package com.github.esrrhs.texas_algorithm;

import java.io.File;
import java.io.FileOutputStream;
import java.util.ArrayList;
import java.util.concurrent.atomic.AtomicInteger;

public class GenExtraUtil
{
	public static long totalKey = 0;
	public static FileOutputStream out;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static final long genNum = GenUtil.genNum;
	public static long total;
	public static int N = 6;
	public static ArrayList<Long> keys = new ArrayList<>();
	public static AtomicInteger progress = new AtomicInteger();
	public static boolean useOpt = GenUtil.useOpt;

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
		ArrayList<Integer> list = GenUtil.genAllCards();

		int[] tmp = new int[N];
		GenUtil.PermutationRun permutationRun = new GenUtil.PermutationRun() {
			@Override
			public void run(int[] tmp, GenUtil.PermutationParam permutationParam) throws Exception
			{
				genCardSave(tmp);
			}
		};
		GenUtil.permutation(permutationRun, list, 0, 0, N, tmp, null);
	}

	private static void genCardSave(int[] tmp) throws Exception
	{
		final long c = GenUtil.genCardBind(tmp);

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

	public static void outputData()
	{
		long begin = System.currentTimeMillis();
		try
		{
			if (useOpt)
			{
				TexasAlgorithmUtil.load();
			}
			File file = new File("texas_data_extra_" + N + ".txt");
			if (file.exists())
			{
				file.delete();
			}
			file.createNewFile();
			out = new FileOutputStream(file, true);

			beginPrint = System.currentTimeMillis();
			lastPrint = 0;

			Sorter.quicksort(keys);

			totalKey = 0;
			lastPrint = 0;
			beginPrint = System.currentTimeMillis();
			int i = 0;
			long iindex = 0;
			long index = 0;
			long lastMax = 0;
			for (Long k : keys)
			{
				long curMax = GenUtil.max(keys.get((int) index));
				String str;
				if (lastMax == 0)
				{
					str = k + " " + i + " " + iindex + " " + keys.size() + " " + GenUtil.toString(keys.get((int) index))
							+ " " + curMax + " " + GenUtil.toString(curMax) + " "
							+ GenUtil.maxType(keys.get((int) index)) + "\n";
					lastMax = curMax;
					iindex = index;
				}
				else
				{
					if (GenUtil.equal(lastMax, curMax))
					{
						str = k + " " + i + " " + iindex + " " + keys.size() + " "
								+ GenUtil.toString(keys.get((int) index)) + " " + curMax + " "
								+ GenUtil.toString(curMax) + " " + GenUtil.maxType(keys.get((int) index)) + "\n";
						lastMax = curMax;
					}
					else
					{
						i++;
						iindex = index;
						str = k + " " + i + " " + iindex + " " + keys.size() + " "
								+ GenUtil.toString(keys.get((int) index)) + " " + curMax + " "
								+ GenUtil.toString(curMax) + " " + GenUtil.maxType(keys.get((int) index)) + "\n";
						lastMax = curMax;
					}
				}

				out.write(str.getBytes("utf-8"));
				index++;

				int cur = (int) (index * 100 / total);
				if (cur != lastPrint)
				{
					lastPrint = cur;

					long now = System.currentTimeMillis();
					float per = (float) (now - beginPrint) / index;
					System.out.println(cur + "% 需要" + per * (total - index) / 60 / 1000 + "分" + " 用时"
							+ (now - beginPrint) / 60 / 1000 + "分" + " 速度" + index / ((float) (now - beginPrint) / 1000)
							+ "条/秒");
				}
			}

			out.close();

			System.out.println("outputData finish " + total + " time:"
					+ (System.currentTimeMillis() - begin) / 1000 / 60 + "分 " + GenExtraUtil.progress);

			keys.clear();
		}
		catch (Exception e)
		{
			e.printStackTrace();
		}
	}

}
