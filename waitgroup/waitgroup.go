package waitgroup

import (
	"errors"
	"sync"
)

var (
	ErrInvalidLimit = errors.New("limit must be at least 1")
)

type WaitGroup interface {
	Add(delta int)
	Done()
	Wait()
}

type LimitWaitGroup struct {
	wg    sync.WaitGroup
	limit chan struct{}
}

type option func(*LimitWaitGroup) error

func WithLimit(limit int) option {
	return func(wg *LimitWaitGroup) error {
		if limit < 1 {
			return ErrInvalidLimit
		}
		wg.limit = make(chan struct{}, limit)
		return nil
	}
}
func WithWaitGroup(wg sync.WaitGroup) option {
	return func(lwg *LimitWaitGroup) error {
		lwg.wg = wg
		return nil
	}
}

// NewWaitGroup creates a WaitGroup with an optional concurrency limit.
// If limit <= 0, it returns a standard sync.WaitGroup.
func NewLimitWaitGroup(opts ...option) (WaitGroup, error) {
	if len(opts) == 0 {
		return &sync.WaitGroup{}, nil
	}

	wg := &LimitWaitGroup{}
	for _, opt := range opts {
		if err := opt(wg); err != nil {
			return nil, err
		}
	}
	return wg, nil
}

func (w *LimitWaitGroup) Add(delta int) {
	for i := 0; i < delta; i++ {
		w.limit <- struct{}{}
	}
	w.wg.Add(delta)
}

func (w *LimitWaitGroup) Done() {
	w.wg.Done()
	<-w.limit
}

func (w *LimitWaitGroup) Wait() {
	w.wg.Wait()
}
