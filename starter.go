package starter

import (
	"context"
	"fmt"
	"math"
	"time"

	"golang.org/x/exp/slices"

	"github.com/anatollupacescu/curly-spin/signaler"
)

type Worker interface {
	Start(context.Context) error
	Stop() error
}

type Node struct {
	Worker

	name string
	wait time.Duration // how much time it takes for this worker to start

	nodes []*Node
}

func (n *Node) DependsOn(d *Node) {
	n.nodes = append(n.nodes, d)
}

func New(name string, worker Worker, wait ...time.Duration) *Node {
	n := &Node{
		name:   name,
		Worker: worker,
	}

	if len(wait) > 0 {
		n.wait = wait[0]
	}

	return n
}

type wrap struct {
	start func(context.Context) error
	stop  func() error
}

func (w *wrap) Stop() error {
	return w.stop()
}
func (w *wrap) Start(ctx context.Context) error {
	return w.start(ctx)
}

func NewFn(name string, start func(context.Context) error, stopFns ...func() error) *Node {
	stop := func() error { return nil }
	if len(stopFns) > 0 {
		stop = stopFns[0]
	}
	return &Node{
		name: name,
		Worker: &wrap{
			start: start,
			stop:  stop,
		},
	}
}

func dependencies(ss map[string]signaler.S, n *Node) (group signaler.Group) {
	for _, dep := range n.nodes {
		group = append(group, ss[dep.name])
	}

	return
}

func dependants(ss map[string]signaler.S, me *Node, all []*Node) (group signaler.Group) {
	for _, n := range all {
		if slices.Contains(n.nodes, me) {
			group = append(group, ss[n.name])
		}
	}

	return
}

type Nodes []*Node

func checkPreconditions(nn Nodes) {
	if len(nn) == 0 {
		panic("no services to run")
	}
	nn.forEach(func(n *Node) {
		Nodes(n.nodes).forEach(func(d *Node) {
			d.mustNotDependOn(n)
		})
	})
}

func (n *Node) mustNotDependOn(d *Node) {
	Nodes(n.nodes).forEach(func(n *Node) {
		if n == d {
			panic("circular dependency")
		}
		n.mustNotDependOn(d)
	})
}

func (nn Nodes) forEach(f func(*Node)) {
	for _, n := range nn {
		f(n)
	}
}

func Run(ctx context.Context, n *Node, nn ...*Node) chan error {
	all := append(nn, n)
	return Nodes(all).Start(ctx)
}

func (nn Nodes) Start(ctx context.Context) chan error {
	checkPreconditions(nn)

	ss := make(map[string]signaler.S)
	for _, n := range nn {
		ss[n.name] = signaler.New(n.name)
	}

	if len(ss) != len(nn) {
		panic("duplicate service")
	}

	errC := make(chan error)

	for _, n := range nn {
		self := ss[n.name]
		ctrl := &controller{
			self:         self,
			dependencies: dependencies(ss, n),
			dependants:   dependants(ss, n, nn),
		}

		n := n
		go func() {
			defer ctrl.SignalDone()
			if ctx.Err() != nil {
				errC <- ctx.Err()
				return
			}
			if err := n.Run(ctx, ctrl); err != nil {
				errC <- err
			}
		}()
	}

	var all signaler.Group

	for _, v := range ss {
		all = append(all, v)
	}

	go func() {
		<-all.All(signaler.Done)
		close(errC)
	}()

	return errC
}

func (n *Node) Run(ctx context.Context, controller *controller) (err error) {
	// block until dependencies signal "i'm ready to start" so we would not compound start times
	select {
	case <-controller.DependenciesReadyToStart():
	case <-ctx.Done():
		return fmt.Errorf("%s: wait for dependencies to report ready: %w", n.name, ctx.Err())
	}

	maxWait := Nodes(n.nodes).startWaitLimit()

	// wait (with timeout) for dependencies to signal "i'm running"
	select {
	case <-controller.DependenciesStarted():
		controller.SignalReadyToStart()
	case <-controller.DependencyFailedToStart():
		err = fmt.Errorf("[%s] some dependency failed to start", n.name)
	case <-time.After(maxWait):
		err = fmt.Errorf("[%s] wait for dependencies to start: timeout", n.name)
	case <-ctx.Done():
		err = fmt.Errorf("[%s] wait for dependencies to start: %w", n.name, ctx.Err())
	}

	if err != nil {
		controller.SignalFailedToStart()
		return
	}

	// open for service
	if err = n.Start(ctx); err != nil {
		controller.SignalFailedToStart()
		return
	}

	controller.SignalStarted()

	<-ctx.Done() // do work until shutdown signal received
	<-controller.DependantsAreDone()

	// shut down
	return n.Stop()
}

func (c Nodes) startWaitLimit() (maxWait time.Duration) {
	var found bool
	for _, d := range c {
		if d.wait > maxWait {
			maxWait = d.wait
			found = true
		}
	}
	if !found {
		return math.MaxInt64
	}
	return
}
