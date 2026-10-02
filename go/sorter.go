package texas_algorithm

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Quicksort sorts the keys ascending by poker strength (GenCompare) using a
// bounded parallel quicksort, mirroring Java's Sorter. Ties (equal strength,
// different keys) may end up in any order, which does not change the rank
// each key maps to in the generated table.
func Quicksort(input []int64) {
	if len(input) <= 1 {
		return
	}

	nThreads := runtime.NumCPU()
	fallback := 2 * nThreads
	total := int64(len(input))

	var wg sync.WaitGroup
	sem := make(chan struct{}, nThreads)
	var inFlight atomic.Int64

	var quicksort func(layer, lowerIndex, higherIndex int)
	quicksort = func(layer, lowerIndex, higherIndex int) {
		if higherIndex < lowerIndex {
			return
		}
		if higherIndex == lowerIndex {
			Progress.Add(1)

			step := Progress.Load()
			cur := step * 10000 / total
			if cur != LastPrint.Load() {
				LastPrint.Store(cur)

				now := time.Now().UnixMilli()
				per := float64(now-BeginPrint.Load()) / float64(step)
				fmt.Println(fmt.Sprint(cur) + "%% 需要" + fmt.Sprint(per*float64(total-step)/60/1000) + "分" +
					" 用时" + fmt.Sprint((now-BeginPrint.Load())/60/1000) + "分" + " 速度" +
					fmt.Sprint(float64(step)/(float64(now-BeginPrint.Load())/1000)) + "条/秒")
			}
			return
		}

		i := lowerIndex
		j := higherIndex
		// calculate pivot number, I am taking pivot as middle index number
		pivot := input[lowerIndex+(higherIndex-lowerIndex)/2]

		totalStep := j - i
		if totalStep == 0 {
			totalStep = 1
		}
		lastPrint := 0
		beginPrint := time.Now().UnixMilli()

		for i <= j {
			// In each iteration, find a number from the left side greater than
			// the pivot, and a number from the right side less than the pivot,
			// then exchange them.
			for GenCompare(input[i], pivot) {
				i++
			}
			for GenCompare(pivot, input[j]) {
				j--
			}
			if i <= j {
				input[i], input[j] = input[j], input[i]
				// move index to next position on both sides
				i++
				j--
			}

			if i <= j && totalStep > 100000 {
				step := totalStep - (j - i)
				if step < 0 {
					step = 0
				}
				cur := step * 100 / totalStep
				if cur != lastPrint {
					lastPrint = cur

					now := time.Now().UnixMilli()
					per := float64(now-beginPrint) / float64(step)
					fmt.Println(fmt.Sprint(layer) + "/" + fmt.Sprint(int(math.Log(float64(total)))) + "层 " +
						fmt.Sprint(cur) + "% 需要" + fmt.Sprint(per*float64(totalStep-step)/60/1000) + "分" +
						" 用时" + fmt.Sprint((now-beginPrint)/60/1000) + "分" + " 速度" +
						fmt.Sprint(float64(step)/(float64(now-beginPrint)/1000)) + "条/秒")
				}
			}
		}

		spawn := func(lo, hi int) {
			inFlight.Add(1)
			wg.Add(1)
			go func() {
				sem <- struct{}{}
				quicksort(layer+1, lo, hi)
				<-sem
				inFlight.Add(-1)
				wg.Done()
			}()
		}

		// call quickSort recursively, inline or in new goroutines
		if int(inFlight.Load()) >= fallback {
			if i-j == 1 {
				quicksort(layer+1, lowerIndex, j)
				quicksort(layer+1, i, higherIndex)
			} else {
				quicksort(layer+1, lowerIndex, j+1)
				quicksort(layer+1, i, higherIndex)
			}
		} else {
			if i-j == 1 {
				spawn(lowerIndex, j)
				spawn(i, higherIndex)
			} else {
				spawn(lowerIndex, j+1)
				spawn(i, higherIndex)
			}
		}
	}

	BeginPrint.Store(time.Now().UnixMilli())
	LastPrint.Store(0)

	inFlight.Add(1)
	wg.Add(1)
	go func() {
		sem <- struct{}{}
		quicksort(0, 0, len(input)-1)
		<-sem
		inFlight.Add(-1)
		wg.Done()
	}()

	wg.Wait()
}
