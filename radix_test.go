package radix

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// ---------------------------------------------------------------------------
// Insert
// ---------------------------------------------------------------------------

func TestInsert_SingleKey_IsRetrievableByGet(t *testing.T) {
	tree := New[string]()
	tree.Insert("hello", "world")

	got, ok := tree.Get("hello")
	if !ok || got != "world" {
		t.Errorf("Get(hello) = (%q, %v), want (world, true)", got, ok)
	}
}

func TestInsert_ReturnsZeroValueAndFalse_WhenKeyIsNew(t *testing.T) {
	tree := New[int]()

	old, replaced := tree.Insert("brand-new", 42)
	if replaced {
		t.Error("expected replaced=false for a new key")
	}
	if old != 0 {
		t.Errorf("expected zero-value (0) for new key, got %d", old)
	}
}

func TestInsert_ReturnsPreviousValueAndTrue_WhenKeyAlreadyExists(t *testing.T) {
	tree := New[int]()
	tree.Insert("key", 1)

	old, replaced := tree.Insert("key", 2)
	if !replaced {
		t.Error("expected replaced=true for existing key")
	}
	if old != 1 {
		t.Errorf("expected previous value 1, got %d", old)
	}

	got, _ := tree.Get("key")
	if got != 2 {
		t.Errorf("value after update should be 2, got %d", got)
	}
}

func TestInsert_UpdatingExistingKey_DoesNotChangLen(t *testing.T) {
	tree := New[string]()
	tree.Insert("key", "v1")
	tree.Insert("key", "v2")

	if tree.Len() != 1 {
		t.Errorf("Len() = %d after insert+update, want 1", tree.Len())
	}
}

func TestInsert_EmptyStringKey_IsStoredAndRetrievable(t *testing.T) {
	tree := New[string]()
	tree.Insert("", "root-value")

	got, ok := tree.Get("")
	if !ok || got != "root-value" {
		t.Errorf("Get(\"\") = (%q, %v), want (root-value, true)", got, ok)
	}
}

func TestInsert_SplitsNodesCorrectly_WhenKeysSharePrefixes(t *testing.T) {
	// "test" and "team" share "te", forcing a split at that point.
	// "toast" and "to" share "to", with "to" being a prefix of "toast".
	tree := New[string]()
	tree.Insert("test", "v1")
	tree.Insert("team", "v2")
	tree.Insert("toast", "v3")
	tree.Insert("to", "v4")

	lookups := []struct {
		key    string
		wantV  string
		wantOK bool
	}{
		{"test", "v1", true},
		{"team", "v2", true},
		{"toast", "v3", true},
		{"to", "v4", true},
		{"te", "", false},     // shared prefix but not a stored key
		{"t", "", false},      // partial match
		{"toaster", "", false}, // extends beyond any stored key
	}

	for _, l := range lookups {
		got, ok := tree.Get(l.key)
		if ok != l.wantOK || got != l.wantV {
			t.Errorf("Get(%q) = (%q, %v), want (%q, %v)",
				l.key, got, ok, l.wantV, l.wantOK)
		}
	}
}

func TestInsert_ClassicRadixTreeExample_AllKeysRetrievable(t *testing.T) {
	// Classic Wikipedia radix tree example: romane, romanus, romulus, etc.
	tree := New[string]()
	entries := map[string]string{
		"romane":     "1",
		"romanus":    "2",
		"romulus":    "3",
		"rubens":     "4",
		"ruber":      "5",
		"rubicon":    "6",
		"rubicundus": "7",
	}
	for k, v := range entries {
		tree.Insert(k, v)
	}

	for k, want := range entries {
		got, ok := tree.Get(k)
		if !ok || got != want {
			t.Errorf("Get(%q) = (%q, %v), want (%q, true)", k, got, ok, want)
		}
	}

	// Intermediate prefixes that were never inserted should not exist.
	for _, prefix := range []string{"rom", "rub", "ruby", "r"} {
		if _, ok := tree.Get(prefix); ok {
			t.Errorf("Get(%q) should return false for non-inserted prefix", prefix)
		}
	}
}

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestGet_ReturnsZeroAndFalse_WhenKeyDoesNotExist(t *testing.T) {
	tree := New[string]()
	tree.Insert("hello", "world")

	got, ok := tree.Get("missing")
	if ok {
		t.Error("expected ok=false for missing key")
	}
	if got != "" {
		t.Errorf("expected zero value for missing key, got %q", got)
	}
}

func TestGet_ReturnsZeroAndFalse_OnEmptyTree(t *testing.T) {
	tree := New[string]()

	_, ok := tree.Get("anything")
	if ok {
		t.Error("Get on empty tree should return false")
	}
}

func TestGet_ReturnsFalse_WhenKeyIsPrefixOfStoredKeyButNotStored(t *testing.T) {
	tree := New[string]()
	tree.Insert("foobar", "value")

	_, ok := tree.Get("foo")
	if ok {
		t.Error("Get(foo) should return false when only foobar is stored")
	}
}

func TestGet_ReturnsFalse_WhenKeyExtendsStoredKey(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "value")

	_, ok := tree.Get("foobar")
	if ok {
		t.Error("Get(foobar) should return false when only foo is stored")
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_ExistingLeaf_ReturnsValueAndTrue(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "bar")
	tree.Insert("foobar", "baz")

	val, found := tree.Delete("foobar")
	if !found || val != "baz" {
		t.Errorf("Delete(foobar) = (%q, %v), want (baz, true)", val, found)
	}
	if tree.Len() != 1 {
		t.Errorf("Len() = %d, want 1", tree.Len())
	}

	// Remaining key should still work.
	v, ok := tree.Get("foo")
	if !ok || v != "bar" {
		t.Errorf("Get(foo) = (%q, %v) after delete, want (bar, true)", v, ok)
	}
}

func TestDelete_NonExistentKey_ReturnsZeroAndFalse(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "bar")

	val, found := tree.Delete("missing")
	if found {
		t.Error("expected found=false for non-existent key")
	}
	if val != "" {
		t.Errorf("expected zero value, got %q", val)
	}
	if tree.Len() != 1 {
		t.Errorf("Len() should remain 1, got %d", tree.Len())
	}
}

func TestDelete_OnlyKey_LeavesTreeEmpty(t *testing.T) {
	tree := New[string]()
	tree.Insert("only", "one")

	tree.Delete("only")

	if tree.Len() != 0 {
		t.Errorf("Len() = %d, want 0", tree.Len())
	}
	if _, ok := tree.Get("only"); ok {
		t.Error("deleted key should not be found")
	}
}

func TestDelete_ParentOfChildren_RemovesParentButKeepsChildren(t *testing.T) {
	tree := New[string]()
	tree.Insert("to", "parent")
	tree.Insert("toast", "child1")
	tree.Insert("toaster", "child2")

	tree.Delete("to")

	if _, ok := tree.Get("to"); ok {
		t.Error("deleted parent should not be found")
	}

	v1, ok1 := tree.Get("toast")
	v2, ok2 := tree.Get("toaster")
	if !ok1 || v1 != "child1" || !ok2 || v2 != "child2" {
		t.Errorf("children should still be accessible after parent deletion")
	}
}

func TestDelete_EmptyStringKey_WorksCorrectly(t *testing.T) {
	tree := New[string]()
	tree.Insert("", "root")
	tree.Insert("a", "alpha")

	val, found := tree.Delete("")
	if !found || val != "root" {
		t.Errorf("Delete(\"\") = (%q, %v), want (root, true)", val, found)
	}

	// "a" should still be accessible.
	v, ok := tree.Get("a")
	if !ok || v != "alpha" {
		t.Errorf("Get(a) = (%q, %v) after root delete, want (alpha, true)", v, ok)
	}
}

func TestDelete_CompactsNodes_WhenSiblingIsRemovedLeavingSingleChild(t *testing.T) {
	// Insert "test" and "team" which share prefix "te".
	// Deleting "team" should compact "te"+"st" back to "test".
	tree := New[string]()
	tree.Insert("test", "v1")
	tree.Insert("team", "v2")

	tree.Delete("team")

	v, ok := tree.Get("test")
	if !ok || v != "v1" {
		t.Errorf("Get(test) = (%q, %v) after compaction, want (v1, true)", v, ok)
	}

	// Inserting new keys sharing the prefix should still work after compaction.
	tree.Insert("tea", "v3")
	v, ok = tree.Get("tea")
	if !ok || v != "v3" {
		t.Errorf("Get(tea) = (%q, %v), want (v3, true)", v, ok)
	}
}

func TestDelete_AllKeys_ThenReinsert_TreeWorksCorrectly(t *testing.T) {
	tree := New[int]()
	keys := []string{"alpha", "alphabet", "alpine", "beta", "betray"}

	for i, k := range keys {
		tree.Insert(k, i)
	}
	for _, k := range keys {
		if _, found := tree.Delete(k); !found {
			t.Fatalf("Delete(%q) should have found the key", k)
		}
	}
	if tree.Len() != 0 {
		t.Fatalf("Len() = %d after deleting all, want 0", tree.Len())
	}

	// Re-insert with different values.
	for i, k := range keys {
		tree.Insert(k, i+100)
	}
	for i, k := range keys {
		v, ok := tree.Get(k)
		if !ok || v != i+100 {
			t.Errorf("Get(%q) = (%d, %v) after re-insert, want (%d, true)", k, v, ok, i+100)
		}
	}
}

func TestDelete_OnEmptyTree_ReturnsFalse(t *testing.T) {
	tree := New[string]()

	_, found := tree.Delete("anything")
	if found {
		t.Error("Delete on empty tree should return false")
	}
}

// ---------------------------------------------------------------------------
// LongestPrefix
// ---------------------------------------------------------------------------

func TestLongestPrefix_ReturnsExactMatch_WhenKeyExistsInTree(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "1")
	tree.Insert("foobar", "2")

	key, val, ok := tree.LongestPrefix("foobar")
	if !ok || key != "foobar" || val != "2" {
		t.Errorf("LongestPrefix(foobar) = (%q, %q, %v), want (foobar, 2, true)", key, val, ok)
	}
}

func TestLongestPrefix_ReturnsLongestMatchingPrefix_WhenQueryExtendsStoredKeys(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "1")
	tree.Insert("foobar", "2")

	key, val, ok := tree.LongestPrefix("foobarbaz")
	if !ok || key != "foobar" || val != "2" {
		t.Errorf("LongestPrefix(foobarbaz) = (%q, %q, %v), want (foobar, 2, true)", key, val, ok)
	}
}

func TestLongestPrefix_ReturnsShorterPrefix_WhenLongerPrefixDoesNotMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "1")
	tree.Insert("foobar", "2")

	// "fooXYZ" matches "foo" but not "foobar"
	key, val, ok := tree.LongestPrefix("fooXYZ")
	if !ok || key != "foo" || val != "1" {
		t.Errorf("LongestPrefix(fooXYZ) = (%q, %q, %v), want (foo, 1, true)", key, val, ok)
	}
}

func TestLongestPrefix_ReturnsEmptyStringPrefix_WhenItIsStored(t *testing.T) {
	tree := New[string]()
	tree.Insert("", "root")
	tree.Insert("foo", "1")

	// Even though the query doesn't match "foo", the empty prefix should match.
	key, val, ok := tree.LongestPrefix("anything")
	if !ok || key != "" || val != "root" {
		t.Errorf("LongestPrefix(anything) = (%q, %q, %v), want (\"\", root, true)", key, val, ok)
	}
}

func TestLongestPrefix_ReturnsFalse_WhenNoPrefixMatches(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "1")

	_, _, ok := tree.LongestPrefix("bar")
	if ok {
		t.Error("LongestPrefix should return false when no prefix matches")
	}
}

func TestLongestPrefix_ReturnsFalse_OnEmptyTree(t *testing.T) {
	tree := New[string]()

	_, _, ok := tree.LongestPrefix("anything")
	if ok {
		t.Error("LongestPrefix on empty tree should return false")
	}
}

// ---------------------------------------------------------------------------
// Walk
// ---------------------------------------------------------------------------

func TestWalk_VisitsAllEntries_InLexicographicOrder(t *testing.T) {
	tree := New[int]()
	tree.Insert("banana", 3)
	tree.Insert("apple", 1)
	tree.Insert("app", 2)
	tree.Insert("band", 4)
	tree.Insert("bat", 5)

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

func TestWalk_StopsEarly_WhenCallbackReturnsFalse(t *testing.T) {
	tree := New[int]()
	for i := 0; i < 10; i++ {
		tree.Insert(fmt.Sprintf("key%02d", i), i)
	}

	visited := 0
	tree.Walk(func(_ string, _ int) bool {
		visited++
		return visited < 3
	})

	if visited != 3 {
		t.Errorf("Walk visited %d entries, want exactly 3", visited)
	}
}

func TestWalk_DoesNothing_OnEmptyTree(t *testing.T) {
	tree := New[string]()

	visited := 0
	tree.Walk(func(_ string, _ string) bool {
		visited++
		return true
	})
	if visited != 0 {
		t.Errorf("Walk visited %d entries on empty tree, want 0", visited)
	}
}

// ---------------------------------------------------------------------------
// WalkPrefix
// ---------------------------------------------------------------------------

func TestWalkPrefix_ReturnsOnlyKeysMatchingPrefix(t *testing.T) {
	tree := New[string]()
	tree.Insert("api/users", "u")
	tree.Insert("api/users/list", "ul")
	tree.Insert("api/orders", "o")
	tree.Insert("web/index", "wi")
	tree.Insert("web/about", "wa")

	var got []string
	tree.WalkPrefix("api/", func(key string, _ string) bool {
		got = append(got, key)
		return true
	})

	want := []string{"api/orders", "api/users", "api/users/list"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkPrefix(api/) = %v, want %v", got, want)
	}
}

func TestWalkPrefix_ReturnsExactMatchAndDescendants(t *testing.T) {
	tree := New[string]()
	tree.Insert("api/users", "u")
	tree.Insert("api/users/list", "ul")
	tree.Insert("api/orders", "o")

	var got []string
	tree.WalkPrefix("api/users", func(key string, _ string) bool {
		got = append(got, key)
		return true
	})

	want := []string{"api/users", "api/users/list"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkPrefix(api/users) = %v, want %v", got, want)
	}
}

func TestWalkPrefix_MatchesPartialEdge_WhenPrefixEndsInsideNodePrefix(t *testing.T) {
	tree := New[int]()
	tree.Insert("foobar", 1)
	tree.Insert("foobaz", 2)

	// "fooba" splits inside the edge prefix
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

func TestWalkPrefix_ReturnsNothing_WhenPrefixDoesNotMatch(t *testing.T) {
	tree := New[string]()
	tree.Insert("foo", "1")

	var got []string
	tree.WalkPrefix("bar", func(key string, _ string) bool {
		got = append(got, key)
		return true
	})

	if len(got) != 0 {
		t.Errorf("WalkPrefix(bar) should return nothing, got %v", got)
	}
}

func TestWalkPrefix_EmptyPrefix_ReturnsAllKeys(t *testing.T) {
	tree := New[string]()
	tree.Insert("a", "1")
	tree.Insert("b", "2")
	tree.Insert("c", "3")

	var got []string
	tree.WalkPrefix("", func(key string, _ string) bool {
		got = append(got, key)
		return true
	})

	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("WalkPrefix(\"\") = %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Keys
// ---------------------------------------------------------------------------

func TestKeys_ReturnsAllKeysInSortedOrder(t *testing.T) {
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

func TestKeys_ReturnsEmptySlice_OnEmptyTree(t *testing.T) {
	tree := New[string]()

	keys := tree.Keys()
	if len(keys) != 0 {
		t.Errorf("Keys() on empty tree = %v, want []", keys)
	}
}

// ---------------------------------------------------------------------------
// ToMap
// ---------------------------------------------------------------------------

func TestToMap_ReturnsAllKeyValuePairs(t *testing.T) {
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

func TestToMap_ReturnsEmptyMap_OnEmptyTree(t *testing.T) {
	tree := New[string]()

	m := tree.ToMap()
	if len(m) != 0 {
		t.Errorf("ToMap() on empty tree = %v, want {}", m)
	}
}

// ---------------------------------------------------------------------------
// Len
// ---------------------------------------------------------------------------

func TestLen_ReturnsZero_ForNewTree(t *testing.T) {
	tree := New[string]()

	if tree.Len() != 0 {
		t.Errorf("Len() = %d, want 0 for new tree", tree.Len())
	}
}

func TestLen_IncrementsOnInsert_DecrementsOnDelete(t *testing.T) {
	tree := New[string]()

	tree.Insert("a", "1")
	tree.Insert("b", "2")
	if tree.Len() != 2 {
		t.Errorf("Len() = %d after 2 inserts, want 2", tree.Len())
	}

	tree.Delete("a")
	if tree.Len() != 1 {
		t.Errorf("Len() = %d after 1 delete, want 1", tree.Len())
	}
}

func TestLen_DoesNotChange_WhenDeletingNonExistentKey(t *testing.T) {
	tree := New[string]()
	tree.Insert("a", "1")

	tree.Delete("zzz")
	if tree.Len() != 1 {
		t.Errorf("Len() = %d after deleting non-existent key, want 1", tree.Len())
	}
}

// ---------------------------------------------------------------------------
// Generic type support
// ---------------------------------------------------------------------------

func TestGenericTypes_IntValues_IncludingZero(t *testing.T) {
	tree := New[int]()
	tree.Insert("one", 1)
	tree.Insert("zero", 0) // zero value should be distinguishable from "not found"

	v, ok := tree.Get("one")
	if !ok || v != 1 {
		t.Errorf("Get(one) = (%d, %v), want (1, true)", v, ok)
	}

	v, ok = tree.Get("zero")
	if !ok || v != 0 {
		t.Errorf("Get(zero) = (%d, %v), want (0, true)", v, ok)
	}
}

func TestGenericTypes_StructValues(t *testing.T) {
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
}

func TestGenericTypes_PointerValues_IncludingNil(t *testing.T) {
	tree := New[*int]()
	n := 42
	tree.Insert("valid", &n)
	tree.Insert("null", nil) // nil pointer is a valid value

	v, ok := tree.Get("valid")
	if !ok || *v != 42 {
		t.Errorf("Get(valid) = (%v, %v), want (ptr→42, true)", v, ok)
	}

	v, ok = tree.Get("null")
	if !ok || v != nil {
		t.Errorf("Get(null) = (%v, %v), want (nil, true)", v, ok)
	}
}

// ---------------------------------------------------------------------------
// Scale
// ---------------------------------------------------------------------------

func TestScale_TenThousandInsertions_AllRetrievableAndSorted(t *testing.T) {
	tree := New[int]()
	const n = 10_000

	for i := 0; i < n; i++ {
		tree.Insert(fmt.Sprintf("key/%05d/value", i), i)
	}

	if tree.Len() != n {
		t.Fatalf("Len() = %d, want %d", tree.Len(), n)
	}

	// Spot-check retrieval.
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key/%05d/value", i)
		v, ok := tree.Get(key)
		if !ok || v != i {
			t.Fatalf("Get(%q) = (%d, %v), want (%d, true)", key, v, ok, i)
		}
	}

	// Keys must be lexicographically sorted.
	keys := tree.Keys()
	if !sort.StringsAreSorted(keys) {
		t.Fatal("Keys() returned unsorted results")
	}
}

func TestScale_TenThousandInsertions_AllDeletable(t *testing.T) {
	tree := New[int]()
	const n = 10_000

	for i := 0; i < n; i++ {
		tree.Insert(fmt.Sprintf("key/%05d/value", i), i)
	}

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("key/%05d/value", i)
		if _, ok := tree.Delete(key); !ok {
			t.Fatalf("Delete(%q) should have found the key", key)
		}
	}

	if tree.Len() != 0 {
		t.Fatalf("Len() = %d after deleting all, want 0", tree.Len())
	}
}
