package btree

import (
	"fmt"
)

// BTree is the in-memory indexing structure
type BTree struct {
	root     *node
	degree   int
	length   int
	freeList *freeList
}

// New returns a BTree with given degree
func New(degree int) *BTree {
	return &BTree{
		degree:   degree,
		length:   0,
		freeList: newFreeList(defaultFreeListSize),
	}
}

func (t *BTree) newNode() *node {
	t.freeList.mu.Lock()
	defer t.freeList.mu.Unlock()
	t.length++
	if len(t.freeList.list) == 0 {
		return &node{
			id:   t.length,
			tree: t,
		}
	}
	index := len(t.freeList.list) - 1
	n := t.freeList.list[index]
	t.freeList.list[index] = nil
	t.freeList.list = t.freeList.list[:index]
	return n
}

// maxItem returns the maxium items in a node
func (t *BTree) maxItem() int {
	return t.degree*2 - 1
}

// minItem returns mininum items in a node, except the root node
func (t *BTree) minItem() int {
	return t.degree - 1
}

// Iterate iterates the tree
func (t *BTree) Iterate(iter iterator) {
	t.root.iterate(iter)
}

// Get returns (found, value) for given key
func (t *BTree) Get(key []byte) (bool, []byte) {
	fmt.Printf("Get %s\n", key)
	if t.root == nil {
		return false, nil
	}
	return t.root.get(key)
}

// Set insert or replace an item in btree
func (t *BTree) Set(key keyType, value []byte) {
	fmt.Printf("Set %s=%s\n", key, value)
	it := newKV(key, value)
	if t.root == nil {
		t.root = t.newNode()
		t.root.inodes.insertAt(0, it)
		return
	}
	if len(t.root.inodes) >= t.maxItem() {
		// Split root and place new root
		oldRoot := t.root
		t.root = t.newNode()
		index, second := oldRoot.split(t.maxItem() / 2)
		t.root.inodes = append(t.root.inodes, index)
		t.root.children = append(t.root.children, oldRoot)
		t.root.children = append(t.root.children, second)
		fmt.Printf("Split root\n")
	}
	t.root.set(key, value, t.maxItem())
}

// Delete removes an item, return value if exist
func (t *BTree) Delete(key keyType) []byte {
	if t.root == nil {
		return nil
	}
	removed := t.root.remove(key, t.minItem())
	if len(t.root.inodes) == 0 && len(t.root.children) > 0 {
		emptyroot := t.root
		t.root = t.root.children[0]
		emptyroot.free()
	}
	return removed.value
}
