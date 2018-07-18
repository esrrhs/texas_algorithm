package com.github.esrrhs.texas_algorithm;

import java.util.ArrayList;
import java.util.concurrent.Executor;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;

public class Sorter
{
	/**
	 * Number of threads to use for sorting.
	 */
	private static final int N_THREADS = Runtime.getRuntime().availableProcessors();

	/**
	 * Multiple to use when determining when to fall back.
	 */
	private static final int FALLBACK = 2;

	/**
	 * Thread pool used for executing sorting Runnables.
	 */
	private static Executor pool = Executors.newFixedThreadPool(N_THREADS);

	/**
	 * Main method used for sorting from clients. Input is sorted in place using multiple threads.
	 *
	 * @param input The array to sort.
	 */
	public static void quicksort(ArrayList<Long> input)
	{
		final AtomicInteger count = new AtomicInteger(1);
		pool.execute(new QuicksortRunnable(input, 0, input.size() - 1, count, 0, input.size()));
		try
		{
			synchronized (count)
			{
				count.wait();
			}
		}
		catch (InterruptedException e)
		{
			e.printStackTrace();
		}
	}

	/**
	 * Sorts a section of an array using quicksort. The method used is not technically recursive as it just creates new
	 * runnables and hands them off to the ThreadPoolExecutor.
	 *
	 * @param <T> The type of the objects being sorted, must extend Comparable.
	 */
	private static class QuicksortRunnable<T extends Comparable<T>> implements Runnable
	{
		/**
		 * The array being sorted.
		 */
		private final ArrayList<Long> values;
		/**
		 * The starting index of the section of the array to be sorted.
		 */
		private final int left;
		/**
		 * The ending index of the section of the array to be sorted.
		 */
		private final int right;
		/**
		 * The number of threads currently executing.
		 */
		private final AtomicInteger count;

		private final int layer;
		private final int total;

		/**
		 * Default constructor. Sets up the runnable object for execution.
		 *
		 * @param values The array to sort.
		 * @param left   The starting index of the section of the array to be sorted.
		 * @param right  The ending index of the section of the array to be sorted.
		 * @param count  The number of currently executing threads.
		 */
		public QuicksortRunnable(ArrayList<Long> values, int left, int right, AtomicInteger count, int layer, int total)
		{
			this.values = values;
			this.left = left;
			this.right = right;
			this.count = count;
			this.layer = layer;
			this.total = total;
		}

		/**
		 * The thread's run logic. When this thread is done doing its stuff it checks to see if all other threads are as
		 * well. If so, then we notify the count object so Sorter.quicksort stops blocking.
		 */
		@Override
		public void run()
		{
			quicksort(layer, left, right);
			synchronized (count)
			{
				// AtomicInteger.getAndDecrement() returns the old value. If the old value is 1, then we know that the actual value is 0.
				if (count.getAndDecrement() == 1)
					count.notify();
			}
		}

		/**
		 * Method which actually does the sorting. Falls back on recursion if there are a certain number of queued /
		 * running tasks.
		 *
		 */
		private void quicksort(int layer, int lowerIndex, int higherIndex)
		{
			int i = lowerIndex;
			int j = higherIndex;
			// calculate pivot number, I am taking pivot as middle index number
			long pivot = values.get(lowerIndex + (higherIndex - lowerIndex) / 2);
			// Divide into two arrays

			int totalStep = j - i;
			totalStep = totalStep == 0 ? 1 : totalStep;
			int lastPrint = 0;
			long beginPrint = System.currentTimeMillis();

			while (i <= j)
			{
				/**
				 * In each iteration, we will identify a number from left side which
				 * is greater then the pivot value, and also we will identify a number
				 * from right side which is less then the pivot value. Once the search
				 * is done, then we exchange both numbers.
				 */
				while (GenUtil.compare(values.get(i), pivot))
				{
					i++;
				}
				while (GenUtil.compare(pivot, values.get(j)))
				{
					j--;
				}
				if (i <= j)
				{
					long temp = values.get(i);
					values.set(i, values.get(j));
					values.set(j, temp);
					//move index to next position on both sides
					i++;
					j--;
				}

				if (i <= j)
				{
					long step = totalStep - (j - i);
					step = step > 0 ? step : 0;
					int cur = (int) (step * 100 / totalStep);
					if (cur != lastPrint)
					{
						lastPrint = cur;

						long now = System.currentTimeMillis();
						float per = (float) (now - beginPrint) / step;
						System.out.println(layer + "/" + (int) Math.log(total) + "层 " + cur + "% 需要"
								+ per * (totalStep - step) / 60 / 1000 + "分" + " 用时" + (now - beginPrint) / 60 / 1000
								+ "分" + " 速度" + step / ((float) (now - beginPrint) / 1000) + "条/秒");
					}
				}
			}
			// call quickSort() method recursively
			if (count.get() >= FALLBACK * N_THREADS)
			{
				if (lowerIndex < j)
					quicksort(layer + 1, lowerIndex, j);
				if (i < higherIndex)
					quicksort(layer + 1, i, higherIndex);
			}
			else
			{
				if (lowerIndex < j)
				{
					count.getAndAdd(1);
					pool.execute(new QuicksortRunnable<T>(values, lowerIndex, j, count, layer + 1, total));
				}
				if (i < higherIndex)
				{
					count.getAndAdd(1);
					pool.execute(new QuicksortRunnable<T>(values, i, higherIndex, count, layer + 1, total));
				}
			}
		}
	}
}