package bt

import "sync"

type freeList struct {
	mu   *sync.Mutex
	list []*node
}

func newFreeList(size int) *freeList {
	return &freeList{
		mu: &sync.Mutex{},
		// Size won't change. Must set 0 or go would think list already has item
		list: make([]*node, 0, size),
	}
}

func (f *freeList) getSize() int {
	return len(f.list)
}
