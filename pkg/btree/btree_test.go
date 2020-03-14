package btree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var numberRunes = []rune("0123456789")

func init() {
	seed := time.Now().Unix()
	fmt.Printf("Testing random seed: %d\n", seed)
	rand.Seed(seed)
}

func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randKV(count, length int) map[string]string {
	kv := map[string]string{}
	for i := 0; i < count; i++ {
		key := randString(length)
		value := randString(length)
		kv[key] = value
	}
	return kv
}

func TestFreeList(t *testing.T) {
	tree := New(3)
	nodes := []*node{}
	size := defaultFreeListSize * 10
	f := tree.freeList
	if f.getSize() != 0 {
		t.Fatalf("expected freeList size: 0, get: %d", f.getSize())
	}
	for i := 0; i < size; i++ {
		node := tree.newNode()
		nodes = append(nodes, node)
	}
	if f.getSize() != 0 {
		t.Fatalf("expected freeList size: 0, get: %d", f.getSize())
	}
	for _, node := range nodes {
		node.free()
	}
	if f.getSize() != defaultFreeListSize {
		t.Fatalf("expected freeList size: %d, get: %d", defaultFreeListSize, f.getSize())
	}
}

func TestBTree_set(t *testing.T) {
	tr := New(3)
	totalCase := 1000
	kv := randKV(totalCase, 10)
	caseID := 0
	inOrder := true
	for k, v := range kv {
		fmt.Printf("\nSet/Get test (%d/%d), key=%s\n", caseID, totalCase, k)
		caseID++
		var lastIt Item
		key := []byte(k)
		value := []byte(v)
		tr.Set(key, value)
		found, getValue := tr.Get(key)
		if !found {
			t.Fatalf("key=%s not found, nodes=%d", string(key), tr.length)
		}
		if string(getValue) != string(value) {
			t.Fatalf("key=%s, expected=%s, got=%s", string(key), string(value), string(getValue))
		}
		tr.Iterate(func(it Item, id int) {
			// fmt.Printf("%s(%d)\n", it, id)
			if lastIt != nil && it.lessThan(lastIt) {
				inOrder = false
			}
			lastIt = it
		})
		if !inOrder {
			t.Fatalf("Tree lose order")
		}
	}
	fmt.Println("PASS BTree Set/Get")
}

func TestBTree_delete(t *testing.T) {
	totalCase := 1000
	tr := New(3)
	kv := randKV(totalCase, 10)
	keys := [][]byte{}
	for k, v := range kv {
		key := []byte(k)
		value := []byte(v)
		tr.Set(key, value)
		keys = append(keys, key)
	}
	perm := rand.Perm(len(keys))
	for caseID, p := range perm {
		fmt.Printf("Delete test (%d/%d)\n", caseID, totalCase)
		key := keys[p]
		value := tr.Delete(key)
		if string(value) != string(kv[string(key)]) {
			t.Fatalf("Delete error: expected %s, get %s", kv[string(key)], value)
		}
		found, _ := tr.Get(key)
		if found {
			t.Fatalf("Item %s not deleted", key)
		}
	}
	fmt.Printf("PASS BTree Delete")
}
