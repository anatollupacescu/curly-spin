package starter

type hook struct {
	c chan struct{}
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

func (h *hook) wait() <-chan struct{} {
	return h.c
}
