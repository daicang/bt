package bt

import (
	"fmt"
)

const (
	freelistSize = 32
)

// BTree is in-memory b-tree index
type BTree struct {
	// Pointer to root node
	root *node

	degree int

	// maxItem returns the maxium items in a node
	maxItem int

	// minItem returns mininum items in a node, except the root node
	minItem int

	// nodeCount represents number of nodes in b-tree
	nodeCount int

	// freeList caches nodes
	freeList *freeList
}

// New returns a btree with given degree
func New(degree int) *BTree {
	return &BTree{
		degree:    degree,
		nodeCount: 0,
		maxItem:   degree*2 - 1,
		minItem:   degree - 1,
		freeList:  newFreeList(freelistSize),
	}
}

// Get returns (found, value) for given key
func (t *BTree) Get(key KeyType) (bool, ValueType) {
	if t.root == nil {
		return false, nil
	}
	return t.root.get(key)
}

// Set insert or replace an item in btree, returns (found, oldValue)
func (t *BTree) Set(key KeyType, value ValueType) (bool, ValueType) {
	if t.root == nil {
		t.root = t.newNode()
		p := newPair(key, value)
		t.root.inodes.insertAt(0, p)
		return false, nil
	}
	if len(t.root.inodes) >= t.maxItem {
		// Split root and place new root
		oldRoot := t.root
		t.root = t.newNode()
		index, second := oldRoot.split(t.maxItem / 2)
		t.root.inodes = append(t.root.inodes, index)
		t.root.children = append(t.root.children, oldRoot)
		t.root.children = append(t.root.children, second)
		fmt.Printf("Split root\n")
	}
	return t.root.set(key, value, t.maxItem)
}

// Delete removes an item, returns (found, oldValue)
func (t *BTree) Delete(key KeyType) (bool, ValueType) {
	if t.root == nil {
		return false, nil
	}
	found, oldValue := t.root.remove(key, t.minItem)
	if len(t.root.inodes) == 0 && len(t.root.children) > 0 {
		emptyroot := t.root
		t.root = t.root.children[0]
		emptyroot.free()
	}
	return found, oldValue
}

// newNode allocates a new node for b-tree
func (t *BTree) newNode() *node {
	t.freeList.mu.Lock()
	defer t.freeList.mu.Unlock()
	t.nodeCount++
	if len(t.freeList.list) == 0 {
		return &node{
			id:   t.nodeCount,
			tree: t,
		}
	}
	index := len(t.freeList.list) - 1
	n := t.freeList.list[index]
	t.freeList.list[index] = nil
	t.freeList.list = t.freeList.list[:index]
	return n
}

// Iterate iterates the tree
func (t *BTree) Iterate(iter iterator) {
	t.root.iterate(iter)
}
