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
	fmt.Println(seed)
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

func TestBTree_set(t *testing.T) {
	tr := New(3)

	kv := randKV(1000, 20)

	for k, v := range kv {
		key := []byte(k)
		value := []byte(v)
		tr.Set(key, value)
		if string(tr.Get(key)) != string(value) {
			t.Fatalf("key=%s, expected=%s, got=%s", string(key), string(value), string(tr.Get(key)))
		}
	}
	fmt.Println("PASS BTree set")
}
