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
	status  int32
}

const (
	Pending = iota
	Started
	Failed
	Cancelled
)

func (c *C) Start(ctx context.Context) {
	defer close(c.done)
	ct := time.After(c.timeout)
	for _, r := range c.waitOn {
		select {
		case <-r.done:
			if atomic.LoadInt32(&r.status) == Started {
				continue
			}
		case <-ctx.Done():
		case <-ct:
		}
		c.setStatus(Cancelled)
		return
	}
	if err := c.fn(ctx); err != nil {
		c.setStatus(Failed)
		return
	}

	c.setStatus(Started)
}

func (c *C) setStatus(in int32) {
	if !atomic.CompareAndSwapInt32(&c.status, Pending, in) {
		panic("unexpected state")
	}
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
	return atomic.LoadInt32(&c.status)
}

func (c *C) Done() chan struct{} {
	return c.done
}
