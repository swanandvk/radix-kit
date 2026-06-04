package radix

// Insert adds or updates a key-value pair in the tree.
//
// If the key already existed, Insert returns the previous value and true.
// If the key is new, it returns the zero value of V and false.
func (t *Tree[V]) Insert(key string, val V) (V, bool) {
	var zeroVal V
	n := t.root
	search := key

	for {
		// If the search key is exhausted, we've arrived at the target node.
		if len(search) == 0 {
			if n.isLeaf {
				old := n.val
				n.val = val
				return old, true
			}
			n.val = val
			n.isLeaf = true
			t.size++
			return zeroVal, false
		}

		// Look for an edge whose label matches the next byte of the search key.
		idx, found := n.findEdge(search[0])
		if !found {
			// No matching edge – create a new child node.
			n.addEdge(edge[V]{
				label: search[0],
				node: &node[V]{
					prefix: search,
					val:    val,
					isLeaf: true,
				},
			})
			t.size++
			return zeroVal, false
		}

		child := n.edges[idx].node
		commonLen := longestCommonPrefix(search, child.prefix)

		// Case 1: The child's prefix matches completely – descend into the child.
		if commonLen == len(child.prefix) {
			n = child
			search = search[commonLen:]
			continue
		}

		// Case 2: Partial match – we need to split the existing child node.
		// Create a new intermediate node that holds the shared prefix.
		splitNode := &node[V]{
			prefix: search[:commonLen],
		}

		// The existing child becomes a child of the split node, with its
		// prefix trimmed to the non-shared suffix.
		child.prefix = child.prefix[commonLen:]
		splitNode.addEdge(edge[V]{
			label: child.prefix[0],
			node:  child,
		})

		// Replace the edge in the parent to point to the split node.
		n.replaceEdge(edge[V]{
			label: splitNode.prefix[0],
			node:  splitNode,
		})

		// If the search key is fully consumed by the common prefix, the split
		// node itself becomes a leaf.
		if commonLen == len(search) {
			splitNode.val = val
			splitNode.isLeaf = true
			t.size++
			return zeroVal, false
		}

		// Otherwise, create a new leaf for the remaining search key.
		splitNode.addEdge(edge[V]{
			label: search[commonLen],
			node: &node[V]{
				prefix: search[commonLen:],
				val:    val,
				isLeaf: true,
			},
		})
		t.size++
		return zeroVal, false
	}
}
