package main_test

import (
	"context"
	"fmt"
	"time"

	starter "github.com/anatollupacescu/curly-spin"
)

func dbFn(ctx context.Context) error {
	fmt.Println("start db")
	return nil
}

func httpFn(ctx context.Context) error {
	fmt.Println("start http")
	return nil
}

func ExampleContainer() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db := starter.NewFn("database", dbFn, func() error { fmt.Println("stop db"); return nil })
	http := starter.NewFn("http", httpFn, func() error { fmt.Println("stop http"); return nil })

	http.DependsOn(db)

	<-starter.Run(ctx, db, http)

	// Output:
	// start db
	// start http
	// stop http
	// stop db
}
