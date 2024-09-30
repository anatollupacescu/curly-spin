package starter

import (
	"context"
	"math"
	"sync/atomic"
	"time"
)

type C struct {
	waitOn  []*C
	fn      func(context.Context) error
	done    chan struct{}
	timeout time.Duration
	status  atomic.Int32
}

const (
	Pending = iota
	starting
	Started
	Failed
	Cancelled
)

func (c *C) Start(ctx context.Context) {
	if !c.status.CompareAndSwap(Pending, starting) {
		return
	}

	defer close(c.done)

	ct := time.After(c.timeout)

	for _, dep := range c.waitOn {
		select {
		case <-dep.done:
			if dep.status.Load() == Started {
				continue
			}
		case <-ctx.Done():
		case <-ct:
		}
		c.status.Store(Cancelled)
		return
	}

	if err := c.fn(ctx); err != nil {
		c.status.Store(Failed)
		return
	}

	c.status.Store(Started)
}

func New(fn func(context.Context) error) *C {
	return &C{
		fn:      fn,
		done:    make(chan struct{}),
		timeout: math.MaxInt64,
	}
}

func (c *C) WaitOn(in ...*C) {
	c.waitOn = append(c.waitOn, in...)
}

func (c *C) WaitForDur(in time.Duration) {
	c.timeout = in
}

func (c *C) Status() int32 {
	return c.status.Load()
}

func (c *C) Done() <-chan struct{} {
	return c.done
}
