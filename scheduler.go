package starter

import (
	"context"
)

type Scheduler struct {
	svc map[string]*unit
	s   chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		svc: make(map[string]*unit),
		s:   make(chan struct{}),
	}
}

func (s *Scheduler) start() {
	close(s.s)
}

func (s *Scheduler) schedule(ctx context.Context, uu ...*unit) {
	for _, u := range uu {
		// skip scheduled
		if _, ok := s.svc[u.target.String()]; ok {
			continue
		}

		u.done = make(chan struct{})
		s.svc[u.target.String()] = u

		// scheduled channel
		sc := make(chan struct{})
		defer func() { <-sc }()

		u := u
		go func() {
			defer close(u.done)

			close(sc)

			<-s.s //start

			if len(u.waitFor) > 0 {
				newHook(u.waitFor...).wait()
			}

			<-u.target.Run(ctx)
		}()
	}
}
