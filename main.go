package main

import (
	"fmt"

	"github.com/daicang/bt/pkg/btree"
)

func main() {
	t := btree.New(3)
	t.Set([]byte("hello"), []byte("world"))
	val := t.Get([]byte("hello"))

	fmt.Printf("%v\n", string(val))
}
