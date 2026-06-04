package radix

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// --------------------------------------------------------------------------
// Insert + Get
// --------------------------------------------------------------------------

func TestInsertAndGet(t *testing.T) {
	tests := []struct {
		name    string
		entries []struct{ k, v string }
		lookups []struct {
			key    string
			wantV  string
			wantOK bool
		}
	}{
		{
			name: "single key",
			entries: []struct{ k, v string }{
				{"hello", "world"},
			},
			lookups: []struct {
				key    string
				wantV  string
				wantOK bool
			}{
				{"hello", "world", true},
				{"hell", "", false},
				{"helloo", "", false},
				{"", "", false},
			},
		},
		{
			name: "prefix splitting",
			entries: []struct{ k, v string }{
				{"test", "v1"},
				{"team", "v2"},
				{"toast", "v3"},
				{"to", "v4"},
			},
			lookups: []struct {
				key    string
				wantV  string
				wantOK bool
			}{
				{"test", "v1", true},
				{"team", "v2", true},
				{"toast", "v3", true},
				{"to", "v4", true},
				{"te", "", false},
				{"t", "", false},
				{"toaster", "", false},
			},
		},
		{
			name: "shared prefixes",
			entries: []struct{ k, v string }{
				{"romane", "1"},
				{"romanus", "2"},
				{"romulus", "3"},
				{"rubens", "4"},
				{"ruber", "5"},
				{"rubicon", "6"},
				{"rubicundus", "7"},
			},
			lookups: []struct {
				key    string
				wantV  string
				wantOK bool
			}{
				{"romane", "1", true},
				{"romanus", "2", true},
				{"romulus", "3", true},
				{"rubens", "4", true},
				{"ruber", "5", true},
				{"rubicon", "6", true},
				{"rubicundus", "7", true},
				{"rom", "", false},
				{"rub", "", false},
				{"ruby", "", false},
			},
		},
		{
			name:    "empty tree",
			entries: []struct{ k, v string }{},
			lookups: []struct {
				key    string
				wantV  string
				wantOK bool
			}{
				{"anything", "", false},
				{"", "", false},
			},
		},
		{
			name: "empty string key",
			entries: []struct{ k, v string }{
				{"", "root"},
			},
			lookups: []struct {
				key    string
				wantV  string
				wantOK bool
			}{
				{"", "root", true},
				{"a", "", false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New[string]()
			for _, e := range tt.entries {
				tree.Insert(e.k, e.v)
			}
			for _, l := range tt.lookups {
				got, ok := tree.Get(l.key)
				if ok != l.wantOK || got != l.wantV {
					t.Errorf("Get(%q) = (%q, %v), want (%q, %v)",
						l.key, got, ok, l.wantV, l.wantOK)
				}
			}
		})
	}
}

// --------------------------------------------------------------------------
// Update existing key
// --------------------------------------------------------------------------

func TestInsertUpdate(t *testing.T) {
	tree := New[int]()

	_, replaced := tree.Insert("key", 1)
	if replaced {
		t.Fatal("first insert should not report replacement")
	}

	old, replaced := tree.Insert("key", 2)
	if !replaced {
		t.Fatal("second insert should report replacement")
	}
	if old != 1 {
		t.Fatalf("old value should be 1, got %d", old)
	}

	got, ok := tree.Get("key")
	if !ok || got != 2 {
		t.Fatalf("Get after update = (%d, %v), want (2, true)", got, ok)
	}

	if tree.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", tree.Len())
	}
}

// --------------------------------------------------------------------------
// Delete
// --------------------------------------------------------------------------

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		inserts    []struct{ k, v string }
		deleteKey  string
		wantVal    string
		wantFound  bool
		wantLen    int
		wantKeys   []string
	}{
		{
			name: "delete existing leaf",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
				{"foobar", "2"},
			},
			deleteKey: "foobar",
			wantVal:   "2",
			wantFound: true,
			wantLen:   1,
			wantKeys:  []string{"foo"},
		},
		{
			name: "delete non-existent key",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
			},
			deleteKey: "bar",
			wantVal:   "",
			wantFound: false,
			wantLen:   1,
			wantKeys:  []string{"foo"},
		},
		{
			name: "delete with compaction",
			inserts: []struct{ k, v string }{
				{"test", "1"},
				{"team", "2"},
				{"toast", "3"},
			},
			deleteKey: "team",
			wantVal:   "2",
			wantFound: true,
			wantLen:   2,
			wantKeys:  []string{"test", "toast"},
		},
		{
			name: "delete only key",
			inserts: []struct{ k, v string }{
				{"only", "1"},
			},
			deleteKey: "only",
			wantVal:   "1",
			wantFound: true,
			wantLen:   0,
			wantKeys:  []string{},
		},
		{
			name: "delete parent of children",
			inserts: []struct{ k, v string }{
				{"to", "1"},
				{"toast", "2"},
				{"toaster", "3"},
			},
			deleteKey: "to",
			wantVal:   "1",
			wantFound: true,
			wantLen:   2,
			wantKeys:  []string{"toast", "toaster"},
		},
		{
			name: "delete empty string key",
			inserts: []struct{ k, v string }{
				{"", "root"},
				{"a", "alpha"},
			},
			deleteKey: "",
			wantVal:   "root",
			wantFound: true,
			wantLen:   1,
			wantKeys:  []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New[string]()
			for _, e := range tt.inserts {
				tree.Insert(e.k, e.v)
			}

			val, found := tree.Delete(tt.deleteKey)
			if found != tt.wantFound || val != tt.wantVal {
				t.Errorf("Delete(%q) = (%q, %v), want (%q, %v)",
					tt.deleteKey, val, found, tt.wantVal, tt.wantFound)
			}
			if tree.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", tree.Len(), tt.wantLen)
			}

			keys := tree.Keys()
			if !reflect.DeepEqual(keys, tt.wantKeys) {
				t.Errorf("Keys() = %v, want %v", keys, tt.wantKeys)
			}
		})
	}
}

// --------------------------------------------------------------------------
// Delete all then re-insert
// --------------------------------------------------------------------------

func TestDeleteAllAndReinsert(t *testing.T) {
	tree := New[int]()
	keys := []string{"alpha", "alphabet", "alpine", "beta", "betray"}
	for i, k := range keys {
		tree.Insert(k, i)
	}

	for _, k := range keys {
		_, found := tree.Delete(k)
		if !found {
			t.Fatalf("Delete(%q) should have found the key", k)
		}
	}
	if tree.Len() != 0 {
		t.Fatalf("Len() = %d after deleting all, want 0", tree.Len())
	}

	// Re-insert
	for i, k := range keys {
		tree.Insert(k, i+100)
	}
	if tree.Len() != len(keys) {
		t.Fatalf("Len() = %d after re-insert, want %d", tree.Len(), len(keys))
	}
	for i, k := range keys {
		v, ok := tree.Get(k)
		if !ok || v != i+100 {
			t.Errorf("Get(%q) = (%d, %v), want (%d, true)", k, v, ok, i+100)
		}
	}
}

// --------------------------------------------------------------------------
// LongestPrefix
// --------------------------------------------------------------------------

func TestLongestPrefix(t *testing.T) {
	tests := []struct {
		name     string
		inserts  []struct{ k, v string }
		query    string
		wantKey  string
		wantVal  string
		wantOK   bool
	}{
		{
			name: "exact match",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
				{"foobar", "2"},
			},
			query:   "foobar",
			wantKey: "foobar",
			wantVal: "2",
			wantOK:  true,
		},
		{
			name: "prefix match",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
				{"foobar", "2"},
			},
			query:   "foobarbaz",
			wantKey: "foobar",
			wantVal: "2",
			wantOK:  true,
		},
		{
			name: "shortest prefix",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
				{"foobar", "2"},
			},
			query:   "fooXYZ",
			wantKey: "foo",
			wantVal: "1",
			wantOK:  true,
		},
		{
			name: "no match",
			inserts: []struct{ k, v string }{
				{"foo", "1"},
			},
			query:   "bar",
			wantKey: "",
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "empty string prefix",
			inserts: []struct{ k, v string }{
				{"", "root"},
				{"foo", "1"},
			},
			query:   "anything",
			wantKey: "",
			wantVal: "root",
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New[string]()
			for _, e := range tt.inserts {
				tree.Insert(e.k, e.v)
			}
			key, val, ok := tree.LongestPrefix(tt.query)
			if key != tt.wantKey || val != tt.wantVal || ok != tt.wantOK {
				t.Errorf("LongestPrefix(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.query, key, val, ok, tt.wantKey, tt.wantVal, tt.wantOK)
			}
		})
	}
}

// --------------------------------------------------------------------------
// Walk
// --------------------------------------------------------------------------

func TestWalk(t *testing.T) {
	tree := New[int]()
	input := map[string]int{
		"apple":  1,
		"app":    2,
		"banana": 3,
		"band":   4,
		"bat":    5,
	}
	for k, v := range input {
		tree.Insert(k, v)
	}

	var got []string
	tree.Walk(func(key string, _ int) bool {
		got = append(got, key)
		return true
	})

	want := []string{"app", "apple", "banana", "band", "bat"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Walk order = %v, want %v", got, want)
	}
}

func TestWalkEarlyStop(t *testing.T) {
	tree := New[int]()
	for i := 0; i < 10; i++ {
		tree.Insert(fmt.Sprintf("key%d", i), i)
	}

	count := 0
	tree.Walk(func(_ string, _ int) bool {
		count++
		return count < 3
	})

	if count != 3 {
		t.Errorf("Walk visited %d entries, want 3", count)
	}
}

// --------------------------------------------------------------------------
// WalkPrefix
// --------------------------------------------------------------------------

func TestWalkPrefix(t *testing.T) {
	tree := New[string]()
	entries := []struct{ k, v string }{
		{"api/users", "u"},
		{"api/users/list", "ul"},
		{"api/orders", "o"},
		{"web/index", "wi"},
		{"web/about", "wa"},
	}
	for _, e := range entries {
		tree.Insert(e.k, e.v)
	}

	tests := []struct {
		prefix   string
		wantKeys []string
	}{
		{"api/", []string{"api/orders", "api/users", "api/users/list"}},
		{"api/users", []string{"api/users", "api/users/list"}},
		{"web/", []string{"web/about", "web/index"}},
		{"web/a", []string{"web/about"}},
		{"missing", nil},
		{"", []string{"api/orders", "api/users", "api/users/list", "web/about", "web/index"}},
	}

	for _, tt := range tests {
		t.Run("prefix="+tt.prefix, func(t *testing.T) {
			var got []string
			tree.WalkPrefix(tt.prefix, func(key string, _ string) bool {
				got = append(got, key)
				return true
			})
			if tt.wantKeys == nil {
				tt.wantKeys = []string{}
			}
			if got == nil {
				got = []string{}
			}
			if !reflect.DeepEqual(got, tt.wantKeys) {
				t.Errorf("WalkPrefix(%q) keys = %v, want %v", tt.prefix, got, tt.wantKeys)
			}
		})
	}
}

// --------------------------------------------------------------------------
// Keys + ToMap
// --------------------------------------------------------------------------

func TestKeys(t *testing.T) {
	tree := New[int]()
	tree.Insert("charlie", 3)
	tree.Insert("alpha", 1)
	tree.Insert("bravo", 2)

	keys := tree.Keys()
	want := []string{"alpha", "bravo", "charlie"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("Keys() = %v, want %v", keys, want)
	}
}

func TestToMap(t *testing.T) {
	tree := New[int]()
	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	m := tree.ToMap()
	want := map[string]int{"a": 1, "b": 2, "c": 3}
	if !reflect.DeepEqual(m, want) {
		t.Errorf("ToMap() = %v, want %v", m, want)
	}
}

// --------------------------------------------------------------------------
// Len tracking
// --------------------------------------------------------------------------

func TestLen(t *testing.T) {
	tree := New[string]()

	if tree.Len() != 0 {
		t.Fatalf("empty tree Len() = %d, want 0", tree.Len())
	}

	tree.Insert("a", "1")
	tree.Insert("b", "2")
	if tree.Len() != 2 {
		t.Fatalf("Len() = %d after 2 inserts, want 2", tree.Len())
	}

	// Update should not change len
	tree.Insert("a", "updated")
	if tree.Len() != 2 {
		t.Fatalf("Len() = %d after update, want 2", tree.Len())
	}

	tree.Delete("a")
	if tree.Len() != 1 {
		t.Fatalf("Len() = %d after delete, want 1", tree.Len())
	}

	// Delete non-existent
	tree.Delete("zzz")
	if tree.Len() != 1 {
		t.Fatalf("Len() = %d after deleting non-existent, want 1", tree.Len())
	}
}

// --------------------------------------------------------------------------
// Empty tree operations
// --------------------------------------------------------------------------

func TestEmptyTree(t *testing.T) {
	tree := New[string]()

	if _, ok := tree.Get("anything"); ok {
		t.Error("Get on empty tree should return false")
	}

	if _, ok := tree.Delete("anything"); ok {
		t.Error("Delete on empty tree should return false")
	}

	if _, _, ok := tree.LongestPrefix("anything"); ok {
		t.Error("LongestPrefix on empty tree should return false")
	}

	keys := tree.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys() on empty tree = %v, want []", keys)
	}

	m := tree.ToMap()
	if len(m) != 0 {
		t.Errorf("ToMap() on empty tree = %v, want {}", m)
	}

	count := 0
	tree.Walk(func(_ string, _ string) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Walk visited %d entries on empty tree, want 0", count)
	}
}

// --------------------------------------------------------------------------
// Large number of insertions
// --------------------------------------------------------------------------

func TestLargeInsertions(t *testing.T) {
	tree := New[int]()
	const n = 10000

	// Insert n keys
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key/%05d/value", i)
		tree.Insert(key, i)
	}

	if tree.Len() != n {
		t.Fatalf("Len() = %d, want %d", tree.Len(), n)
	}

	// Verify all keys are retrievable
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key/%05d/value", i)
		v, ok := tree.Get(key)
		if !ok || v != i {
			t.Fatalf("Get(%q) = (%d, %v), want (%d, true)", key, v, ok, i)
		}
	}

	// Keys should be sorted
	keys := tree.Keys()
	if !sort.StringsAreSorted(keys) {
		t.Fatal("Keys() are not sorted")
	}

	// Delete all keys
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key/%05d/value", i)
		_, ok := tree.Delete(key)
		if !ok {
			t.Fatalf("Delete(%q) should have found the key", key)
		}
	}

	if tree.Len() != 0 {
		t.Fatalf("Len() = %d after deleting all, want 0", tree.Len())
	}
}

// --------------------------------------------------------------------------
// Generic type usage
// --------------------------------------------------------------------------

func TestGenericInt(t *testing.T) {
	tree := New[int]()
	tree.Insert("one", 1)
	tree.Insert("two", 2)
	tree.Insert("zero", 0) // zero value is valid

	v, ok := tree.Get("one")
	if !ok || v != 1 {
		t.Errorf("Get(one) = (%d, %v), want (1, true)", v, ok)
	}

	v, ok = tree.Get("zero")
	if !ok || v != 0 {
		t.Errorf("Get(zero) = (%d, %v), want (0, true)", v, ok)
	}
}

func TestGenericStruct(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	tree := New[User]()
	tree.Insert("alice", User{"Alice", 30})
	tree.Insert("bob", User{"Bob", 25})

	v, ok := tree.Get("alice")
	if !ok || v.Name != "Alice" || v.Age != 30 {
		t.Errorf("Get(alice) = (%+v, %v), want ({Alice 30}, true)", v, ok)
	}

	v, ok = tree.Get("bob")
	if !ok || v.Name != "Bob" || v.Age != 25 {
		t.Errorf("Get(bob) = (%+v, %v), want ({Bob 25}, true)", v, ok)
	}
}

func TestGenericPointer(t *testing.T) {
	tree := New[*int]()
	a, b := 42, 0
	tree.Insert("a", &a)
	tree.Insert("b", &b)
	tree.Insert("nil", nil) // nil pointer is valid

	v, ok := tree.Get("a")
	if !ok || *v != 42 {
		t.Errorf("Get(a) = (%v, %v), want (ptr to 42, true)", v, ok)
	}

	v, ok = tree.Get("nil")
	if !ok || v != nil {
		t.Errorf("Get(nil) = (%v, %v), want (nil, true)", v, ok)
	}
}

// --------------------------------------------------------------------------
// Node compaction verification
// --------------------------------------------------------------------------

func TestCompaction(t *testing.T) {
	// Insert three keys that share a prefix, then delete the middle one
	// to verify that compaction merges nodes correctly.
	tree := New[string]()
	tree.Insert("test", "v1")
	tree.Insert("team", "v2")

	// At this point the tree should have:
	// root -> "te" -> "st" (leaf) and "am" (leaf)

	tree.Delete("team")

	// After deletion, "te" + "st" should be compacted into "test".
	// Verify the key is still accessible.
	v, ok := tree.Get("test")
	if !ok || v != "v1" {
		t.Errorf("Get(test) after compaction = (%q, %v), want (v1, true)", v, ok)
	}

	// Can still insert new keys that share the prefix.
	tree.Insert("tea", "v3")
	v, ok = tree.Get("tea")
	if !ok || v != "v3" {
		t.Errorf("Get(tea) = (%q, %v), want (v3, true)", v, ok)
	}
	v, ok = tree.Get("test")
	if !ok || v != "v1" {
		t.Errorf("Get(test) = (%q, %v), want (v1, true)", v, ok)
	}
}

// --------------------------------------------------------------------------
// WalkPrefix with partial edge match
// --------------------------------------------------------------------------

func TestWalkPrefixPartialEdge(t *testing.T) {
	tree := New[int]()
	tree.Insert("foobar", 1)
	tree.Insert("foobaz", 2)

	// Prefix "fooba" partially matches the edge "foobar" and "foobaz"
	var got []string
	tree.WalkPrefix("fooba", func(key string, _ int) bool {
		got = append(got, key)
		return true
	})
	want := []string{"foobar", "foobaz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkPrefix(fooba) = %v, want %v", got, want)
	}
}
