package signaler

import "sync"

type Signal int

const (
	Ready = iota
	Started
	Done
	Fault
)

type signalChan chan struct{}

type S struct {
	name   string
	steady signalChan
	start  signalChan
	stop   signalChan
	fault  signalChan
}

func (s S) Signal(signal Signal) {
	switch signal {
	case Ready:
		close(s.steady)
	case Started:
		close(s.start)
	case Done:
		close(s.stop)
	case Fault:
		close(s.fault)
	default:
		panic(signal)
	}
}

type Group []S

func (c Group) All(signal Signal) signalChan {
	out := make(chan struct{})

	if len(c) == 0 {
		close(out)
		return out
	}

	var wg sync.WaitGroup

	for _, n := range c {
		n := n
		wg.Add(1)
		go func() {
			<-mapSignal(signal, n)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (c Group) Any(signal Signal) signalChan {
	out := make(chan struct{})

	var once sync.Once

	for _, n := range c {
		n := n
		go func() {
			select {
			case <-out:
			case <-mapSignal(signal, n):
			}
			once.Do(func() {
				close(out)
			})
		}()
	}

	return out
}

func mapSignal(signal Signal, s S) signalChan {
	switch signal {
	case Ready:
		return s.steady
	case Started:
		return s.start
	case Done:
		return s.stop
	case Fault:
		return s.fault
	default:
		panic(signal)
	}
}

func New(name string) S {
	return S{
		name:   name,
		steady: make(signalChan),
		start:  make(signalChan),
		stop:   make(signalChan),
		fault:  make(signalChan),
	}
}
