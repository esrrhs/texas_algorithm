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
		ArrayList<Integer> list = GenUtil.genAllCards();

		int[] hand = new int[2];
		GenUtil.PermutationRun permutationRun = new GenUtil.PermutationRun() {
			@Override
			public void run(int[] tmp, GenUtil.PermutationParam permutationParam) throws Exception
			{
				onHandGen(tmp, permutationParam);
			}
		};
		GenUtil.permutation(permutationRun, list, 0, 0, 2, hand, new GenUtil.PermutationParam());
	}

	private static void onHandGen(int[] hand, GenUtil.PermutationParam permutationParam) throws Exception
	{
		permutationParam.o1 = hand;

		File file = new File("hand" + N + "/texas_hand_" + GenUtil.toString(hand[0] * 100 + hand[1]) + ".txt");
		if (!file.exists())
		{
			ArrayList<Integer> list1 = GenUtil.genAllCards();
			list1.remove((Integer) hand[0]);
			list1.remove((Integer) hand[1]);

			file.createNewFile();
			FileOutputStream out = new FileOutputStream(file, true);
			permutationParam.o2 = out;

			int[] pub = new int[N];
			GenUtil.PermutationRun permutationRun = new GenUtil.PermutationRun() {
				@Override
				public void run(int[] tmp, GenUtil.PermutationParam permutationParam) throws Exception
				{
					onPubGen(tmp, permutationParam);
				}
			};
			GenUtil.permutation(permutationRun, list1, 0, 0, N, pub, permutationParam);
		}
	}

	private static void onPubGen(int[] pub, GenUtil.PermutationParam permutationParam) throws Exception
	{
		int[] hand = (int[]) permutationParam.o1;
		FileOutputStream out = (FileOutputStream) permutationParam.o2;

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
		ArrayList<Integer> list = GenUtil.genAllCards();

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

		GenUtil.PermutationRun permutationRun = new GenUtil.PermutationRun() {
			@Override
			public void run(int[] tmp, GenUtil.PermutationParam permutationParam) throws Exception
			{
				onOtherHandGen(tmp, permutationParam);
			}
		};
		int[] otherhand = new int[2];
		GenUtil.PermutationParam permutationParam = new GenUtil.PermutationParam();
		permutationParam.o1 = data;
		permutationParam.o2 = hand;
		permutationParam.o3 = pub;
		permutationParam.o4 = index;
		permutationParam.o5 = list;
		GenUtil.permutation(permutationRun, list, 0, 0, 2, otherhand, permutationParam);

		return ((double) data.win + (double) data.tie * 0.5) / data.total;
	}

	private static void onOtherHandGen(int[] otherhand, GenUtil.PermutationParam permutationParam) throws Exception
	{
		ArrayList<Integer> a = (ArrayList<Integer>) permutationParam.o5;
		for (Integer p : otherhand)
		{
			a.remove((Integer) p);
		}
		Collections.sort(a);

		GenUtil.PermutationRun permutationRun = new GenUtil.PermutationRun() {
			@Override
			public void run(int[] tmp, GenUtil.PermutationParam permutationParam) throws Exception
			{
				onLeftPubGen(tmp, permutationParam);
			}
		};
		int[] leftpub = new int[5 - N];
		permutationParam.o6 = otherhand;
		GenUtil.permutation(permutationRun, a, 0, 0, 5 - N, leftpub, permutationParam);

		for (Integer p : otherhand)
		{
			a.add((Integer) p);
		}
		Collections.sort(a);
	}

	private static void onLeftPubGen(int[] leftpub, GenUtil.PermutationParam permutationParam) throws Exception
	{
		CalcData data = (CalcData) permutationParam.o1;
		int[] hand = (int[]) permutationParam.o2;
		int[] pub = (int[]) permutationParam.o3;
		CalcShowData index = (CalcShowData) permutationParam.o4;
		int[] otherhand = (int[]) permutationParam.o6;

		calc(data, hand, pub, leftpub, otherhand, index);
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
