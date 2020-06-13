package btree

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unsafe"

	"github.com/daicang/bt/pkg/page"
)

type inode []KV
type children []*node

// len(children) is 0 or len(inodes) + 1
type node struct {
	pgid     page.Pgid
	tree     *BTree
	inodes   inode
	children children
}

func (n *node) free() {
	f := n.tree.freeList
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.list) < cap(f.list) {
		n.inodes = []KV{}
		n.children = []*node{}
		f.list = append(f.list, n)
	}
}

func (in inode) check() {
	var last KV
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
func (in inode) search(key keyType) (bool, int) {
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
func (in *inode) set(kv KV) {
	found, i := in.search(kv.key)
	if found {
		(*in)[i] = kv
	} else {
		in.insertAt(i, kv)
	}
	// in.check()
}

// insert inserts it at given position, pushing subsequent values
func (in *inode) insertAt(i int, kv KV) {
	*in = append((*in), KV{})
	copy((*in)[i+1:], (*in)[i:])
	(*in)[i] = kv
	// in.check()
}

// remove removes and return KV at given index
func (in *inode) removeAt(i int) KV {
	it := (*in)[i]
	copy((*in)[i:], (*in)[i+1:])
	(*in)[len(*in)-1] = KV{}
	(*in) = (*in)[:len(*in)-1]
	return it
}

func (in *inode) pop() KV {
	it := (*in)[len(*in)-1]
	(*in)[len(*in)-1] = KV{}
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
func (n *node) get(key keyType) (bool, []byte) {
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
			panic(fmt.Sprintf("node %d has %d inode, %d children\n", n.pgid, len(n.inodes), len(n.children)))
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

// set sets value to given key in subtree, return old value
func (n *node) set(key keyType, value []byte, maxItem int) []byte {
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
		return oldValue
	}
	// Leaf node, add to index
	if len(n.children) == 0 {
		// fmt.Printf(" insert leaf at node=%d index=%d\n", n.id, i)
		n.inodes.insertAt(i, newKV(key, value))
		return nil
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
			return oldValue
		}
		if newIndex.key.lessThan(key) {
			i++
		}
	}
	// n.checkChildInodes(i)
	// fmt.Printf("(%d) Insert into child node %d\n", n.id, n.children[i].id)
	ret := n.children[i].set(key, value, maxItem)
	n.checkChildInodes(i)
	return ret
}

// iterator function receives KV and nodeID
type iterator func(KV, page.Pgid)

func (n *node) iterate(iter iterator) {
	hasChild := false
	if len(n.children) > 0 {
		hasChild = true
	}
	for index := 0; index < len(n.inodes); index++ {
		if hasChild {
			n.children[index].iterate(iter)
		}
		iter(n.inodes[index], n.pgid)
	}
	if hasChild {
		n.children[len(n.inodes)].iterate(iter)
	}
}

func (n *node) remove(key keyType, minItem int) KV {
	removedValue := n.removeKey(key, minItem)
	if len(removedValue) > 0 {
		return newKV(key, removedValue)
	}
	return KV{}
}

func (n *node) removeMin(minItem int) KV {
	if len(n.children) == 0 {
		return n.inodes.removeAt(0)
	}
	if len(n.children[0].inodes) <= minItem {
		n.extendChild(0, minItem)
	}
	return n.children[0].removeMin(minItem)
}

func (n *node) removeMax(minItem int) KV {
	if len(n.children) == 0 {
		return n.inodes.pop()
	}
	if len(n.children[len(n.children)-1].inodes) <= minItem {
		n.extendChild(len(n.children)-1, minItem)
	}
	return n.children[len(n.children)-1].removeMax(minItem)
}

func (n *node) removeKey(key keyType, minItem int) []byte {
	found, i := n.inodes.search(key)
	if len(n.children) == 0 {
		if found {
			return n.inodes.removeAt(i).value
		}
		return nil
	}
	// Must check child before removeMax
	if len(n.children[i].inodes) <= minItem {
		n.extendChild(i, minItem)
		found, i = n.inodes.search(key)
	}
	if found {
		removed := n.inodes[i].value
		n.inodes[i] = n.children[i].removeMax(minItem)
		return removed
	}
	n.checkChildInodes(i)
	return n.children[i].removeKey(key, minItem)
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
func (n *node) split(i int) (KV, *node) {
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

// read initializes node from page
func (n *node) read(p *page.Page) {
	isLeaf := ((p.Flags & page.LeafFlag) != 0)
	n.inodes = make(inode, int(p.Count))
	n.pgid = p.ID

	// Restore keys and values
	for i := 0; i < int(p.Count); i++ {
		inode := &n.inodes[i]
		kvm := p.GetKVMeta(uint16(i))
		inode.key = kvm.Key()
		inode.value = kvm.Value()
	}

	// Restore children
	if !isLeaf {
		n.children = make([]*node, int(p.Count)+1)
		for i := 0; i < int(p.Count); i++ {
			inode := &n.inodes[i]
			elem := p.InnerPageElement(uint16(i))
			// inode.pgid = elem.pgid
			inode.key = elem.Key()
			inode.value = elem.Value()

			// _assert(len(inode.key) > 0, "read: zero-length inode key")
		}
	}

	// Save first key so we can find the node in the parent when we spill.
	// if len(n.inodes) > 0 {
	// 	n.key = n.inodes[0].key
	// 	_assert(len(n.key) > 0, "read: zero-length node key")
	// } else {
	// 	n.key = nil
	// }
}

// write writes node to a page
func (n *node) write(p *page.Page) {
	p.Count = uint16(len(n.inodes))
	if len(n.children) != 0 {
		p.Flags |= page.InternalFlag
		childArr := (*[page.MaxAllocSize]page.Pgid)(unsafe.Pointer(p.ChildPtr))
		for i, child := range n.children {
			childArr[i] = child.pgid
		}
	} else {
		p.Flags |= page.LeafFlag
	}

	b := (*[page.MaxAllocSize]byte)(unsafe.Pointer(&p.MetaPtr))[len(n.inodes)*page.KVMetaSize:]
	for i, inode := range n.inodes {
		// Write KVMeta
		kvm := p.GetKVMeta(uint16(i))
		kvm.Keysz = uint32(len(inode.key))
		kvm.Valuesz = uint32(len(inode.value))
		kvm.Offset = uint32(uintptr(unsafe.Pointer(&b[0])) - uintptr(unsafe.Pointer(kvm)))
		// Write key and value
		copy(b[0:], inode.key)
		b = b[len(inode.key):]
		copy(b[0:], inode.value)
		b = b[len(inode.value):]
	}

	// Page overflow

	// klen, vlen := len(item.key), len(item.value)
	// if len(b) < klen+vlen {
	// 	b = (*[maxAllocSize]byte)(unsafe.Pointer(&b[0]))[:]
	// }

	// p.data

}
