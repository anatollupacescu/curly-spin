package signaler_test

import (
	"testing"
	"time"

	"github.com/anatollupacescu/curly-spin/signaler"
)

func TestEmpty(t *testing.T) {
	t.Parallel()

	c := signaler.Group(nil)

	select {
	case <-c.All(signaler.Done):
	default:
		t.Fatal()
	}
}

func TestSignal(t *testing.T) {
	t.Parallel()

	s := signaler.New("svc")
	c := signaler.Group([]signaler.S{s})

	s.Signal(signaler.Ready)

	t.Run("wrong signal", func(t *testing.T) {
		select {
		case <-c.All(signaler.Started):
			t.Fatal()
		case <-c.Any(signaler.Started):
			t.Fatal()
		case <-time.After(10 * time.Millisecond):
		}
	})

	t.Run("right signal", func(t *testing.T) {
		select {
		case <-c.All(signaler.Ready):
		case <-time.After(10 * time.Millisecond):
			t.Fatal()
		}
	})
}

func TestSignalMultiService(t *testing.T) {
	t.Parallel()

	s1 := signaler.New("svc1")
	s2 := signaler.New("svc2")

	c := signaler.Group([]signaler.S{s1, s2})

	go func() {
		time.Sleep(2 * time.Second)
		s2.Signal(signaler.Started)
	}()

	st := time.Now()

	t.Run("all", func(t *testing.T) {
		s1.Signal(signaler.Started)

		select {
		case <-c.All(signaler.Started):
			if time.Since(st) < 2*time.Second {
				t.Fatal("expected at least a 2 second wait")
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timeout")
		}
	})

	t.Run("any", func(t *testing.T) {
		s2.Signal(signaler.Done)

		select {
		case <-c.Any(signaler.Done):
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	})
}

func TestFault(t *testing.T) {
	t.Parallel()

	s1 := signaler.New("svc1")
	s2 := signaler.New("svc2")

	c := signaler.Group([]signaler.S{s1, s2})

	s1.Signal(signaler.Started)

	go func() {
		time.Sleep(time.Second)
		s2.Signal(signaler.Fault)
	}()

	select {
	case <-c.Any(signaler.Fault):
	case <-c.All(signaler.Started):
		t.Fatal("expected fault")
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}
