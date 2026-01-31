package waitgroup

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWaitGroup(t *testing.T) {
	wg := NewWaitGroup(5)
	if wg.Limit() != 5 {
		t.Errorf("expected limit = 5, got %d", wg.Limit())
	}
}

func TestLimitWaitGroup_BasicUsage(t *testing.T) {
	wg := NewWaitGroup(3)

	var counter int64
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()

	if counter != 10 {
		t.Errorf("expected counter = 10, got %d", counter)
	}
}

func TestLimitWaitGroup_ConcurrencyLimit(t *testing.T) {
	maxConcurrent := 3
	wg := NewWaitGroup(maxConcurrent)

	var currentConcurrent int64
	var maxObserved int64
	var mu sync.Mutex

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			current := atomic.AddInt64(&currentConcurrent, 1)

			mu.Lock()
			if current > maxObserved {
				maxObserved = current
			}
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)

			atomic.AddInt64(&currentConcurrent, -1)
		}()
	}

	wg.Wait()

	if maxObserved > int64(maxConcurrent) {
		t.Errorf("max concurrent goroutines = %d, exceeded limit of %d", maxObserved, maxConcurrent)
	}
}

func TestLimitWaitGroup_RaceCondition(t *testing.T) {
	wg := NewWaitGroup(5)

	var counter int64

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()

	if counter != 100 {
		t.Errorf("expected counter = 100, got %d", counter)
	}
}

func TestLimitWaitGroup_RaceMultipleAdds(t *testing.T) {
	wg := NewWaitGroup(10)

	var counter int64
	numBatches := 20
	batchSize := 5

	for i := 0; i < numBatches; i++ {
		wg.Add(batchSize)
		for j := 0; j < batchSize; j++ {
			go func() {
				defer wg.Done()
				atomic.AddInt64(&counter, 1)
			}()
		}
	}

	wg.Wait()

	expected := int64(numBatches * batchSize)
	if counter != expected {
		t.Errorf("expected counter = %d, got %d", expected, counter)
	}
}

func TestLimitWaitGroup_RapidAddDone(t *testing.T) {
	wg := NewWaitGroup(2)

	iterations := 1000
	var completed int64

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&completed, 1)
		}()
	}

	wg.Wait()

	if completed != int64(iterations) {
		t.Errorf("expected completed = %d, got %d", iterations, completed)
	}
}

func TestLimitWaitGroup_ConcurrentWaiters(t *testing.T) {
	wg := NewWaitGroup(3)

	var workDone int64
	var waitersFinished int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt64(&workDone, 1)
		}()
	}

	var waiterWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		waiterWg.Add(1)
		go func() {
			defer waiterWg.Done()
			wg.Wait()
			atomic.AddInt64(&waitersFinished, 1)
		}()
	}

	waiterWg.Wait()

	if workDone != 10 {
		t.Errorf("expected workDone = 10, got %d", workDone)
	}
	if waitersFinished != 5 {
		t.Errorf("expected waitersFinished = 5, got %d", waitersFinished)
	}
}

func TestLimitWaitGroup_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	wg := NewWaitGroup(50)

	var counter int64
	numGoroutines := 10000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()

	if counter != int64(numGoroutines) {
		t.Errorf("expected counter = %d, got %d", numGoroutines, counter)
	}
}

func TestLimitWaitGroup_SingleConcurrency(t *testing.T) {
	wg := NewWaitGroup(1)

	var currentConcurrent int64
	var maxObserved int64
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			current := atomic.AddInt64(&currentConcurrent, 1)

			mu.Lock()
			if current > maxObserved {
				maxObserved = current
			}
			mu.Unlock()

			time.Sleep(1 * time.Millisecond)

			atomic.AddInt64(&currentConcurrent, -1)
		}()
	}

	wg.Wait()

	if maxObserved > 1 {
		t.Errorf("max concurrent goroutines = %d, should never exceed 1", maxObserved)
	}
}
