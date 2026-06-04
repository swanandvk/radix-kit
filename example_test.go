package radix_test

import (
	"fmt"

	radix "github.com/swanandvk/radix-kit"
)

func ExampleNew() {
	tree := radix.New[string]()
	tree.Insert("hello", "world")

	val, ok := tree.Get("hello")
	fmt.Println(val, ok)
	// Output: world true
}

func ExampleTree_Insert() {
	tree := radix.New[int]()

	// First insert: key is new.
	_, replaced := tree.Insert("score", 100)
	fmt.Println("replaced:", replaced)

	// Second insert: key exists, returns old value.
	old, replaced := tree.Insert("score", 200)
	fmt.Println("old:", old, "replaced:", replaced)

	// Output:
	// replaced: false
	// old: 100 replaced: true
}

func ExampleTree_Get() {
	tree := radix.New[string]()
	tree.Insert("foo", "bar")

	val, ok := tree.Get("foo")
	fmt.Println(val, ok)

	val, ok = tree.Get("missing")
	fmt.Println(val, ok)

	// Output:
	// bar true
	//  false
}

func ExampleTree_Delete() {
	tree := radix.New[string]()
	tree.Insert("temporary", "data")
	fmt.Println("before:", tree.Len())

	tree.Delete("temporary")
	fmt.Println("after:", tree.Len())

	// Output:
	// before: 1
	// after: 0
}

func ExampleTree_LongestPrefix() {
	tree := radix.New[string]()
	tree.Insert("/api", "api-handler")
	tree.Insert("/api/v2", "v2-handler")
	tree.Insert("/api/v2/users", "users-handler")

	key, val, _ := tree.LongestPrefix("/api/v2/users/123")
	fmt.Println(key, "→", val)

	key, val, _ = tree.LongestPrefix("/api/v3/new")
	fmt.Println(key, "→", val)

	// Output:
	// /api/v2/users → users-handler
	// /api → api-handler
}

func ExampleTree_Walk() {
	tree := radix.New[int]()
	tree.Insert("cherry", 3)
	tree.Insert("apple", 1)
	tree.Insert("banana", 2)

	tree.Walk(func(key string, val int) bool {
		fmt.Printf("%s=%d\n", key, val)
		return true
	})

	// Output:
	// apple=1
	// banana=2
	// cherry=3
}

func ExampleTree_WalkPrefix() {
	tree := radix.New[string]()
	tree.Insert("config/db/host", "localhost")
	tree.Insert("config/db/port", "5432")
	tree.Insert("config/cache/ttl", "60s")

	tree.WalkPrefix("config/db/", func(key string, val string) bool {
		fmt.Printf("%s = %s\n", key, val)
		return true
	})

	// Output:
	// config/db/host = localhost
	// config/db/port = 5432
}

func ExampleTree_Keys() {
	tree := radix.New[int]()
	tree.Insert("c", 3)
	tree.Insert("a", 1)
	tree.Insert("b", 2)

	fmt.Println(tree.Keys())

	// Output:
	// [a b c]
}

func ExampleTree_ToMap() {
	tree := radix.New[int]()
	tree.Insert("x", 1)
	tree.Insert("y", 2)

	m := tree.ToMap()
	fmt.Println(m["x"], m["y"])

	// Output:
	// 1 2
}
