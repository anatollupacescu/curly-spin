package starter

type node struct {
	ID    service
	edges []*node
}

func (n *node) equals(match *node) bool {
	return n.ID == match.ID
}

func (g *node) add(s service) (nd *node) {
	if g.has(s) {
		return nil
	}
	nd = &node{ID: s}
	g.edges = append(g.edges, nd)
	return
}

func (g *node) has(id service) bool {
	_, found := g.find(id)
	return found
}

func (g *node) addTo(target, edge *node) {
	nd, found := g.find(target.ID)
	if !found {
		panic("target not found")
	}

	if nd.has(edge.ID) {
		return
	}

	nd.edges = append(nd.edges, edge)
	g.remove(edge)
}

//go:nolint
func (n *node) remove(match *node) {
	num := 0
	for _, child := range n.edges {
		if !child.equals(match) {
			n.edges[num] = child
			num++
		}
	}

	n.edges = n.edges[:num]
}

func (g node) find(id service) (target *node, found bool) {
	g.walk(func(n *node) {
		if found {
			return
		}
		if n.ID == id {
			found = true
			target = n
		}
	})

	return
}

// depth first
// TODO early stop
func (g *node) walk(f func(*node)) {
	for _, child := range g.edges {
		child.walk(f)
	}
	f(g)
}

func (g *node) Len() (total int) {
	total = 1 // we

	for _, edge := range g.edges {
		total += edge.Len()
	}

	return
}
