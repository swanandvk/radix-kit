package radix

// Get looks up a key in the tree and returns the associated value.
//
// The second return value is true if the key was found, false otherwise.
// When the key is not found, the zero value of V is returned.
func (t *Tree[V]) Get(key string) (V, bool) {
	var zeroVal V
	n := t.root
	search := key

	for {
		// If the search key is exhausted, check if the current node is a leaf.
		if len(search) == 0 {
			if n.isLeaf {
				return n.val, true
			}
			return zeroVal, false
		}

		// Find the edge matching the next byte of the search key.
		idx, found := n.findEdge(search[0])
		if !found {
			return zeroVal, false
		}

		child := n.edges[idx].node

		// The child's prefix must be a prefix of the remaining search key.
		if len(search) < len(child.prefix) || search[:len(child.prefix)] != child.prefix {
			return zeroVal, false
		}

		// Consume the matched prefix and continue.
		n = child
		search = search[len(child.prefix):]
	}
}

// LongestPrefix searches the tree for the longest key that is a prefix of the
// given input string.
//
// For example, if the tree contains "foo", "foobar", and "foobarbaz", then
// LongestPrefix("foobarbazqux") returns ("foobarbaz", value, true).
//
// If no prefix match is found, it returns ("", zero, false).
func (t *Tree[V]) LongestPrefix(key string) (string, V, bool) {
	var zeroVal V
	var lastMatchKey string
	var lastMatchVal V
	var lastMatchFound bool

	n := t.root
	search := key

	for {
		// If the current node is a leaf, it represents a valid prefix of key.
		if n.isLeaf {
			lastMatchKey = key[:len(key)-len(search)]
			lastMatchVal = n.val
			lastMatchFound = true
		}

		// If the search key is exhausted, we're done.
		if len(search) == 0 {
			break
		}

		// Try to follow the next edge.
		idx, found := n.findEdge(search[0])
		if !found {
			break
		}

		child := n.edges[idx].node

		// The child's prefix must match the corresponding portion of search.
		if len(search) < len(child.prefix) || search[:len(child.prefix)] != child.prefix {
			break
		}

		n = child
		search = search[len(child.prefix):]
	}

	if lastMatchFound {
		return lastMatchKey, lastMatchVal, true
	}
	return "", zeroVal, false
}
