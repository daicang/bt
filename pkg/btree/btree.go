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
)

// Item is the element type
type Item interface {
	lessThan(Item) bool
	equalTo(Item) bool
}

// KV implements Item
type KV struct {
	key   []byte
	value []byte
}

// newKV creates a new KV
func newKV(key, value []byte) Item {
	return KV{
		key:   key,
		value: value,
	}
}

func (kv KV) lessThan(other Item) bool {
	otherKV, ok := other.(KV)
	if !ok {
		return false
	}
	return bytes.Compare(kv.key, otherKV.key) == -1
}

func (kv KV) equalTo(other Item) bool {
	otherKV, ok := other.(KV)
	if !ok {
		return false
	}
	return bytes.Equal(kv.key, otherKV.key)
}

func (kv KV) String() string {
	return string(kv.key)
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

func (n *node) free() {
	f := n.tree.freeList
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) < cap(f.list) {
		n.inode = []Item{}
		n.children = []*node{}
		f.list = append(f.list, n)
	}
}

// BTree is the in-memory indexing structure
type BTree struct {
	root     *node
	degree   int
	length   int
	freeList *freeList
}

type inode []Item

type children []*node

// len(children) is 0 or len(inodes) + 1
type node struct {
	id       int
	tree     *BTree
	inode    inode
	children children
}

func (in inode) check() {
	var last Item
	for _, curr := range in {
		if last != nil {
			if curr.lessThan(last) {
				for _, p := range in {
					fmt.Printf(" %s", p)
				}
				fmt.Printf("\n%s !< %s\n", last, curr)
				panic("inode order error")
			}
		}
		last = curr
	}
}

// search returns (found, firstGreaterEqIndex)
func (in inode) search(it Item) (bool, int) {
	in.check()
	i := sort.Search(len(in), func(i int) bool {
		// Return index of first item greater then it
		return it.lessThan(in[i])
	})
	if i > 0 && it.equalTo(in[i-1]) {
		return true, i - 1
	}
	return false, i
}

// set inserts or replaces item into inode
func (in *inode) set(it Item) {
	found, i := in.search(it)
	if found {
		(*in)[i] = it
	} else {
		in.insertAt(i, it)
	}
	in.check()
}

// insert inserts it at given position, pushing subsequent values
func (in *inode) insertAt(i int, it Item) {
	in.check()

	if i > 0 && it.lessThan((*in)[i-1]) {
		fmt.Printf("1: %d %s %s\n", i, (*in)[i-1], it)
		panic("insert: left inconsistent")
	}
	if i < len(*in) && (*in)[i].lessThan(it) {
		fmt.Printf("2: %d %s %s\n", i, it, (*in)[i])
		panic("insert: right inconsistent")
	}
	*in = append((*in), nil)
	copy((*in)[i+1:], (*in)[i:])
	(*in)[i] = it
	in.check()
}

// remove removes item at given index
func (in *inode) removeAt(i int) Item {
	it := (*in)[i]
	copy((*in)[i:], (*in)[i+1:])
	(*in)[len(*in)-1] = nil
	(*in) = (*in)[:len(*in)-1]
	return it
}

func (in *inode) pop() Item {
	it := (*in)[len(*in)-1]
	(*in) = (*in)[:len(*in)-1]
	return it
}

func (c *children) insertAt(i int, child *node) {
	*c = append(*c, nil)
	copy((*c)[i+1:], (*c)[i:])
	(*c)[i] = child
}

func (c *children) removeAt(i int) *node {
	n := (*c)[i]
	copy((*c)[i:], (*c)[i+1:])
	(*c)[len(*c)-1] = nil
	*c = (*c)[:len(*c)-1]
	return n
}

func (c *children) pop() *node {
	n := (*c)[len(*c)-1]
	(*c) = (*c)[:len(*c)-1]
	return n
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

// Iterate iterates the tree
func (t *BTree) Iterate(iter iterator) {
	t.root.iterate(iter)
}

// Get returns (found, item) in btree
func (t *BTree) Get(key []byte) (bool, []byte) {
	fmt.Printf("Get %s\n", key)
	if t.root == nil {
		return false, nil
	}
	it := newKV(key, nil)
	found, result := t.root.get(it)
	if found {
		resultKV, _ := result.(KV)
		return true, resultKV.value
	}
	return false, nil
}

// Set insert or replace an item in btree
func (t *BTree) Set(key, value []byte) {
	fmt.Printf("Set %s\n", key)
	it := newKV(key, value)
	if t.root == nil {
		t.root = t.newNode()
		t.root.inode.insertAt(0, it)
		return
	}
	if len(t.root.inode) >= t.maxItem() {
		// Split root and place new root
		oldRoot := t.root
		t.root = t.newNode()
		index, second := oldRoot.split(t.maxItem() / 2)
		t.root.inode = append(t.root.inode, index)
		t.root.children = append(t.root.children, oldRoot)
		t.root.children = append(t.root.children, second)
		fmt.Printf("Split root\n")
	}
	t.root.set(it, t.maxItem())
}

// Delete removes an item, return value if exist
func (t *BTree) Delete(key []byte) []byte {
	if t.root == nil {
		return nil
	}
	it := t.root.remove(newKV(key, nil), t.minItem(), removeItem)
	kv, ok := it.(KV)
	if !ok {
		panic("Invalid type")
	}
	return kv.value
}

// get returns (found, Item)
func (n *node) get(it Item) (bool, Item) {
	found, i := n.inode.search(it)
	if found {
		return true, n.inode[i]
	}
	if len(n.children) == 0 {
		// We have reached leaf
		return false, nil
	}
	n.checkChildInodes(i)
	return n.children[i].get(it)
}

func (n *node) checkChildInodes(i int) {
	// inode[i-1] < all inode in child[i] < inode[i]
	// fmt.Printf("Checking node %d\n", n.id)
	child := n.children[i]
	for _, childInode := range child.inode {
		if i > 0 && childInode.lessThan(n.inode[i-1]) {
			fmt.Printf("error: child inode=%s inode[i-1]=%s\n", childInode, n.inode[i-1])
			panic("Split: new inode error 1")
		}
		if i < len(n.inode) && n.inode[i].lessThan(childInode) {
			fmt.Printf("error: child inode=%s inode[i]=%s\n", childInode, n.inode[i])
			panic("Split: new inode error 2")
		}
	}
}

// set sets key-value in subtree, return old value
func (n *node) set(it Item, maxItem int) Item {
	found, i := n.inode.search(it)
	if len(n.children) > 0 {
		n.checkChildInodes(i)
	}
	if found {
		// Item already in index, rewrite
		// fmt.Println("rewrite index")
		old := n.inode[i]
		n.inode[i] = it
		return old
	}
	// Leaf node, add to index
	if len(n.children) == 0 {
		// fmt.Printf(" insert leaf at node=%d index=%d\n", n.id, i)
		n.inode.insertAt(i, it)
		return nil
	}
	n.inode.check()
	// Check whether child i need split
	child := n.children[i]

	if len(child.inode) > maxItem {
		fmt.Printf("Split child %d at %d/%d\n", i, maxItem/2, len(child.inode))

		newIndex, newChild := child.split(maxItem / 2)

		// child.inode.check()
		// newChild.inode.check()
		// n.inode.check()

		n.inode.insertAt(i, newIndex)

		n.children.insertAt(i+1, newChild)
		if it.equalTo(newIndex) {
			old := n.inode[i]
			n.inode[i] = it
			return old
		}
		if newIndex.lessThan(it) {
			i++
		}
	}
	// n.checkChildInodes(i)
	// fmt.Printf("(%d) Insert into child node %d\n", n.id, n.children[i].id)
	ret := n.children[i].set(it, maxItem)
	n.checkChildInodes(i)
	return ret
}

type iterator func(Item, int)

func (n *node) iterate(iter iterator) {
	hasChild := false
	if len(n.children) > 0 {
		hasChild = true
	}
	for index := 0; index < len(n.inode); index++ {
		if hasChild {
			n.children[index].iterate(iter)
		}
		iter(n.inode[index], n.id)
	}
	if hasChild {
		n.children[len(n.inode)].iterate(iter)
	}
}

type toRemove int

const (
	removeItem toRemove = iota
	removeMin
	removeMax
)

// remove removes specified or minium/maxium item from current node,
// returns the removed item.
func (n *node) remove(it Item, minItem int, typ toRemove) Item {
	if len(n.children) == 0 {
		switch typ {
		case removeItem:
			found, i := n.inode.search(it)
			if found {
				return n.inode.removeAt(i)
			}
			return nil
		case removeMin:
			return n.inode.removeAt(0)
		case removeMax:
			return n.inode.pop()
		default:
			panic("Invalid remove type")
		}
	}
	// Now node must have child. All 3 types would invoke remove on child
	var i int
	switch typ {
	case removeItem:
		found, i := n.inode.search(it)
		if found {
			removed := n.inode[i]
			// TODO: growChildAndRemove
			n.inode[i] = n.children[i].remove(it, minItem, removeMax)
			return removed
		}
		return n.children[i].remove(it, minItem, removeItem)
	case removeMin:
		i = 0
	case removeMax:
		i = len(n.children) - 1
	default:
		panic("Invalid remove type")
	}
	if len(n.children[i].inode) <= minItem {
		return n.growChildAndRemove(it, i, minItem, typ)
	}
	return n.children[i].remove(it, minItem, typ)
}

func (n *node) growChildAndRemove(it Item, i, minItem int, typ toRemove) Item {
	if i > 0 && len(n.children[i-1].inode) > minItem {
		// Borrow from left sibling
		n.children[i].inode.insertAt(0, n.inode[i-1])
		n.inode[i-1] = n.children[i-1].inode.pop()
		if len(n.children[i-1].children) > 0 {
			n.children[i].children.insertAt(0, n.children[i-1].children.pop())
		}
	} else if i < len(n.children)-1 && len(n.children[i+1].inode) > minItem {
		// Borrow from right sibling
		n.children[i].inode = append(n.children[i].inode, n.inode[i])
		n.inode[i] = n.children[i+1].inode.removeAt(0)
		if len(n.children[i+1].children) > 0 {
			n.children[i].children = append(n.children[i].children, n.children[i+1].children.removeAt(0))
		}
	} else {
		if i >= len(n.children)-1 {
			// i is the rightmost child
			i--
		}
		// Merge with right child
		left := n.children[i]
		right := n.children[i+1]
		left.inode = append(left.inode, n.inode.removeAt(i))
		left.inode = append(left.inode, right.inode...)
		left.children = append(left.children, right.children...)
		right.free()
	}
	return n.remove(it, minItem, typ)
}

// split Splits node at given index, return element at index and new node
func (n *node) split(i int) (Item, *node) {
	n.inode.check()
	new := n.tree.newNode()
	item := n.inode[i]
	// Never use direct slice slicing!! Causes very wired bug
	new.inode = append(new.inode, n.inode[i+1:]...)
	n.inode = n.inode[:i]
	if len(n.children) > 0 {
		new.children = append(new.children, n.children[i+1:]...)
		n.children = n.children[:i+1]
	}
	return item, new
}
