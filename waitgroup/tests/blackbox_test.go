package tests

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/bpurdy1/golang-packages/waitgroup"
)

func TestNewLimitWaitGroup(t *testing.T) {
	wg := waitgroup.NewLimitWaitGroup(5)

	if wg.Limit() != 5 {
		t.Errorf("expected limit = 5, got %d", wg.Limit())
	}
}

func TestBasicAddDoneWait(t *testing.T) {
	wg := waitgroup.NewLimitWaitGroup(2)
	counter := int64(0)

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

func TestLimitBlocking(t *testing.T) {
	wg := waitgroup.NewLimitWaitGroup(2)
	running := int64(0)
	maxRunning := int64(0)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			cur := atomic.AddInt64(&running, 1)

			// Track max concurrent
			for {
				old := atomic.LoadInt64(&maxRunning)
				if cur <= old || atomic.CompareAndSwapInt64(&maxRunning, old, cur) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&running, -1)
		}()
	}

	wg.Wait()

	if maxRunning > 2 {
		t.Errorf("expected max concurrent <= 2, got %d", maxRunning)
	}
}

func TestHighLimitDoesNotBlock(t *testing.T) {
	wg := waitgroup.NewLimitWaitGroup(100)
	started := int64(0)
	done := make(chan struct{})

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&started, 1)
			<-done
		}()
	}

	time.Sleep(50 * time.Millisecond)

	if started != 50 {
		t.Errorf("expected all 50 to start, got %d", started)
	}

	close(done)
	wg.Wait()
}
