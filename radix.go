// Package radix implements a generic radix tree (also known as a compact prefix
// tree or patricia trie). A radix tree stores strings efficiently by sharing
// common prefixes, making it ideal for tasks like routing tables, autocomplete
// systems, and IP lookups.
//
// This implementation uses Go generics so that any value type can be associated
// with the stored keys. Edges within each node are kept sorted by label byte,
// enabling O(log n) child lookups via binary search.
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

// Tree is the main structure for a generic radix tree. It maps string keys to
// values of type V. The zero value is not usable; create instances with [New].
type Tree[V any] struct {
	root *node[V]
	size int
}

// New creates and returns an empty radix tree.
func New[V any]() *Tree[V] {
	return &Tree[V]{
		root: &node[V]{},
	}
}

// Len returns the number of keys stored in the tree.
func (t *Tree[V]) Len() int {
	return t.size
}
