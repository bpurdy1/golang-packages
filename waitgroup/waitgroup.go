package waitgroup

import (
	"errors"
	"sync"
)

var ErrDeltaExceedingLimit = errors.New("waitgroup: Add called with delta exceeding limit")

type WaitGroup interface {
	Add(delta int)
	Done()
	Wait()
	Limit() int
	WithWaitGroup(wg *sync.WaitGroup) WaitGroup
}

type LimitWaitGroup struct {
	wg    sync.WaitGroup
	limit chan struct{}
}

// NewLimitWaitGroup creates a new LimitWaitGroup with no limit.
func NewLimitWaitGroup(limit int) WaitGroup {
	lwg := &LimitWaitGroup{
		wg:    sync.WaitGroup{},
		limit: make(chan struct{}, limit),
	}
	return lwg
}

func (w *LimitWaitGroup) Limit() int {
	return cap(w.limit)
}

func (w *LimitWaitGroup) WithWaitGroup(wg *sync.WaitGroup) WaitGroup {
	w.wg = *wg //nolint:govet // intentional copy to replace the internal waitgroup
	return w
}

func (w *LimitWaitGroup) Add(delta int) {
	if w.limit != nil {
		if delta > cap(w.limit) {
			panic(ErrDeltaExceedingLimit)
		}
		for range delta {
			w.limit <- struct{}{}
		}
	}
	w.wg.Add(delta)
}

func (w *LimitWaitGroup) Done() {
	w.wg.Done()
	if w.limit != nil {
		<-w.limit
	}
}

func (w *LimitWaitGroup) Wait() {
	w.wg.Wait()
}
