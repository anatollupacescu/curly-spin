package starter

import (
	"context"
)

type runnable interface {
	Run(context.Context) <-chan error
	String() string
}

type unit struct {
	target  runnable
	waitFor []*unit
	done    chan struct{}
}

type scheduler interface {
	schedule(context.Context, ...*unit)
	start()
}

type runner struct {
	svc       map[string]*unit
	scheduler scheduler
}

func NewRunner() *runner {
	return &runner{
		svc:       make(map[string]*unit),
		scheduler: NewScheduler(),
	}
}

func (r *runner) Register(target runnable) {
	if _, ok := r.svc[target.String()]; ok {
		panic("duplicate registration")
	}

	r.svc[target.String()] = &unit{
		target: target,
	}
}

func (r *runner) Order(first, second runnable) {
	var p1, p2 *unit
	var ok bool
	if p1, ok = r.svc[first.String()]; !ok {
		panic("first runnable not registered")
	}
	if p2, ok = r.svc[second.String()]; !ok {
		panic("second runnable not registered")
	}
	// TODO panic on circular dep
	p2.waitFor = append(p2.waitFor, p1)
	r.svc[second.String()] = p2
}

func (r *runner) Start(ctx context.Context) {
	batch := free(r.svc)

	r.scheduler.schedule(ctx, batch...)

	var pb []*unit // previous batch

	for {
		batch = dependants(r.svc, batch)

		if len(batch) == 0 {
			break
		}

		pb = batch

		r.scheduler.schedule(ctx, batch...)
	}

	go func() {
		r.scheduler.start()
	}()

	newHook(pb...).wait()
}

func free(units map[string]*unit) (out []*unit) {
	for _, unit := range units {
		if len(unit.waitFor) == 0 {
			out = append(out, unit)
		}
	}
	return
}

func dependants(units map[string]*unit, scheduled []*unit) (deps []*unit) {
	out := make(map[string]*unit)

	for _, u := range units {
		if len(u.waitFor) == 0 {
			continue
		}
		if each(u.waitFor).in(scheduled) {
			out[u.target.String()] = u
		}
	}

	for _, d := range out {
		deps = append(deps, d)
	}

	return
}

type each []*unit

func (s each) in(set []*unit) bool {
outer:
	for _, v := range s {
		for _, s := range set {
			if v.target.String() == s.target.String() {
				continue outer
			}
		}
		return false
	}

	return true
}
