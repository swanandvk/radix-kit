package radix

// Walk visits every key-value pair in the tree in lexicographic order.
//
// The callback fn is invoked for each entry. If fn returns false, the walk is
// terminated early.
func (t *Tree[V]) Walk(fn func(key string, val V) bool) {
	walkRecursive(t.root, "", fn)
}

// WalkPrefix visits every key-value pair whose key starts with the given
// prefix, in lexicographic order.
//
// The callback fn is invoked for each matching entry. If fn returns false, the
// walk is terminated early.
func (t *Tree[V]) WalkPrefix(prefix string, fn func(key string, val V) bool) {
	n := t.root
	search := prefix

	for {
		// If the prefix is fully consumed, we've found the subtree.
		if len(search) == 0 {
			walkRecursive(n, prefix[:len(prefix)-len(search)], fn)
			return
		}

		// Try to follow the next edge.
		idx, found := n.findEdge(search[0])
		if !found {
			return
		}

		child := n.edges[idx].node

		// If the remaining search key is shorter than the child's prefix, the
		// child's prefix must start with the search key for any matches to
		// exist.
		if len(search) <= len(child.prefix) {
			if child.prefix[:len(search)] != search {
				return
			}
			// The entire prefix is consumed within this child's prefix.
			walkRecursive(child, prefix[:len(prefix)-len(search)]+child.prefix, fn)
			return
		}

		// The search key is longer than the child's prefix – the child's prefix
		// must match exactly.
		if search[:len(child.prefix)] != child.prefix {
			return
		}

		n = child
		search = search[len(child.prefix):]
	}
}

// walkRecursive performs a depth-first traversal of the subtree rooted at n.
// The accumulated key prefix is passed in via key. Returns false if the walk
// was terminated early by fn.
func walkRecursive[V any](n *node[V], key string, fn func(string, V) bool) bool {
	if n.isLeaf {
		if !fn(key, n.val) {
			return false
		}
	}
	for _, e := range n.edges {
		if !walkRecursive(e.node, key+e.node.prefix, fn) {
			return false
		}
	}
	return true
}

// Keys returns all keys stored in the tree in lexicographic order.
func (t *Tree[V]) Keys() []string {
	keys := make([]string, 0, t.size)
	t.Walk(func(key string, _ V) bool {
		keys = append(keys, key)
		return true
	})
	return keys
}

// ToMap returns all key-value pairs stored in the tree as a map.
func (t *Tree[V]) ToMap() map[string]V {
	m := make(map[string]V, t.size)
	t.Walk(func(key string, val V) bool {
		m[key] = val
		return true
	})
	return m
}
