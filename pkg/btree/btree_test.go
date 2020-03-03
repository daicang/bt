package btree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
	kv := randKV(1000, 20)
	count := 0
	for k, v := range kv {
		key := []byte(k)
		value := []byte(v)
		tr.Set(key, value)
		found, getValue := tr.Get(key)
		if !found {
			t.Fatalf("key=%s not found, count=%d, nodes=%d", string(key), count, tr.length)
		}
		if string(getValue) != string(value) {
			t.Fatalf("key=%s, expected=%s, got=%s", string(key), string(value), string(getValue))
		}
		count++
	}
	fmt.Println("PASS BTree set")
}
