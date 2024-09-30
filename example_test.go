package starter

import (
	"context"
	"fmt"
	"sync"
)

func ExampleStarter() {
	var (
		calls []string
		mu    sync.Mutex
	)

	acc := func(in string) {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, in)
	}

	db := &capture{name: "DB", fn: acc}
	ft := &capture{name: "FT", fn: acc}
	web := &capture{name: "WEB", fn: acc}

	sdb := New(db.Start)
	sft := New(ft.Start)
	swb := New(web.Start)

	// WEB should wait for its dependencies
	swb.WaitOn(sdb, sft)

	for _, i := range []*C{swb, sdb, sft, swb} {
		go i.Start(context.Background())
	}

	<-swb.Done()

	if len(calls) != 3 {
		panic("expected calls: 3")
	}

	fmt.Println(calls[2])
	// Output: WEB
}

type capture struct {
	name string
	fn   func(string)
}

func (c *capture) Start(context.Context) error {
	c.fn(c.name)
	return nil
}
