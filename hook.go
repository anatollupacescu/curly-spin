package starter

import (
	"sync"
)

type hook struct {
	mu       sync.Mutex
	isClosed bool
	c        chan struct{}
}

func newHook(uu ...*unit) *hook {
	h := new(hook)

	h.c = make(chan struct{})

	a := make(chan struct{}, len(uu))

	for _, u := range uu {
		ws := make(chan struct{})
		go func() {
			close(ws)
			<-u.done
			a <- struct{}{}
		}()
		<-ws
	}

	go func() {
		for range len(uu) {
			<-a
		}

		close(a)
		close(h.c)
	}()

	return h
}

func (h *hook) wait() {
	h.mu.Lock()
	if h.isClosed {
		return
	}
	h.mu.Unlock()

	<-h.c
}
