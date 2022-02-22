package starter

import (
	"github.com/anatollupacescu/curly-spin/signaler"
)

type controller struct {
	self         signaler.S
	dependencies signaler.Group
	dependants   signaler.Group
}

func (c *controller) DependenciesReadyToStart() chan struct{} {
	return c.dependencies.All(signaler.Ready)
}

func (c *controller) DependenciesStarted() chan struct{} {
	return c.dependencies.All(signaler.Started)
}

func (c *controller) DependencyFailedToStart() chan struct{} {
	return c.dependencies.Any(signaler.Fault)
}

func (c *controller) DependantsAreDone() chan struct{} {
	return c.dependants.All(signaler.Done)
}

func (c *controller) SignalReadyToStart() {
	c.self.Signal(signaler.Ready)
}

func (c *controller) SignalFailedToStart() {
	c.self.Signal(signaler.Fault)
}

func (c *controller) SignalStarted() {
	c.self.Signal(signaler.Started)
}

func (c *controller) SignalDone() {
	c.self.Signal(signaler.Done)
}
