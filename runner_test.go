package starter

import (
	"context"
	"testing"

	"sync/atomic"

	"github.com/stretchr/testify/assert"
)

func TestRunner(t *testing.T) {
	r := &runner{
		svc:       make(map[string]*unit),
		scheduler: NewScheduler(),
	}

	var order int32

	db := &rmock{name: "db", c: make(chan error), r: func(c chan error) {
		atomic.AddInt32(&order, 1)
		close(c)
	}}
	tport := &rmock{name: "tport", c: make(chan error), r: func(c chan error) {
		atomic.AddInt32(&order, 2)
		close(c)
	}}
	main := &rmock{name: "main", c: make(chan error), r: func(c chan error) {
		atomic.CompareAndSwapInt32(&order, 3, 100)
		close(c)
	}}

	r.Register(db)
	r.Register(main)
	r.Register(tport)

	r.Order(db, main)
	r.Order(tport, main)

	r.Start(context.Background())

	assert.Equal(t, order, int32(100))
}
