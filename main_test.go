package starter_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/anatollupacescu/starter"
)

func TestDependencyNeverStarts(t *testing.T) {
	ws := new(mock)
	web := starter.New(ws.Start)

	dbs := new(mock)
	db := starter.New(dbs.Start)

	web.WaitOn(db)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	web.Start(ctx)

	assert(t, db.Status(), starter.Pending)
	assert(t, web.Status(), starter.Cancelled)
}

func TestDependantWaitsForDependency(t *testing.T) {
	webs := &mock{name: "web service"}
	web := starter.New(webs.Start)

	dbs := &mock{name: "DB service"}
	db := starter.New(dbs.Start)

	web.WaitOn(db)

	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(cancel)

	go func() {
		time.Sleep(time.Millisecond * 100)
		db.Start(ctx)
	}()

	web.Start(ctx)

	assert(t, db.Status(), starter.Started)
	assert(t, web.Status(), starter.Started)
}

func TestDependencyErrors(t *testing.T) {
	webs := &mock{name: "web service"}
	web := starter.New(webs.Start)

	dbs := &mock{name: "DB service", err: errors.New("db conn")}
	db := starter.New(dbs.Start)

	web.WaitOn(db)

	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(cancel)

	db.Start(ctx)
	web.Start(ctx)

	assert(t, db.Status(), starter.Failed)
	assert(t, web.Status(), starter.Cancelled)
}

func TestDoesNotWaitOverAllowed(t *testing.T) {
	webs := &mock{name: "web service"}
	web := starter.New(webs.Start)

	dbs := &mock{name: "DB service"}
	db := starter.New(dbs.Start)

	web.WaitForDur(500 * time.Millisecond)
	web.WaitOn(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	t.Cleanup(cancel)

	web.Start(ctx)

	assert(t, db.Status(), starter.Pending)
	assert(t, web.Status(), starter.Cancelled)
}

func assert(t *testing.T, got, want int32) {
	t.Helper()
	if got != want {
		t.Fatalf("want %v, got %v", want, got)
	}
}

type mock struct {
	name string
	err  error
}

func (s *mock) Start(context.Context) error {
	return s.err
}
