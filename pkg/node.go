package bt

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type inode []pair
type children []*node

// len(children) is 0 or len(inodes) + 1
type node struct {
	tree     *BTree
	id       int
	inodes   inode
	children children
}

func (n *node) free() {
	f := n.tree.freeList
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) < cap(f.list) {
		n.inodes = []pair{}
		n.children = []*node{}
		f.list = append(f.list, n)
	}
}

func (in inode) check() {
	var last pair
	for _, curr := range in {
		if len(last.key) > 0 {
			if curr.key.lessThan(last.key) {
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
func (in inode) search(key KeyType) (bool, int) {
	in.check()
	i := sort.Search(len(in), func(i int) bool {
		// Return index of first item greater than it
		return key.lessThan(in[i].key)
	})
	if i > 0 && key.equalTo(in[i-1].key) {
		return true, i - 1
	}
	return false, i
}

// set inserts or replaces item into inode
func (in *inode) set(p pair) {
	found, i := in.search(p.key)
	if found {
		(*in)[i] = p
	} else {
		in.insertAt(i, p)
	}
	// in.check()
}

// insert inserts it at given position, pushing subsequent values
func (in *inode) insertAt(i int, p pair) {
	*in = append((*in), pair{})
	copy((*in)[i+1:], (*in)[i:])
	(*in)[i] = p
	// in.check()
}

// remove removes and return kv at given index
func (in *inode) removeAt(i int) pair {
	it := (*in)[i]
	copy((*in)[i:], (*in)[i+1:])
	(*in)[len(*in)-1] = pair{}
	(*in) = (*in)[:len(*in)-1]
	return it
}

func (in *inode) pop() pair {
	it := (*in)[len(*in)-1]
	(*in)[len(*in)-1] = pair{}
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
	(*c)[len(*c)-1] = nil
	(*c) = (*c)[:len(*c)-1]
	return n
}

// get returns (found, value) from given key
func (n *node) get(key KeyType) (bool, []byte) {
	found, i := n.inodes.search(key)
	if found {
		return true, n.inodes[i].value
	}
	if len(n.children) == 0 {
		// Leaf
		return false, nil
	}
	n.checkChildInodes(i)
	return n.children[i].get(key)
}

// check checks 1. inode is sorted 2. for internal nodes, len(inode) = len(children) - 1
func (n *node) check() {
	n.inodes.check()
	if len(n.children) > 0 {
		if len(n.inodes) != len(n.children)-1 {
			panic(fmt.Sprintf("node %d has %d inode, %d children\n", n.id, len(n.inodes), len(n.children)))
		}
	}
}

func (n *node) checkChildInodes(i int) {
	// inode[i-1] < all inode in child[i] < inode[i]
	// fmt.Printf("Checking node %d\n", n.id)
	child := n.children[i]
	for _, childInode := range child.inodes {
		if i > 0 && childInode.key.lessThan(n.inodes[i-1].key) {
			fmt.Printf("error: child key=%s < inode[i-1] key=%s\n", childInode.key, n.inodes[i-1].key)
			panic("Split: new inode error 1")
		}
		if i < len(n.inodes)-1 && n.inodes[i].key.lessThan(childInode.key) {
			fmt.Printf("error: child key=%s >= inode[i] key=%s\n", childInode.key, n.inodes[i].key)
			panic("Split: new inode error 2")
		}
	}
}

// set sets value to given key in subtree, return (found, oldValue)
func (n *node) set(key KeyType, value []byte, maxItem int) (bool, ValueType) {
	found, i := n.inodes.search(key)
	n.check()
	if len(n.children) > 0 {
		n.checkChildInodes(i)
	}
	if found {
		// Item already in index, rewrite
		// fmt.Println("rewrite index")
		oldValue := n.inodes[i].value
		n.inodes[i].value = value
		return true, oldValue
	}
	// Leaf node, add to index
	if len(n.children) == 0 {
		// fmt.Printf(" insert leaf at node=%d index=%d\n", n.id, i)
		n.inodes.insertAt(i, newPair(key, value))
		return false, ValueType{}
	}
	// n.inode.check()
	// Check whether child i need split
	child := n.children[i]

	if len(child.inodes) > maxItem {
		fmt.Printf("Split child %d at %d/%d\n", i, maxItem/2, len(child.inodes))

		newIndex, newChild := child.split(maxItem / 2)

		// child.inode.check()
		// newChild.inode.check()
		// n.inode.check()

		n.inodes.insertAt(i, newIndex)

		n.children.insertAt(i+1, newChild)
		if key.equalTo(newIndex.key) {
			oldValue := n.inodes[i].value
			n.inodes[i].value = value
			return true, oldValue
		}
		if newIndex.key.lessThan(key) {
			i++
		}
	}
	// n.checkChildInodes(i)
	// fmt.Printf("(%d) Insert into child node %d\n", n.id, n.children[i].id)
	return n.children[i].set(key, value, maxItem)
	// n.checkChildInodes(i)
}

// iterator function receives kv pair
type iterator func(pair)

func (n *node) iterate(iter iterator) {
	hasChild := false
	if len(n.children) > 0 {
		hasChild = true
	}
	for index := 0; index < len(n.inodes); index++ {
		if hasChild {
			n.children[index].iterate(iter)
		}
		iter(n.inodes[index])
	}
	if hasChild {
		n.children[len(n.inodes)].iterate(iter)
	}
}

func (n *node) removeMin(minItem int) pair {
	if len(n.children) == 0 {
		return n.inodes.removeAt(0)
	}
	if len(n.children[0].inodes) <= minItem {
		n.extendChild(0, minItem)
	}
	return n.children[0].removeMin(minItem)
}

func (n *node) removeMax(minItem int) pair {
	if len(n.children) == 0 {
		return n.inodes.pop()
	}
	if len(n.children[len(n.children)-1].inodes) <= minItem {
		n.extendChild(len(n.children)-1, minItem)
	}
	return n.children[len(n.children)-1].removeMax(minItem)
}

// remove removes given key from b-tree, returns (found, oldValue)
func (n *node) remove(key KeyType, minItem int) (bool, ValueType) {
	found, i := n.inodes.search(key)
	if len(n.children) == 0 {
		if found {
			return true, n.inodes.removeAt(i).value
		}
		return false, ValueType{}
	}
	// Must check child before removeMax
	if len(n.children[i].inodes) <= minItem {
		n.extendChild(i, minItem)
		found, i = n.inodes.search(key)
	}
	if found {
		removed := n.inodes[i].value
		n.inodes[i] = n.children[i].removeMax(minItem)
		return true, removed
	}
	// n.checkChildInodes(i)
	return n.children[i].remove(key, minItem)
}

func (n *node) extendChild(i, minItem int) int {
	if i > 0 && len(n.children[i-1].inodes) > minItem {
		fmt.Println("Borrow from left sibling")
		n.extendChildWithLeftSibling(i)
		return i
	}
	if i < len(n.children)-1 && len(n.children[i+1].inodes) > minItem {
		fmt.Println("Borrow from right sibling")
		n.extendChildWithRightSibling(i)
		return i
	}
	if i == len(n.children)-1 {
		// i is the rightmost child
		i--
	}
	fmt.Println("Merging")
	n.mergeChildWithRightSibling(i)
	return i
}

func (n *node) extendChildWithLeftSibling(i int) {
	child := n.children[i]
	leftSibling := n.children[i-1]
	child.inodes.insertAt(0, n.inodes[i-1])
	n.inodes[i-1] = leftSibling.inodes.pop()
	if len(child.children) > 0 {
		child.children.insertAt(0, leftSibling.children.pop())
	}
}

func (n *node) extendChildWithRightSibling(i int) {
	child := n.children[i]
	rightSibling := n.children[i+1]
	child.inodes = append(child.inodes, n.inodes[i])
	n.inodes[i] = rightSibling.inodes.removeAt(0)
	if len(child.children) > 0 {
		child.children = append(child.children, rightSibling.children.removeAt(0))
	}
}

func (n *node) mergeChildWithRightSibling(i int) {
	child := n.children[i]
	rightSibling := n.children.removeAt(i + 1)
	child.inodes = append(child.inodes, n.inodes.removeAt(i))
	child.inodes = append(child.inodes, rightSibling.inodes...)
	child.children = append(child.children, rightSibling.children...)
	rightSibling.free()
}

func (n *node) print(w io.Writer, level int) {
	fmt.Fprintf(w, "%sNODE:%v\n", strings.Repeat("  ", level), n.inodes)
	for _, c := range n.children {
		c.print(w, level+1)
	}
}

// split Splits node at given index, return element at index and new node
func (n *node) split(i int) (pair, *node) {
	n.inodes.check()
	new := n.tree.newNode()
	kv := n.inodes[i]
	// Never use direct slice slicing!! Causes very wired bug
	new.inodes = append(new.inodes, n.inodes[i+1:]...)
	n.inodes = n.inodes[:i]
	if len(n.children) > 0 {
		new.children = append(new.children, n.children[i+1:]...)
		n.children = n.children[:i+1]
	}
	return kv, new
}
