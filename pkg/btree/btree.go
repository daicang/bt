// Thanks to following projects:
// btree(https://github.com/google/btree)

package btree

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
)

const (
	defaultFreeListSize = 32
	lesser              = -1
	eq                  = 0
	greater             = 1
)

// Item holds key/value pair
type Item struct {
	key   []byte
	value []byte
}

func (it *Item) lessThan(other *Item) bool {
	return bytes.Compare(it.key, other.key) == -1
}

func (it *Item) equalTo(other *Item) bool {
	return bytes.Equal(it.key, other.key)
}

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

func (t *BTree) newNode() *node {
	t.freeList.mu.Lock()
	defer t.freeList.mu.Unlock()
	t.length++
	if len(t.freeList.list) == 0 {
		return &node{
			tree: t,
		}
	}
	index := len(t.freeList.list) - 1
	n := t.freeList.list[index]
	t.freeList.list[index] = nil
	t.freeList.list = t.freeList.list[:index]
	return n
}

func (n *node) free() {
	f := n.tree.freeList
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) < cap(f.list) {
		n.inode = []*Item{}
		n.children = []*node{}
		f.list = append(f.list, n)
	}
}

// NewItem creates a new item
func NewItem(key, value []byte) *Item {
	return &Item{
		key:   key,
		value: value,
	}
}

// BTree is the in-memory indexing structure
type BTree struct {
	root     *node
	degree   int
	length   int
	freeList *freeList
}

type inode []*Item

func (in inode) check() {
	var last *Item
	for _, curr := range in {
		if last != nil {
			if bytes.Compare(last.key, curr.key) != -1 {
				for _, p := range in {
					fmt.Printf(" %s", p.key)
				}
				fmt.Printf("\n%s !< %s, %d\n", last.key, curr.key, bytes.Compare(last.key, curr.key))
				panic("inode order error")
			}
		}
		last = curr
	}
}

// search returns (found, firstGreaterEqIndex)
func (in inode) search(it *Item) (bool, int) {
	in.check()
	i := sort.Search(len(in), func(i int) bool {
		// Return index of first not-less-than item
		return !in[i].lessThan(it)
	})
	if i < len(in) && it.equalTo(in[i]) {
		return true, i
	}
	return false, i
}

// set inserts or replaces item into inode
func (in *inode) set(it *Item) {
	found, i := in.search(it)
	if found {
		(*in)[i] = it
	} else {
		in.insert(i, it)
	}
	in.check()
}

// insert inserts it at given position, pushing subsequent values
func (in *inode) insert(i int, it *Item) {
	if i > 0 && bytes.Compare(it.key, (*in)[i-1].key) < 0 {
		fmt.Printf("1: %d %s %s\n", i, (*in)[i-1].key, it.key)
		panic("insert: left inconsistent")
	}
	if i < len(*in) && bytes.Compare(it.key, (*in)[i].key) > 0 {
		fmt.Printf("2: %d %s %s\n", i, it.key, (*in)[i].key)
		panic("insert: right inconsistent")
	}
	in.check()
	*in = append((*in), &Item{})
	copy((*in)[i+1:], (*in)[i:])
	(*in)[i] = it
	in.check()
}

// remove removes item at given index
func (in *inode) remove(i int) {
	copy((*in)[i:], (*in)[i+1:])
	(*in) = (*in)[:len(*in)-1]
}

// len(children) is 0 or len(inodes) + 1
type node struct {
	tree     *BTree
	inode    inode
	children []*node
}

// New returns a BTree with given degree
func New(degree int) *BTree {
	return &BTree{
		degree:   degree,
		length:   0,
		freeList: newFreeList(defaultFreeListSize),
	}
}

// maxItem returns the maxium items in a node
func (t *BTree) maxItem() int {
	return t.degree*2 - 1
}

// minItem returns mininum items in a node, except the root node
func (t *BTree) minItem() int {
	return t.degree - 1
}

// Get returns (found, item) in btree
func (t *BTree) Get(key []byte) (bool, []byte) {
	fmt.Printf("Get ")
	if t.root == nil {
		return false, nil
	}
	it := NewItem(key, nil)
	found := t.root.get(it)
	return found, it.value
}

// Set insert or replace an item in btree
func (t *BTree) Set(key, value []byte) {
	fmt.Printf("Set ")
	it := NewItem(key, value)
	if t.root == nil {
		t.root = t.newNode()
		t.root.inode.insert(0, it)
		return
	}
	if len(t.root.inode) >= t.maxItem() {
		// Split root and place new root
		i := t.maxItem() / 2
		newRoot := t.newNode()
		it, second := t.root.split(i)
		newRoot.inode = append(newRoot.inode, it)
		newRoot.children = append(newRoot.children, t.root)
		newRoot.children = append(newRoot.children, second)
		t.root = newRoot
	}
	t.root.set(it, t.maxItem())
}

// Delete removes an item in btree
func (t *BTree) Delete(key []byte) bool {
	return true
}

// get gets the item, return false if not found
func (n *node) get(it *Item) bool {
	found, i := n.inode.search(it)
	if found {
		it.value = n.inode[i].value
		fmt.Printf("\n")
		return true
	}
	if len(n.children) == 0 {
		// We have reached leaf
		fmt.Printf("\n")
		return false
	}
	fmt.Printf("%d", i)
	return n.children[i].get(it)
}

// set sets key-value in subtree, return old value
func (n *node) set(it *Item, maxItem int) (old *Item) {
	found, i := n.inode.search(it)
	if found {
		// Item already in index, rewrite
		fmt.Println("rewrite index")
		old = n.inode[i]
		n.inode[i] = it
		return
	}
	// Leaf node, add to index
	if len(n.children) == 0 {
		old = nil
		fmt.Printf(" insert leaf at %d\n", i)
		n.inode.insert(i, it)
		n.inode.check()
		return
	}
	// Check whether child i need split
	child := n.children[i]
	if len(child.inode) > maxItem {
		// fmt.Printf("Split child %d at %d/%d\n", i, maxItem/2, len(child.inode))
		for _, ii := range child.inode {
			if i > 0 && i < len(n.inode) && bytes.Compare(ii.key, n.inode[i].key) <= 0 {
				fmt.Printf("\n%d %s %s\n", i, n.inode[i].key, ii.key)
				panic("left inconsistent")
			}
			if i < len(n.inode)-1 && bytes.Compare(ii.key, n.inode[i+1].key) > 0 {
				panic("right inconsistent")
			}
		}

		newIndex, newChild := child.split(maxItem / 2)
		// FIXME: WRONG index!!!!
		n.inode.insert(i, newIndex)
		n.insertChildAt(i+1, newChild)
		if it.equalTo(newIndex) {
			old = n.inode[i]
			n.inode[i] = it
			return
		}
		if newIndex.lessThan(it) {
			i++
		}
	}
	fmt.Printf("%d", i)
	return n.children[i].set(it, maxItem)
}

// split Splits node at given index, return element at index and new node
func (n *node) split(i int) (*Item, *node) {
	new := n.tree.newNode()
	new.inode = n.inode[i+1:]
	item := n.inode[i]
	n.inode = n.inode[:i]
	if len(n.children) > 0 {
		new.children = n.children[i+1:]
		n.children = n.children[:i+1]
	}
	return item, new
}

func (n *node) insertChildAt(i int, child *node) {
	n.children = append(n.children, &node{})
	copy(n.children[i+1:], n.children[i:])
	n.children[i] = child
}
