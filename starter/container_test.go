package starter_test

import (
	"sync"
	"testing"
	"time"

	lib "github.com/anatollupacescu/wire-starter/starter"
	"github.com/stretchr/testify/assert"
)

var defaultMock = serviceMock{
	startFn: func() {
	},
}

func TestContainerLen(t *testing.T) {
	var c lib.Container
	s1 := defaultMock
	c.Add(&s1)
	assert.Equal(t, 1, c.Len())

	s2 := defaultMock
	c.Add(&s2)
	assert.Equal(t, 2, c.Len())

	c.WaitFor(&s1, &s2)
	assert.Equal(t, 2, c.Len())

	s3 := defaultMock
	c.WaitFor(&s2, &s3)
	assert.Equal(t, 3, c.Len())
}

func TestContainer(t *testing.T) {
	// t.Run("can not add cyclic dependencies", func(t *testing.T) {
	// 	c := make(lib.Container)
	// 	s1 := defaultMock
	// 	s2 := defaultMock
	// 	c.Add(&s1, &s2)
	// 	assert.NoError(t, err)

	// 	s3 := defaultMock
	// 	c.Add(&s2, &s3)
	// 	assert.NoError(t, err)

	// 	c.Add(&s3, &s1)
	// 	assert.Equal(t, err, ErrCyclicDependency)
	// })
}

// // Empty container
// func TestRunEmptyContainer(t *testing.T) {
// 	c := make(lib.Container)

// 	ctrl := c.Start()

// 	select {
// 	case <-ctrl:
// 	case <-time.After(1 * time.Millisecond):
// 		t.Error("timeout")
// 	}

// 	err := c.Shutdown()
// 	if err != nil {
// 		t.Error("shutdown", err)
// 	}
// }

// // Tests that a single service container acts exactly as the wrapped service
// func TestRunSingleService(t *testing.T) {
// 	c := make(lib.Container)

// 	s1 := defaultMock

// 	c.Add(&s1)

// 	ctrl := c.Start()

// 	select {
// 	case <-ctrl:
// 	case <-time.After(1 * time.Millisecond):
// 		t.Error("timeout")
// 	}

// 	err := c.Shutdown()
// 	if err != nil {
// 		t.Error("shutdown", err)
// 	}
// }

// func TestRunSingleServiceTimeouts(t *testing.T) {
// 	c := make(lib.Container)

// 	s1 := &serviceMock{
// 		startFn: func() {
// 			time.Sleep(2 * time.Millisecond)
// 		},
// 	}

// 	c.Add(s1)

// 	ctrl := c.Start()

// 	select {
// 	case <-ctrl:
// 		t.Error("expected timeout")
// 	case <-time.After(1 * time.Millisecond):
// 	}

// 	err := c.Shutdown()
// 	if err != nil {
// 		t.Error("shutdown", err)
// 	}
// }

// func TestRunContainerShutdown(t *testing.T) {
// 	c := make(lib.Container)

// 	s1 := &serviceMock{
// 		startFn: func() {
// 			time.Sleep(2 * time.Millisecond)
// 		},
// 	}

// 	c.Add(s1)

// 	ctrl := c.Start()
// 	<-ctrl

// 	err := c.Shutdown()
// 	if err != nil {
// 		t.Error("shutdown", err)
// 	}
// }

func TestMultiService(t *testing.T) {
	order := make(chan uint8, 3)

	load := &serviceMock{
		startFn: func() {
			order <- 1
		},
	}
	transform := &serviceMock{
		startFn: func() {
			order <- 2
			time.Sleep(99 * time.Millisecond)
		},
	}
	extract := &serviceMock{
		startFn: func() {
			order <- 3
		},
	}

	var c lib.Container

	c.Add(extract)
	c.Add(load)
	c.Add(transform)

	c.WaitFor(load, transform)
	c.WaitFor(transform, extract)

	start := new(sync.WaitGroup)
	start.Add(1)

	go func() {
		start.Done()
		c.Start()
	}()

	start.Wait()

	done := make(chan struct{})
	timeout := make(chan struct{})

	assertion := new(sync.WaitGroup)
	assertion.Add(1)

	go func() {
		assertion.Done()
		select {
		case <-done:
			close(done)
		case <-time.After(100 * time.Millisecond):
			timeout <- struct{}{}
		}
	}()

	assertion.Wait() // waiting for the assertion routine to be scheduled

	var count int

	for {
		select {
		case <-timeout:
			t.Error("timeout")
			close(timeout)
			return
		case o := <-order:
			if o != uint8(3-count) {
				t.Error("wrong order: want", 3-count, "got", o)
				done <- struct{}{}
				return
			}
			count++
			if count == 3 {
				done <- struct{}{}
				close(order)
				return
			}
		}
	}
}

type serviceMock struct {
	running bool
	startFn func()
}

func (s *serviceMock) Running() bool {
	return s.running
}

func (s *serviceMock) Stop() {

}

func (s *serviceMock) Start() {
	s.startFn()
	s.running = true
}
