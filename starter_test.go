package starter_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	starter "github.com/anatollupacescu/curly-spin"
	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	t.Run("given duplicate service", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		a := defaultAgent("a1")
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		_ = starter.Nodes([]*starter.Node{a, a}).Start(ctx)

	})
	t.Run("given cyclic dependency", func(t *testing.T) {
		t.Run("waits on itself", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()

			a := defaultAgent("a1")
			a.DependsOn(a)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			_ = starter.Nodes([]*starter.Node{a}).Start(ctx)
		})
		t.Run("waits on parent", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()

			a := defaultAgent("a")
			b := defaultAgent("b")
			a.DependsOn(b)
			b.DependsOn(a)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			_ = starter.Nodes([]*starter.Node{a, b}).Start(ctx)
		})
		t.Run("waits on grandparent", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()

			a := defaultAgent("a")
			b := defaultAgent("b")
			c := defaultAgent("c")
			a.DependsOn(b)
			b.DependsOn(c)
			c.DependsOn(a)
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			_ = starter.Nodes([]*starter.Node{a, b, c}).Start(ctx)
		})
	})
	t.Run("given service started", func(t *testing.T) {
		t.Run("no dependencies", func(t *testing.T) {
			a := defaultAgent("a1")
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)
			errC := starter.Nodes([]*starter.Node{a}).Start(ctx)
			t.Run("assert started", func(t *testing.T) {
				select {
				case err := <-errC:
					t.Fatal("start with no deps", err)
				default:
				}
			})
		})
		t.Run("one dependency", func(t *testing.T) {
			a := defaultAgent("a1")
			a2 := defaultAgent("a2")

			a.DependsOn(a2) // a2 starts first

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			errC := starter.Nodes([]*starter.Node{a, a2}).Start(ctx)

			t.Run("assert all started", func(t *testing.T) {
				select {
				case err := <-errC:
					t.Fatal("start with one dep", err)
				default:
				}
			})
		})
		t.Run("two dependencies", func(t *testing.T) {
			a := defaultAgent("a1")
			a2 := defaultAgent("a2")

			a.DependsOn(a2) // a2 starts first

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			errC := starter.Nodes([]*starter.Node{a, a2}).Start(ctx)

			t.Run("assert all started", func(t *testing.T) {
				select {
				case err := <-errC:
					t.Fatal("start with one dep", err)
				default:
				}
			})
		})
	})
	t.Run("given service not started", func(t *testing.T) {
		t.Run("one dependency, times out", func(t *testing.T) {
			name := "a1"
			ag1 := &agent{name: name}
			a1 := starter.New(name, ag1)

			name = "a2"
			a2 := starter.New(name, &agent{
				name:       name,
				startSleep: 110 * time.Millisecond,
			}, 100*time.Millisecond)

			a1.DependsOn(a2) // a2 starts first

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			err := <-starter.Nodes([]*starter.Node{a1, a2}).Start(ctx)

			t.Run("assert no started", func(t *testing.T) {
				if ag1.state != 0 {
					t.Fatal("expected not started")
				}
			})
			t.Run("assert error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
		t.Run("two dependencies, one times out", func(t *testing.T) {
			name := "a1"
			ag1 := &agent{name: name}
			a1 := starter.New(name, ag1)

			name = "a2"
			a2 := starter.New(name, &agent{
				name:       name,
				startSleep: 110 * time.Millisecond,
			}, 100*time.Millisecond)

			name = "a3"
			ag3 := &agent{name: name}
			a3 := starter.New(name, ag3)

			a1.DependsOn(a2)
			a1.DependsOn(a3)

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			err := <-starter.Nodes([]*starter.Node{a1, a2, a3}).Start(ctx)

			t.Run("assert one started", func(t *testing.T) {
				if ag1.state != 0 {
					t.Fatal("expected not started")
				}
				if ag3.state != 2 {
					t.Fatal("expected dep started")
				}
			})
			t.Run("assert error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
		t.Run("two dependencies, one errors", func(t *testing.T) {
			name := "a1"
			ag1 := &agent{name: name}
			a1 := starter.New(name, ag1)

			name = "a2"
			ag2 := &agent{
				name:  name,
				start: errors.New("test"),
			}
			a2 := starter.New(name, ag2)

			name = "a3"
			ag3 := &agent{name: name}
			a3 := starter.New(name, ag3)

			a1.DependsOn(a2)
			a1.DependsOn(a3)

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			err := <-starter.Nodes([]*starter.Node{a1, a2, a3}).Start(ctx)

			t.Run("assert one started", func(t *testing.T) {
				if ag1.state != 0 {
					t.Fatal("expected not started")
				}
				if ag2.state != 2 {
					t.Fatal("expected dep started")
				}
			})
			t.Run("assert error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
	})
}

func TestStop(t *testing.T) {
	t.Run("given service stopped", func(t *testing.T) {
		t.Run("one dependency, stops in time", func(t *testing.T) {
			name := "a1"
			ag1 := &agent{name: name}
			a1 := starter.New(name, ag1)

			name = "a2"
			ag2 := &agent{name: name}
			a2 := starter.New(name, ag2)

			a1.DependsOn(a2)

			ctx, cancel := context.WithCancel(context.Background())

			var wg1, wg2 sync.WaitGroup
			wg1.Add(1)
			wg2.Add(1)
			go func() {
				wg1.Done()
				<-starter.Nodes([]*starter.Node{a1, a2}).Start(ctx)
				wg2.Done()
			}()

			wg1.Wait()
			time.Sleep(100 * time.Millisecond)
			cancel()
			wg2.Wait()

			t.Run("assert success", func(t *testing.T) {
				assert.Equal(t, 3, ag1.state)
				assert.Equal(t, 3, ag2.state)
			})
		})
		t.Run("one dependency, takes time to finish", func(t *testing.T) {
			name := "a1"
			ag1 := &agent{name: name}
			a1 := starter.New(name, ag1)

			name = "a2"
			ag2 := &agent{name: name, stopSleep: time.Second}
			a2 := starter.New(name, ag2)

			a1.DependsOn(a2)

			ctx, cancel := context.WithCancel(context.Background())

			var start time.Time

			var wg1, wg2 sync.WaitGroup
			wg1.Add(1)
			wg2.Add(1)
			go func() {
				wg1.Done()
				start = time.Now()
				<-starter.Nodes([]*starter.Node{a1, a2}).Start(ctx)
				wg2.Done()
			}()

			wg1.Wait()
			time.Sleep(100 * time.Millisecond)
			cancel()
			wg2.Wait()

			t.Run("assert waits", func(t *testing.T) {
				assert.Equal(t, 3, ag1.state)
				assert.Equal(t, 3, ag2.state)
				assert.True(t, time.Since(start) > 1100*time.Millisecond)
			})
		})
	})
	t.Run("given service not stopped", func(t *testing.T) {
		t.Run("assert error", func(t *testing.T) {
		})
	})
}

func defaultAgent(name string) *starter.Node {
	return starter.New(name, &agent{
		name: name,
	})
}

type agent struct {
	name        string
	startSleep  time.Duration
	stopSleep   time.Duration
	start, stop error

	state int
}

func (a *agent) Start(ctx context.Context) error {
	a.state = 1 //start called
	time.Sleep(a.startSleep)
	a.state = 2
	return a.start
}

func (a *agent) Stop() error {
	a.state = 3 //stop called
	time.Sleep(a.stopSleep)
	return a.stop
}
