# radix-kit

A high-performance, memory-efficient **Radix Tree** (compact prefix tree) library for Go, built with **Generics**.

Radix trees store strings efficiently by sharing common prefixes, making them ideal for routing tables, autocomplete systems, IP lookups, and any scenario where prefix-based operations matter.

## Features

- **Generic values** — store any type (`int`, `string`, structs, pointers) via `Tree[V any]`
- **Memory efficient** — sorted edge slices with binary search instead of maps
- **Prefix operations** — `LongestPrefix`, `WalkPrefix` for powerful prefix-based queries
- **Compaction on delete** — automatically merges single-child nodes to maintain a compact tree
- **Lexicographic iteration** — `Walk`, `Keys`, `ToMap` all return results in sorted order
- **Zero dependencies** — only uses the Go standard library

## Installation

```bash
go get github.com/swanandvk/radix-kit
```

Requires **Go 1.20+**.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/swanandvk/radix-kit"
)

func main() {
    // Create a tree that maps strings to integers.
    tree := radix.New[int]()

    // Insert key-value pairs.
    tree.Insert("api/users", 1)
    tree.Insert("api/users/list", 2)
    tree.Insert("api/orders", 3)
    tree.Insert("web/index", 4)

    fmt.Println("Tree size:", tree.Len()) // 4
}
```

## Usage Examples

### Insert and Get

```go
tree := radix.New[string]()

// Insert returns (oldValue, wasUpdated).
_, isNew := tree.Insert("foo", "bar")
fmt.Println(isNew) // false — key didn't exist before

old, replaced := tree.Insert("foo", "baz")
fmt.Println(old, replaced) // "bar" true — previous value returned

// Get returns (value, found).
val, ok := tree.Get("foo")
fmt.Println(val, ok) // "baz" true

_, ok = tree.Get("missing")
fmt.Println(ok) // false
```

### Delete

```go
tree := radix.New[int]()
tree.Insert("alpha", 1)
tree.Insert("alphabet", 2)

// Delete returns (oldValue, wasFound).
old, found := tree.Delete("alphabet")
fmt.Println(old, found) // 2 true

// The tree automatically compacts nodes after deletion.
val, ok := tree.Get("alpha")
fmt.Println(val, ok) // 1 true
```

### Longest Prefix Match

Perfect for routing tables or configuration lookups:

```go
tree := radix.New[string]()
tree.Insert("/api", "api-handler")
tree.Insert("/api/v2", "v2-handler")
tree.Insert("/api/v2/users", "users-handler")

// Find the longest stored key that is a prefix of the query.
key, handler, ok := tree.LongestPrefix("/api/v2/users/123")
fmt.Println(key, handler) // "/api/v2/users" "users-handler"

key, handler, ok = tree.LongestPrefix("/api/v3/new")
fmt.Println(key, handler) // "/api" "api-handler"
```

### Walking the Tree

```go
tree := radix.New[int]()
tree.Insert("banana", 1)
tree.Insert("apple", 2)
tree.Insert("avocado", 3)
tree.Insert("blueberry", 4)

// Walk visits all entries in lexicographic order.
tree.Walk(func(key string, val int) bool {
    fmt.Printf("%s → %d\n", key, val)
    return true // return false to stop early
})
// Output:
//   apple → 2
//   avocado → 3
//   banana → 1
//   blueberry → 4
```

### Walking by Prefix

```go
tree := radix.New[string]()
tree.Insert("config/db/host", "localhost")
tree.Insert("config/db/port", "5432")
tree.Insert("config/cache/ttl", "60s")
tree.Insert("config/app/name", "myapp")

// Visit only keys under "config/db/".
tree.WalkPrefix("config/db/", func(key string, val string) bool {
    fmt.Printf("%s = %s\n", key, val)
    return true
})
// Output:
//   config/db/host = localhost
//   config/db/port = 5432
```

### Bulk Export

```go
tree := radix.New[int]()
tree.Insert("c", 3)
tree.Insert("a", 1)
tree.Insert("b", 2)

// Keys() returns all keys in sorted order.
fmt.Println(tree.Keys()) // [a b c]

// ToMap() returns all key-value pairs as a Go map.
fmt.Println(tree.ToMap()) // map[a:1 b:2 c:3]
```

### Using with Custom Types

```go
type Route struct {
    Handler string
    Methods []string
}

tree := radix.New[Route]()
tree.Insert("/users", Route{
    Handler: "UsersController",
    Methods: []string{"GET", "POST"},
})
tree.Insert("/users/:id", Route{
    Handler: "UserController",
    Methods: []string{"GET", "PUT", "DELETE"},
})

route, ok := tree.Get("/users")
fmt.Println(route.Handler, route.Methods)
// UsersController [GET POST]
```

## API Reference

| Method | Signature | Description |
|--------|-----------|-------------|
| `New` | `New[V any]() *Tree[V]` | Create an empty tree |
| `Insert` | `(t *Tree[V]) Insert(key string, val V) (V, bool)` | Insert or update a key; returns old value if updated |
| `Get` | `(t *Tree[V]) Get(key string) (V, bool)` | Look up a key |
| `Delete` | `(t *Tree[V]) Delete(key string) (V, bool)` | Remove a key; returns old value if found |
| `Len` | `(t *Tree[V]) Len() int` | Number of stored keys |
| `LongestPrefix` | `(t *Tree[V]) LongestPrefix(key string) (string, V, bool)` | Find longest key that prefixes the query |
| `Walk` | `(t *Tree[V]) Walk(fn func(key string, val V) bool)` | Visit all entries in sorted order |
| `WalkPrefix` | `(t *Tree[V]) WalkPrefix(prefix string, fn func(key string, val V) bool)` | Visit entries matching a prefix |
| `Keys` | `(t *Tree[V]) Keys() []string` | All keys in sorted order |
| `ToMap` | `(t *Tree[V]) ToMap() map[string]V` | Export as a Go map |

## Design

- **Sorted edge slices** with `sort.Search` for O(log K) child lookups per node (K = number of children)
- **Explicit `isLeaf` flag** to distinguish stored zero-values from absent keys
- **Automatic compaction** on delete: merges single-child non-leaf nodes back together
- **No external dependencies** — only uses Go's standard library

## License

See [LICENSE](LICENSE) for details.
