// Package radix implements a generic radix tree (also known as a compact prefix
// tree or patricia trie). A radix tree stores strings efficiently by sharing
// common prefixes, making it ideal for tasks like routing tables, autocomplete
// systems, and IP lookups.
//
// This implementation uses Go generics so that any value type can be associated
// with the stored keys. Edges within each node are kept sorted by label byte,
// enabling O(log K) child lookups via binary search, where K is the number of
// children at a given node.
//
// # Getting Started
//
// Create a tree with [New], then use [Tree.Insert] and [Tree.Get] for basic
// operations:
//
//	tree := radix.New[string]()
//	tree.Insert("foo", "bar")
//	val, ok := tree.Get("foo") // "bar", true
//
// # Prefix Operations
//
// The tree supports powerful prefix-based queries:
//
//   - [Tree.LongestPrefix] finds the longest stored key that is a prefix of a
//     given query string — useful for routing and configuration lookups.
//   - [Tree.WalkPrefix] iterates all entries whose keys share a given prefix.
//
// # Iteration
//
// [Tree.Walk] visits every key-value pair in lexicographic order. [Tree.Keys]
// and [Tree.ToMap] provide convenient bulk exports.
package radix
