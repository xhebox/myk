package tbp

import (
	//"bytes"
	//"encoding/binary"
	//"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/suite"
)

func TestNode(t *testing.T) {
	suite.Run(t, new(TestSuiteNode))
}

type TestSuiteNode struct {
	TestSuiteUtil
}

func (t *TestSuiteNode) TestItemInsert() {
	cow := NewCow(1)

	leaf := cow.newLeaf(4)

	// normal insert
	leaf.itemInsert(0, false, []byte("001"), []byte("002"))
	t.CheckItem(leaf, 0, []byte("001"), []byte("002"))

	// found replace
	leaf.itemInsert(0, true, []byte("001"), []byte("78910"))
	t.CheckItem(leaf, 0, []byte("001"), []byte("78910"))

	// direct insert for split/fixup
	leaf.itemInsert(0, false, nil, append([]byte{}, leaf.items[0][:7]...))
	t.CheckItem(leaf, 0, []byte("001"), nil)

	// multiple insertion
	leaf.itemInsert(0, true, []byte("001"), []byte("002"))
	leaf.itemInsert(1, false, []byte("001"), []byte("003"))
	leaf.itemInsert(0, false, []byte("001"), []byte("005"))

	t.CheckItem(leaf, 0, []byte("001"), []byte("005"))
	t.CheckItem(leaf, 1, []byte("001"), []byte("002"))
	t.CheckItem(leaf, 2, []byte("001"), []byte("003"))
	t.CheckItem(leaf, 3, []byte("001"), []byte("78910"))
}

func (t *TestSuiteNode) TestItemFind() {
	cow := NewCow(1)

	leaf := cow.newLeaf(4)
	leaf.itemInsert(0, false, []byte("004"), []byte("002"))
	leaf.itemInsert(0, false, []byte("003"), []byte("002"))
	leaf.itemInsert(0, false, []byte("002"), []byte("002"))
	leaf.itemInsert(0, false, []byte("001"), []byte("002"))

	i, found := leaf.itemFind([]byte("000"))
	t.Equal(0, i)
	t.Equal(false, found)

	i, found = leaf.itemFind([]byte("009"))
	t.Equal(4, i)
	t.Equal(false, found)

	i, found = leaf.itemFind([]byte("002"))
	t.Equal(1, i)
	t.Equal(true, found)
}

func (t *TestSuiteNode) TestItemRemove() {
	cow := NewCow(1)

	leaf := cow.newLeaf(4)
	leaf.itemInsert(0, false, []byte("004"), []byte("002"))
	leaf.itemInsert(0, false, []byte("003"), []byte("002"))
	leaf.itemInsert(0, false, []byte("002"), []byte("002"))
	leaf.itemInsert(0, false, []byte("001"), []byte("002"))

	leaf.itemRemove(-1)
	leaf.itemRemove(1)

	t.CheckItem(leaf, 0, []byte("001"), []byte("002"))
	t.CheckItem(leaf, 1, []byte("003"), []byte("002"))

	leaf.itemRemove(0)

	t.CheckItem(leaf, 0, []byte("003"), []byte("002"))
}

func (t *TestSuiteNode) TestSplit() {
	cow := NewCow(1)

	first := cow.newLeaf(4)
	second := cow.newLeaf(4)
	inode := cow.newInter(4)

	// dispatching
	t.NotPanics(func() {
		first.itemInsert(0, false, []byte("004"), []byte("002"))
		first.itemInsert(0, false, []byte("003"), []byte("002"))
		first.itemInsert(0, false, []byte("002"), []byte("002"))
		first.itemInsert(0, false, []byte("001"), []byte("002"))
		first.node.split(2)

		inode.itemInsert(0, false, []byte("004"), []byte("002"))
		inode.itemInsert(0, false, []byte("003"), []byte("002"))
		inode.itemInsert(0, false, []byte("002"), []byte("002"))
		inode.itemInsert(0, false, []byte("001"), []byte("002"))
		inode.children = append(inode.children, (*node)(unsafe.Pointer(first)), nil, nil, nil, (*node)(unsafe.Pointer(second)))
		inode.node.split(2)
	})

	t.Panics(func() {
		first.typ = Invalid
		first.itemInsert(0, false, []byte("002"), []byte("002"))
		first.itemInsert(0, false, []byte("001"), []byte("002"))
		first.node.split(2)
	})
}

func (t *TestSuiteNode) TestInsert() {
	cow := NewCow(1)

	l1 := cow.newLeaf(4)
	l1.itemInsert(0, false, []byte("001"), []byte("002"))
	l1.itemInsert(1, false, []byte("002"), []byte("002"))
	l1.itemInsert(2, false, []byte("003"), []byte("002"))

	l2 := cow.newLeaf(4)
	l2.itemInsert(0, false, []byte("004"), []byte("002"))
	l2.itemInsert(1, false, []byte("005"), []byte("002"))
	l2.itemInsert(2, false, []byte("006"), []byte("002"))

	inode := cow.newInter(4)
	inode.itemInsert(0, false, []byte("004"), nil)
	inode.children = append(inode.children, (*node)(unsafe.Pointer(l1)), (*node)(unsafe.Pointer(l2)))

	t.Panics(func() {
		inode.typ = Invalid
		inode.node.insert([]byte("007"), []byte("002"))
	})
	inode.typ = Inter

	t.NotPanics(func() {
		inode.node.insert([]byte("007"), []byte("002"))
	})
}

func (t *TestSuiteNode) TestRemove() {
	cow := NewCow(1)

	l1 := cow.newLeaf(4)
	l1.itemInsert(0, false, []byte("001"), []byte("002"))
	l1.itemInsert(1, false, []byte("002"), []byte("002"))
	l1.itemInsert(2, false, []byte("003"), []byte("002"))

	l2 := cow.newLeaf(4)
	l2.itemInsert(0, false, []byte("004"), []byte("002"))
	l2.itemInsert(1, false, []byte("005"), []byte("002"))
	l2.itemInsert(2, false, []byte("006"), []byte("002"))

	inode := cow.newInter(4)
	inode.itemInsert(0, false, []byte("004"), nil)
	inode.children = append(inode.children, (*node)(unsafe.Pointer(l1)), (*node)(unsafe.Pointer(l2)))

	t.Panics(func() {
		inode.typ = Invalid
		inode.node.remove([]byte("004"), removeItem)
	})

	inode.typ = Inter
	t.NotPanics(func() {
		inode.node.remove([]byte("004"), removeItem)
	})
}
