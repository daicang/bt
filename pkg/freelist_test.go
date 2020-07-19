package bt

import "testing"

func TestFreeList(t *testing.T) {
	tree := New(3)
	nodes := []*node{}
	nodeCount := 320

	if tree.freeList.getSize() != 0 {
		t.Fatalf("expected freeList size: 0, get: %d", tree.freeList.getSize())
	}

	for i := 0; i < nodeCount; i++ {
		node := tree.newNode()
		nodes = append(nodes, node)
	}

	if tree.freeList.getSize() != 0 {
		t.Fatalf("expected freeList size: 0, get: %d", tree.freeList.getSize())
	}

	for _, node := range nodes {
		node.free()
	}

	if tree.freeList.getSize() != freelistSize {
		t.Fatalf("expected freeList size: %d, get: %d", freelistSize, tree.freeList.getSize())
	}
}
