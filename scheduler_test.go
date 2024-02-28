package starter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerTwoDepsHook(t *testing.T) {
	s := NewScheduler()

	var c = make(chan struct{})
	r := &rmock{c: make(chan error), name: "db", r: func(ci chan error) {
		<-c
		close(ci)
	}}

	var c2 = make(chan struct{})
	m := &rmock{c: make(chan error), name: "main", r: func(ci chan error) {
		<-c2
		close(ci)
	}}

	u := &unit{target: r}
	v := &unit{target: m, waitFor: []*unit{u}}

	s.schedule(context.Background(), u)
	s.schedule(context.Background(), v)

	s.start()

	bl := make(chan struct{})

	ss := make(chan struct{})

	go func() {
		close(ss)
		<-newHook(v).wait()
		close(bl)
	}()

	<-ss

	select {
	case <-bl:
		t.Fatal("should block here")
	default:
	}

	close(c)

	select {
	case <-bl:
		t.Fatal("should block here")
	default:
	}

	close(c2)

	<-bl
}

func TestSchedulerTwoDeps(t *testing.T) {
	s := NewScheduler()

	var ev []string
	var c = make(chan struct{})

	r := &rmock{c: make(chan error), name: "db", r: func(ci chan error) {
		ev = append(ev, "db started")
		close(c)
		close(ci)
	}}

	var c2 = make(chan struct{})
	m := &rmock{c: make(chan error), name: "main", r: func(ci chan error) {
		ev = append(ev, "m started")
		close(c2)
		close(ci)
	}}

	u := &unit{target: r}
	v := &unit{target: m, waitFor: []*unit{u}}

	s.schedule(context.Background(), u)
	s.schedule(context.Background(), v)

	assert.Empty(t, ev)
	s.start()

	<-c
	assert.Contains(t, ev, "db started")
	assert.Len(t, ev, 1)

	<-c2
	assert.Contains(t, ev, "m started")
	assert.Len(t, ev, 2)
}

type rmock struct {
	name string
	c    chan error
	r    func(c chan error)
}

func (r *rmock) Run(context.Context) <-chan error {
	r.r(r.c)
	return r.c
}

func (r *rmock) String() string {
	return r.name
}
