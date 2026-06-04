package radix

// Delete removes a key from the tree.
//
// It returns the value that was associated with the key and true if the key
// existed. If the key was not found, it returns the zero value of V and false.
//
// After removing a leaf, Delete compacts the tree by merging any node that has
// exactly one child and is not itself a leaf back into its single child.
func (t *Tree[V]) Delete(key string) (V, bool) {
	var zeroVal V
	n := t.root
	search := key

	// parent tracks the parent of the current node so we can perform
	// compaction after deletion. parentLabel is the edge label from parent
	// to n.
	var parent *node[V]
	var parentLabel byte

	for {
		// Search key exhausted – check if this node holds the key.
		if len(search) == 0 {
			if !n.isLeaf {
				return zeroVal, false
			}
			break
		}

		// Follow the next edge.
		idx, found := n.findEdge(search[0])
		if !found {
			return zeroVal, false
		}

		child := n.edges[idx].node

		// Verify that the child's prefix matches the search key.
		if len(search) < len(child.prefix) || search[:len(child.prefix)] != child.prefix {
			return zeroVal, false
		}

		parent = n
		parentLabel = search[0]
		n = child
		search = search[len(child.prefix):]
	}

	// We've found the node to delete.
	old := n.val
	n.isLeaf = false
	n.val = zeroVal
	t.size--

	// Compaction: if the deleted node has no children, remove it from its
	// parent entirely.
	if len(n.edges) == 0 && parent != nil {
		parent.delEdge(parentLabel)

		// After removing the child, if the parent is not a leaf and has
		// exactly one remaining child, merge the parent with that child.
		if !parent.isLeaf && len(parent.edges) == 1 && parent != t.root {
			mergeChild := parent.edges[0].node
			parent.prefix = parent.prefix + mergeChild.prefix
			parent.val = mergeChild.val
			parent.isLeaf = mergeChild.isLeaf
			parent.edges = mergeChild.edges
		}
		return old, true
	}

	// If the node still exists (has children) but is no longer a leaf, and
	// it has exactly one child, merge it with that child. The root node is
	// excluded because it must remain the tree's entry point with an empty
	// prefix.
	if n != t.root && !n.isLeaf && len(n.edges) == 1 {
		mergeChild := n.edges[0].node
		n.prefix = n.prefix + mergeChild.prefix
		n.val = mergeChild.val
		n.isLeaf = mergeChild.isLeaf
		n.edges = mergeChild.edges
	}

	return old, true
}
