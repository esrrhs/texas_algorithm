package com.github.esrrhs.texas_algorithm;

import java.io.File;
import java.io.FileOutputStream;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;

public class GenHandUtil
{
	public static long totalKey = 0;
	public static int lastPrint = 0;
	public static long beginPrint;
	public static final long genNum = GenUtil.genNum;
	public static int N = 6;
	public static long totalhand = genNum * (genNum - 1) / 2;
	public static long totalpub = 1;
	public static long total = 1;
	public static long totalcalc = 1;
	public static final int N_THREADS = Runtime.getRuntime().availableProcessors();
	public static final int FALLBACK = 2;
	public static Executor pool;
	public static AtomicInteger count;

	public static class CalcData
	{
		long win;
		long lose;
		long tie;
		long total;
	}

	public static void genKey()
	{
		try
		{
			totalpub = 1;
			for (int i = 0; i < N; i++)
			{
				totalpub = totalpub * (genNum - 2 - i);
			}
			for (int i = N; i >= 1; i--)
			{
				totalpub = totalpub / i;
			}
			total = totalpub * totalhand;

			totalcalc = 1;
			for (int i = 0; i < 5 - N; i++)
			{
				totalcalc = totalcalc * (genNum - 2 - N - i);
			}
			for (int i = 5 - N; i >= 1; i--)
			{
				totalcalc = totalcalc / i;
			}
			for (int i = 0; i < 2; i++)
			{
				totalcalc = totalcalc * (genNum - 7 - i);
			}
			for (int i = 2; i >= 1; i--)
			{
				totalcalc = totalcalc / i;
			}

			TexasAlgorithmUtil.load();

			File dir = new File("hand" + N + "/");
			if (!dir.exists())
			{
				dir.mkdirs();
			}

			beginPrint = System.currentTimeMillis();
			totalKey = 0;
			lastPrint = 0;

			pool = Executors.newFixedThreadPool(N_THREADS);
			count = new AtomicInteger(0);

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

		int[] hand = new int[2];
		permutation(list, 0, 0, 2, hand);
	}

	public static void permutation(ArrayList<Integer> a, int count, int count2, int except, int[] hand) throws Exception
	{
		if (count2 == except)
		{
			File file = new File("hand" + N + "/texas_hand_" + GenUtil.toString(hand[0] * 100 + hand[1]) + ".txt");
			if (!file.exists())
			{
				ArrayList<Integer> list1 = new ArrayList<>();
				for (byte i = 0; i < 4; ++i)
				{
					for (byte j = 0; j < genNum / 4; ++j)
					{
						int z = (Integer) (int) (new Poke(i, (byte) (j + 2))).toByte();
						if (hand[0] != z && hand[1] != z)
						{
							list1.add(z);
						}
					}
				}
				Collections.sort(list1);

				file.createNewFile();
				FileOutputStream out = new FileOutputStream(file, true);

				int[] pub = new int[N];
				permutation1(out, list1, 0, 0, N, pub, hand);
			}
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				hand[count2] = a.get(i);
				permutation(a, i + 1, count2 + 1, except, hand);
			}
		}
	}

	public static void permutation1(FileOutputStream out, ArrayList<Integer> a, int count, int count2, int except,
			int[] pub, int[] hand) throws Exception
	{
		if (count2 == except)
		{
			if (GenHandUtil.count.get() >= FALLBACK * N_THREADS)
			{
				genCardSave(out, hand, pub);
			}
			else
			{
				GenHandUtil.count.getAndIncrement();
				final int[] fpub = Arrays.copyOf(pub, pub.length);
				final int[] fhand = Arrays.copyOf(hand, hand.length);
				final FileOutputStream fout = out;
				pool.execute(new Runnable() {
					@Override
					public void run()
					{
						try
						{
							genCardSave(fout, fhand, fpub);
						}
						catch (Exception e)
						{
							e.printStackTrace();
						}
						GenHandUtil.count.getAndDecrement();
					}
				});
			}
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				pub[count2] = a.get(i);
				permutation1(out, a, i + 1, count2 + 1, except, pub, hand);
			}
		}
	}

	private static void genCardSave(FileOutputStream out, int[] hand, int[] pub) throws Exception
	{
		long h = GenUtil.genCardBind(hand);
		long p = GenUtil.genCardBind(pub);
		for (int i = 0; i < N; i++)
		{
			h = h * 100;
		}
		h += p;

		double pp = calc(h, hand, pub);

		String tmp = h + " " + pp + " " + GenUtil.toString(h) + "\n";
		out.write(tmp.getBytes("utf-8"));

		totalKey++;

		int cur = (int) (totalKey * 10000 / total);
		if (cur != lastPrint)
		{
			lastPrint = cur;

			long now = System.currentTimeMillis();
			float per = (float) (now - beginPrint) / totalKey;
			System.out.println("N" + N + " " + cur + "%% 需要" + per * (total - totalKey) / 60 / 1000 + "分" + " 用时"
					+ (now - beginPrint) / 60 / 1000 + "分" + " 速度" + totalKey / ((float) (now - beginPrint) / 1000)
					+ "条/秒");
		}
	}

	private static class CalcShowData
	{
		long num;
		long last;
		long begin;
	}

	private static double calc(long key, int[] hand, int[] pub) throws Exception
	{
		ArrayList<Integer> list = new ArrayList<>();
		for (byte i = 0; i < 4; ++i)
		{
			for (byte j = 0; j < genNum / 4; ++j)
			{
				int z = (Integer) (int) (new Poke(i, (byte) (j + 2))).toByte();
				list.add(z);
			}
		}

		for (Integer p : hand)
		{
			list.remove((Integer) p);
		}
		for (Integer p : pub)
		{
			list.remove((Integer) p);
		}
		Collections.sort(list);

		CalcData data = new CalcData();

		CalcShowData index = new CalcShowData();
		index.begin = System.currentTimeMillis();

		int[] otherhand = new int[2];
		permutation2(data, hand, pub, list, 0, 0, 2, otherhand, index);

		return ((double) data.win + (double) data.tie * 0.5) / data.total;
	}

	public static void permutation2(CalcData data, int[] hand, int[] pub, ArrayList<Integer> a, int count, int count2,
			int except, int[] otherhand, CalcShowData index) throws Exception
	{
		if (count2 == except)
		{
			for (Integer p : otherhand)
			{
				a.remove((Integer) p);
			}
			Collections.sort(a);

			int[] leftpub = new int[5 - N];
			permutation3(data, hand, pub, a, 0, 0, 5 - N, leftpub, otherhand, index);

			for (Integer p : otherhand)
			{
				a.add((Integer) p);
			}
			Collections.sort(a);
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				otherhand[count2] = a.get(i);
				permutation2(data, hand, pub, a, i + 1, count2 + 1, except, otherhand, index);
			}
		}
	}

	public static void permutation3(CalcData data, int[] hand, int[] pub, ArrayList<Integer> a, int count, int count2,
			int except, int[] leftpub, int[] otherhand, CalcShowData index) throws Exception
	{
		if (count2 == except)
		{
			calc(data, hand, pub, leftpub, otherhand, index);
		}
		else
		{
			for (int i = count; i < a.size(); i++)
			{
				leftpub[count2] = a.get(i);
				permutation3(data, hand, pub, a, i + 1, count2 + 1, except, leftpub, otherhand, index);
			}
		}
	}

	public static void calc(CalcData data, int[] hand, int[] pub, int[] leftpub, int[] otherhand, CalcShowData index)
	{
		ArrayList<Byte> my = new ArrayList<>();
		ArrayList<Byte> other = new ArrayList<>();
		for (Integer i : hand)
		{
			my.add((byte) (int) i);
		}
		for (Integer i : pub)
		{
			my.add((byte) (int) i);
		}
		for (Integer i : leftpub)
		{
			my.add((byte) (int) i);
		}

		for (Integer i : otherhand)
		{
			other.add((byte) (int) i);
		}
		for (Integer i : pub)
		{
			other.add((byte) (int) i);
		}
		for (Integer i : leftpub)
		{
			other.add((byte) (int) i);
		}

		int ret = TexasAlgorithmUtil.compare(my, other);
		if (ret == 0)
		{
			data.tie++;
		}
		else if (ret < 0)
		{
			data.lose++;
		}
		else if (ret > 0)
		{
			data.win++;
		}
		data.total++;

		if (totalcalc > 100000000)
		{
			index.num++;
			int cur = (int) (index.num * 1000 / totalcalc);
			if (cur != index.last)
			{
				index.last = cur;

				long now = System.currentTimeMillis();
				float per = (float) (now - index.begin) / index.num;
				System.out.println("N" + N + " " + cur + "‰ 需要" + per * (totalcalc - index.num) / 60 / 1000 + "分"
						+ " 用时" + (now - index.begin) / 60 / 1000 + "分" + " 速度"
						+ index.num / ((float) (now - index.begin) / 1000) + "条/秒");
			}
		}
	}
}
