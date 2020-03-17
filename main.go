package main

import (
	"fmt"
	"os"

	"github.com/daicang/bt/pkg/btree"
)

type db struct {
	path string
	file *os.File
}

func main() {
	t := btree.New(3)
	t.Set([]byte("hello"), []byte("world"))
	val := t.Get([]byte("hello"))

	fmt.Printf("%v\n", string(val))
}
