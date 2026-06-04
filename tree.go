package radix

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
