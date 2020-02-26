// Thanks to following projects:
// btree(https://github.com/google/btree)

package btree

import (
	"bytes"
	"sort"
	"sync"
)

const (
	defaultFreeListSize = 32
	lesser              = -1
	eq                  = 0
	greater             = 1
)

// BTree is the in-memory indexing structure
type BTree struct {
	root     *node
	degree   int
	freeList *freeList
}

// Item holds key/value pair
type Item struct {
	key   []byte
	value []byte
}

type freeList struct {
	mu   sync.Mutex
	list []*node
}

// len(children) is 0 or len(inodes) + 1
type node struct {
	isLeaf   bool
	free     *freeList
	indexes  []*Item
	children []*node
}

func newFreeList(size int) *freeList {
	return &freeList{
		mu: sync.Mutex{},
		// Size won't change. Must set 0 or go would think list already has item
		list: make([]*node, 0, size),
	}
}

func (f *freeList) newNode() *node {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) == 0 {
		return &node{
			free:     f,
			indexes:  []*Item{},
			children: []*node{},
		}
	}
	n := f.list[len(f.list)-1]
	f.list[len(f.list)-1] = nil
	f.list = f.list[:len(f.list)-1]
	return n
}

func (f *freeList) freeNode(n *node) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) < cap(f.list) {
		f.list = append(f.list, n)
	}
}

func NewItem(key, value []byte) *Item {
	return &Item{
		key:   key,
		value: value,
	}
}

// compare returns -1 when it < other, 0 when equal, 1 when greater
func (it *Item) compare(other *Item) int {
	return bytes.Compare(it.key, other.key)
}

func New(degree int) *BTree {
	return &BTree{
		degree:   degree,
		freeList: newFreeList(defaultFreeListSize),
	}
}

func (b *BTree) Get(key []byte) []byte {
	if b.root == nil {
		return nil
	}
	it := NewItem(key, nil)
	b.root.get(it)
	return it.value
}

func (b *BTree) Set(key, value []byte) {
	it := NewItem(key, value)
	if b.root == nil {
		b.root = b.freeList.newNode()
		b.root.insertIndexAt(0, it)
		return
	}
	if len(b.root.indexes) > b.degree {
		// Split root and place new root
		i := b.degree / 2
		r := b.freeList.newNode()
		it, second := b.root.split(i)
		r.indexes = append(r.indexes, it)
		r.children = append(r.children, b.root)
		r.children = append(r.children, second)
		b.root = r
	}
	b.root.set(it, b.degree)
}

// get gets the item, return false if not found
func (n *node) get(it *Item) bool {
	i, found := n.searchIndex(it)
	if found {
		it.value = n.indexes[i].value
		return true
	}
	if len(n.children) == 0 {
		return false
	}
	return n.children[i].get(it)
}

// set sets key-value in subtree, return old value
func (n *node) set(it *Item, degree int) (old *Item) {
	i, found := n.searchIndex(it)
	if found {
		// Item already in index, rewrite
		old = n.indexes[i]
		n.indexes[i] = it
		return
	}
	// When item not in index, and is leaf node, add to index
	if len(n.children) == 0 {
		old = nil
		n.insertIndexAt(i, it)
		return
	}
	// Find a child to set
	// FIXME: what if child[i] not exists?
	if n.maybeSplitChild(i, degree) {
		switch it.compare(n.indexes[i]) {
		case lesser:
		case greater:
			// Search in the new child
			i++
		case eq:
			old = n.indexes[i]
			n.indexes[i] = it
			return
		}
	}
	return n.children[i].set(it, degree)
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
func (n *node) split(i int) (*Item, *node) {
	new := n.free.newNode()
	new.indexes = n.indexes[i+1:]
	item := n.indexes[i]
	n.indexes = n.indexes[:i]
	if len(n.children) > 0 {
		new.children = n.children[i+1:]
		n.children = n.children[:i+1]
	}
	return item, new
}

func (n *node) insertChildAt(i int, child *node) {
	n.children = append(n.children, &node{})
	copy(n.children[i:], n.children[i+1:])
	n.children[i] = child
}

// searchInode returns (firstGreaterEqIndex, found)
func (n *node) searchIndex(it *Item) (int, bool) {
	i := n.getFirstNonLessIndex(it)
	if i < len(n.indexes) && bytes.Compare(it.key, n.indexes[i].key) == 0 {
		return i, true
	}
	return i, false
}

func (n *node) getFirstNonLessIndex(it *Item) int {
	return sort.Search(len(n.indexes), func(i int) bool {
		return bytes.Compare(n.indexes[i].key, it.key) != -1
	})
}

// insertInode inserts kv on given node
func (n *node) insertIndex(it *Item) {
	index := sort.Search(len(n.indexes), func(i int) bool {
		return bytes.Compare(n.indexes[i].key, it.key) != -1
	})
	n.insertIndexAt(index, it)
}

// insertIndexAt inserts index at given position, pushing subsequent values
func (n *node) insertIndexAt(i int, it *Item) {
	n.indexes = append(n.indexes, &Item{})
	copy(n.indexes[i+1:], n.indexes[i:])
	n.indexes[i] = it
}

// removeItemAt removes item at given index
func (n *node) removeInodeAt(i int) {
	copy(n.indexes[i:], n.indexes[i+1:])
	n.indexes = n.indexes[:len(n.indexes)-1]
}
