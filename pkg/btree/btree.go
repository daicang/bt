// Thanks to following projects:
// btree(https://github.com/google/btree)
//
// TODO:
// - Add free list

package btree

import (
	"bytes"
	"sort"
)

type BTree struct {
	root   *node
	degree int
}

// Item
type Item struct {
	key   []byte
	value []byte
}

func NewBTree(degree int) *BTree {
	return &BTree{
		root:   nil,
		degree: degree,
	}
}

func (b *BTree) Get(it *Item) bool {
	if b.root == nil {
		return false
	}
	return b.root.get(it)
}

func (b *BTree) Set(it *Item) {

}

// len(children) == 0 or
// len(children) = len(inodes) + 1
type node struct {
	isLeaf   bool
	indexes  []*Item
	children []*node
}

func NewNode() *node {
	return &node{}
}

// get gets the item, return false if not found
func (n *node) get(it *Item) bool {
	i, found := n.searchIndex(it)
	if found {
		it.value = n.indexes[i].value
		return true
	}
	if n.isLeaf {
		return false
	}
	return n.children[i].get(it)
}

// set sets key-value in subtree
func (n *node) set(it *Item, degree int) {
	i, found := n.searchIndex(it)
	if found {
		// Set found index
		n.indexes[i] = it
		return
	}
	// Add to index
	if n.isLeaf {
		n.insertIndexAt(i, it)
		return
	}
	// Add to child
	if n.maybeSplitChild(i, degree) {

	} else {

	}
}

// maybeSplitChild returns whether i-th child should be splitted,
// if so, split the child
func (n *node) maybeSplitChild(i, degree int) bool {
	child := n.children[i]
	if len(child.indexes) < degree {
		return false
	}
	it, newChild := child.split(i)
	// Split i-th child, child-i < inode-i => child-i < new-inode < new-child < inode-i
	n.insertIndexAt(i, it)
	n.insertChildAt(i+1, newChild)
	return true
}

// split Splits node at given index, return element at index and new node
func (n *node) split(i int) (inode, *node) {
	new := NewNode()
	new.inodes = n.inodes[i+1:]
	item := n.inodes[i]
	n.inodes = n.inodes[:i]
	if !n.isLeaf {
		new.children = n.children[i+1:]
		n.children = n.children[:i+1]
	}
	return item, new
}

func (n *node) insertChildAt(i int, child *node) {
	n.children = append(n.children, node{})
	copy(n.children[i:], n.children[i+1:])
	n.children[i] = child
}

// searchInode returns index, found
func (n *node) searchIndex(it *Item) (int, bool) {
	index := n.getFirstNonLessIndex(it)
	if bytes.Compare(key, n.inodes[index].key) == 0 {
		return index, true
	}
	return index, false
}

func (n *node) getFirstNonLessIndex(it *Item) int {
	sort.Search(len(n.indexes), func(i int) bool {
		return bytes.Compare(n.indexes[i].key, it.key) != -1
	})
}

// insertInode inserts kv on given node
func (n *node) insertIndex(it Item) {
	index := sort.Search(len(n.inodes), func(i int) bool {
		return bytes.Compare(n.inodes[i].key, item.key) != -1
	})
	n.insertInodeAt(index, item)
}

// insertIndexAt inserts index at given position, pushing subsequent values
func (n *node) insertIndexAt(i int, it *Item) {
	n.inodes = append(n.inodes, inode{})
	copy(n.inodes[i+1:], n.inodes[i:])
	n.inodes[i] = it
}

// removeItemAt removes item at given index
func (n *node) removeInodeAt(index int) {
	copy(n.inodes[index:], n.inodes[index+1:])
	n.inodes = n.inodes[:len(n.inodes)-1]
}
