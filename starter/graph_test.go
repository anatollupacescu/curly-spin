package starter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) {
	var (
		graph  node
		n1, n2 *node
	)

	s1 := &serviceMock{1}
	s2 := &serviceMock{2}

	{ // s1
		_, found := graph.find(s1)
		assert.False(t, found)
		assert.Len(t, graph.edges, 0)

		n1 = graph.add(s1)

		_, found = graph.find(s1)
		assert.Len(t, graph.edges, 1)
		assert.True(t, found)
	}

	{ // s2
		_, found := graph.find(s2)
		assert.False(t, found)

		assert.Len(t, n1.edges, 0)

		n2 = graph.add(s2)

		assert.Len(t, graph.edges, 2)

		_, found = graph.find(s2)
		assert.True(t, found)
	}

	graph.addTo(n2, n1)

	assert.True(t, graph.has(n2.ID))
	assert.Len(t, n1.edges, 0)
	assert.True(t, n2.has(n1.ID))
}

type serviceMock struct {
	id int // empty struct{} comparison is tricky
}

func (s *serviceMock) Running() bool {
	return false
}

func (s *serviceMock) Stop() {
}

func (s *serviceMock) Start() {
}
