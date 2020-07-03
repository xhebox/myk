package tbp

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/suite"
)

func TestCow(t *testing.T) {
	suite.Run(t, new(TestSuiteCow))
}

type TestSuiteCow struct {
	TestSuiteUtil
}

func (t *TestSuiteCow) TestLeafAlloc() {
	cow1 := NewCow(1)

	// leaf allocation
	leaf1 := cow1.newLeaf(4)
	t.CheckType(leaf1, Leaf)
	t.CheckItemLen(leaf1, 0, 4)
}

func (t *TestSuiteCow) TestLeafMutable() {
	cow1 := NewCow(1)
	cow2 := NewCow(1)

	leaf1 := cow1.newLeaf(4)
	leaf1.items = append(leaf1.items, nil)

	// leaf mutableFor
	leaf2 := leaf1.mutableFor(cow2)
	t.CheckType(leaf2, Leaf)
	t.CheckItemLen(leaf2, 1, 4)
	// mutation should allocate new slice, otherwise value copy
	t.NotSame(leaf2.cow, leaf1.cow)
	t.NotSame(leaf2.items, leaf1.items)
}

func (t *TestSuiteCow) TestLeafFree() {
	cow1 := NewCow(1)
	cow2 := NewCow(1)

	leaf1 := cow1.newLeaf(4)
	leaf2 := leaf1.mutableFor(cow2)

	// free in another list
	cow1.freeLeaf(leaf2)
	cow2.freeLeaf(leaf1)

	leaf3 := cow1.newLeaf(4)
	t.CheckItemLen(leaf3, 0, 4)

	leaf4 := cow2.newLeaf(8)
	t.CheckItemLen(leaf4, 0, 8)
}

func (t *TestSuiteCow) TestInterAlloc() {
	cow1 := NewCow(1)

	// inter allocation
	inter1 := cow1.newInter(4)
	t.CheckType(inter1, Inter)
	t.CheckItemLen(inter1, 0, 4)
	t.CheckInterLen(inter1, 0, 5)
}

func (t *TestSuiteCow) TestInterMutable() {
	cow1 := NewCow(1)
	cow2 := NewCow(1)

	inter1 := cow1.newInter(4)
	inter1.items = append(inter1.items, nil)

	// inter mutableFor
	inter2 := inter1.mutableFor(cow2)
	t.Equal(Inter, inter2.typ)
	t.CheckItemLen(inter2, 1, 4)
	t.CheckInterLen(inter2, 0, 5)
	t.NotSame(inter2.cow, inter1.cow)
	t.NotSame(inter2.items, inter1.items)
	t.NotSame(inter2.children, inter1.children)

	// inter mutableChild
	inter2.children = append(inter2.children, &inter1.node)
	inter3 := (*inter)(unsafe.Pointer(inter2.mutableChild(0)))
	t.CheckItemLen(inter3, 1, 4)
	t.CheckInterLen(inter3, 0, 5)
	t.NotSame(inter3.cow, inter1.cow)
	t.NotSame(inter3.items, inter1.items)
	t.NotSame(inter3.children, inter1.children)
}

func (t *TestSuiteCow) TestInterFree() {
	cow1 := NewCow(1)
	cow2 := NewCow(1)

	inter1 := cow1.newInter(4)
	inter2 := inter1.mutableFor(cow2)

	cow1.freeInter(inter2)
	cow2.freeInter(inter1)

	inter3 := cow1.newInter(4)
	t.CheckItemLen(inter3, 0, 4)
	t.CheckInterLen(inter3, 0, 5)

	inter4 := cow2.newInter(8)
	t.CheckItemLen(inter4, 0, 8)
	t.CheckInterLen(inter4, 0, 9)
}

func (t *TestSuiteCow) TestNodeMutable() {
	cow1 := NewCow(1)

	// same cow
	inter1 := cow1.newInter(4)
	inter2 := inter1.node.mutableFor(cow1)
	t.Same((*inter)(unsafe.Pointer(inter2)), inter1)

	leaf1 := cow1.newLeaf(4)
	leaf2 := leaf1.node.mutableFor(cow1)
	t.Same((*leaf)(unsafe.Pointer(leaf2)), leaf1)

	// different cow
	cow2 := NewCow(1)

	t.NotPanics(func() {
		leaf1.node.mutableFor(cow2)
		inter1.node.mutableFor(cow2)
	})

	t.Panics(func() {
		leaf1.typ = Invalid
		leaf1.node.mutableFor(cow2)
	})
}
