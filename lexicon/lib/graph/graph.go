package graph

import (
	"github.com/codelingo/lexicon/lib/graph/edge"
)

// TODO(waigani) find a better home for this. Probably should live in CLQL.

// TODO(waigani) See if we need PropertyFinder. The idea is, Node.properties
// would be a map[string]PropertyFinder. The funcs would be set when the Fact
// is deduced, but they are not called until queried by a leaf node. For now,
// we set all the properties of the node, even if some are not used in the
// query.

// possibly return Node here
// type PropertyFinder func() interface{}

// Node represents a node in a tree graph.
type Node struct {
	uuid int

	// Node edges are built up by each fact scope.
	edges []edge.Edge

	// Properties can be queried by leaf nodes.
	properties map[string]interface{}
}

func (n *Node) Property(key string) interface{} {
	if prop, ok := n.properties[key]; ok {
		return prop
	}
	return nil
}

func NewNode(uuid int, properties map[string]interface{}, edges ...edge.Edge) *Node {
	if properties == nil {
		properties = map[string]interface{}{}
	}

	return &Node{
		uuid:       uuid,
		edges:      edges,
		properties: properties,
	}
}

func (n *Node) AddEdge(e edge.Edge) {
	// TODO(waigani) allow only one edge for each dimension. The edge
	// should be added by the closest parent node for each dimension.
	n.edges = append(n.edges, e)
}

// Nodes only have one edge per dimension.
func (n *Node) GetEdge(d edge.Dimension) edge.Edge {
	for _, e := range n.edges {
		if e.Dimension() == d {
			return e
		}
	}
	return nil
}
