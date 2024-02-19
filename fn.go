package starter

import "context"

type Fn struct {
	name string
	run  func(ctx context.Context) <-chan error
}

func (f *Fn) Run(ctx context.Context) <-chan error {
	return f.run(ctx)
}

func (f *Fn) String() string {
	return f.name
}

func NewFn(name string, run func(ctx context.Context) <-chan error) *Fn {
	return &Fn{
		name: name,
		run:  run,
	}
}
