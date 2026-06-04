package radix

import "sort"

// edge represents a branch in the tree. It stores the first byte of the
// child's prefix alongside the child pointer so that binary searches over a
// node's children can be performed without dereferencing into the child node.
type edge[V any] struct {
	// label is the first character of the child's prefix.
	// Storing this explicitly allows for very fast binary searches
	// without needing to dereference the node pointer.
	label byte
	node  *node[V]
}

// node represents a single node within the radix tree.
type node[V any] struct {
	// prefix is the compressed path segment associated with this node.
	prefix string

	// val holds the value stored at this node.
	val V

	// isLeaf indicates whether this node represents a complete key.
	// This is necessary because the zero value of V might be a valid insertion.
	isLeaf bool

	// edges contains the children of this node, sorted by edge.label.
	// A slice is used instead of a map to save memory and improve cache locality.
	edges []edge[V]
}

// findEdge performs a binary search over the node's sorted edges and returns
// the index and a boolean indicating whether an edge with the given label was
// found.
func (n *node[V]) findEdge(label byte) (int, bool) {
	idx := sort.Search(len(n.edges), func(i int) bool {
		return n.edges[i].label >= label
	})
	if idx < len(n.edges) && n.edges[idx].label == label {
		return idx, true
	}
	return idx, false
}

// addEdge inserts a new edge into the node's edge list, maintaining sorted
// order by label. The caller must ensure that no edge with the same label
// already exists.
func (n *node[V]) addEdge(e edge[V]) {
	idx, _ := n.findEdge(e.label)
	// Grow the slice by one element and shift everything after idx to the right.
	n.edges = append(n.edges, edge[V]{})
	copy(n.edges[idx+1:], n.edges[idx:])
	n.edges[idx] = e
}

// replaceEdge replaces the child pointer for an existing edge identified by
// label. It panics if no edge with the given label exists.
func (n *node[V]) replaceEdge(e edge[V]) {
	idx, found := n.findEdge(e.label)
	if !found {
		panic("radix: replacing a non-existent edge")
	}
	n.edges[idx] = e
}

// delEdge removes the edge with the given label from the node's edge list.
func (n *node[V]) delEdge(label byte) {
	idx, found := n.findEdge(label)
	if !found {
		return
	}
	n.edges = append(n.edges[:idx], n.edges[idx+1:]...)
}

// longestCommonPrefix returns the length of the longest common prefix shared
// by a and b.
func longestCommonPrefix(a, b string) int {
	max := len(a)
	if len(b) < max {
		max = len(b)
	}
	for i := 0; i < max; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}
