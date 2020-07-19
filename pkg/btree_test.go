package bt

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

// randString generates string for given length
func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randStringMap(count, length int) map[string]string {
	m := map[string]string{}
	for i := 0; i < count; i++ {
		key := randString(length)
		value := randString(length)
		m[key] = value
	}
	return m
}

func TestBTree_set(t *testing.T) {
	tr := New(3)
	totalCase := 1000
	kv := randStringMap(totalCase, 10)
	caseID := 0
	inOrder := true
	for k, v := range kv {
		fmt.Printf("\nSet/Get test (%d/%d), key=%s\n", caseID, totalCase, k)
		caseID++
		var lastIt pair
		key := []byte(k)
		value := []byte(v)
		tr.Set(key, value)
		found, getValue := tr.Get(key)
		if !found {
			t.Fatalf("key=%s not found, nodes=%d", string(key), tr.nodeCount)
		}
		if string(getValue) != string(value) {
			t.Fatalf("key=%s, expected=%s, got=%s", string(key), string(value), string(getValue))
		}
		tr.Iterate(func(it pair) {
			// fmt.Printf("%s(%d)\n", it, id)
			if len(lastIt.key) > 0 && it.key.lessThan(lastIt.key) {
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
	kv := randStringMap(totalCase, 10)
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
		found, value := tr.Delete(key)
		if found == false || string(value) != string(kv[string(key)]) {
			t.Fatalf("Expected %s, get %s", kv[string(key)], value)
		}
		found, _ = tr.Get(key)
		if found {
			t.Fatalf("Key %s not deleted", key)
		}
	}
	fmt.Printf("PASS BTree Delete")
}
