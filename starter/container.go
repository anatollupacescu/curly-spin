package starter

type service interface {
	Start()
	Stop()
	Running() bool
}

type Container node

func (c *Container) Add(s service) {
	if s == nil {
		panic("nil service")
	}

	_, found := node(*c).find(s)

	// duplicates not allowed
	if found {
		return
	}

	c.edges = append(c.edges, &node{
		ID: s,
	})
}

func (c Container) WaitFor(s service, deps ...service) {
	if s == nil {
		panic("nil service")
	}

	if len(deps) == 0 {
		panic("no dependencies")
	}

	cn := node(c)

	n, found := cn.find(s)

	if !found {
		panic("service not found")
	}

	for _, dep := range deps {
		ed := cn.add(dep)
		if ed != nil {
			cn.addTo(n, ed)
		}
	}
}

func (c *Container) Len() (total int) {
	total = 1 // we

	for _, edge := range c.edges {
		total += edge.Len()
	}

	total--

	return
}

func (c *Container) Shutdown() error {
	for _, edge := range c.edges {
		if !edge.ID.Running() {
			continue
		}

		// recursive async function that stops parent
		// then all deps
	}

	return nil
}

func (c *Container) startService(s service) {
	if s.Running() {
		return
	}

	n, found := node(*c).find(s)

	if !found {
		panic("service not found")
	}

	for _, dep := range n.edges {
		if !dep.ID.Running() {
			c.startService(dep.ID)
		}
	}

	s.Start()
}

func (c *Container) Start() {
	for _, s := range node(*c).edges {
		c.startService(s.ID)
	}
}
