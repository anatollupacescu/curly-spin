package main_test

import (
	"context"
	"fmt"
	"time"

	starter "github.com/anatollupacescu/curly-spin"
)

func dbFn(ctx context.Context) <-chan error {
	out := make(chan error)
	go func() {
		fmt.Println("start db")
		time.Sleep(time.Second)
		close(out)
	}()
	return out
}

func httpFn(ctx context.Context) <-chan error {
	out := make(chan error)
	go func() {
		fmt.Println("start http")
		close(out)
	}()
	return out
}

func ExampleContainer() {
	db := starter.NewFn("database", dbFn)
	http := starter.NewFn("http", httpFn)

	r := starter.NewRunner()

	r.Register(db)
	r.Register(http)

	r.Order(db, http)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.Start(ctx)

	// Output:
	// start db
	// start http
}

// stop http
// stop db
