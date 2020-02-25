package btree

import (
	"fmt"
	"testing"
)

func TestBTree_set(t *testing.T) {
	tr := New(3)

	data := []*Item{
		NewItem([]byte("hello"), []byte("world")),
		NewItem([]byte("foo"), []byte("bar")),
		NewItem([]byte("bar"), []byte("bar")),
		NewItem([]byte("foo"), []byte("foo")),
		NewItem([]byte("hello"), []byte("btree")),
	}

	for _, it := range data {
		tr.Set(it.key, it.value)
		if string(tr.Get(it.key)) != string(it.value) {
			t.Fatalf("key=%s, expected=%s, got=%s", string(it.key), string(it.value), string(tr.Get(it.key)))
		}
	}
	fmt.Println("PASS BTree set")
}
