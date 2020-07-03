package tbp

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/suite"
)

func TestInter(t *testing.T) {
	suite.Run(t, new(TestSuiteInter))
}

type TestSuiteInter struct {
	TestSuiteUtil
}

func (t *TestSuiteInter) TestChildInsert() {
	cow := NewCow(1)

	inode := cow.newInter(4)
	inode.childInsert(0, nil)
	inode.childInsert(1, &inode.node)

	t.CheckInterLen(inode, 2, 5)
	t.Nil(inode.children[0])
	t.Same(&inode.node, inode.children[1])
}

func (t *TestSuiteInter) TestChildRemove() {
	cow := NewCow(1)

	inode := cow.newInter(4)
	inode.childInsert(0, &inode.node)
	inode.childInsert(0, nil)
	inode.childInsert(0, &inode.node)

	t.CheckInterLen(inode, 3, 5)
	t.Same((*node)(nil), inode.childRemove(1))
	t.Same(&inode.node, inode.childRemove(1))
}

func (t *TestSuiteInter) TestSplit() {
	cow := NewCow(1)

	l1 := cow.newLeaf(4)
	l2 := cow.newLeaf(4)

	i1 := cow.newInter(4)
	i1.itemInsert(0, false, []byte("001"), nil)
	i1.itemInsert(1, false, []byte("002"), nil)
	i1.itemInsert(2, false, []byte("003"), nil)
	i1.itemInsert(3, false, []byte("004"), nil)
	i1.children = append(i1.children, (*node)(unsafe.Pointer(l1)), nil, nil, nil, (*node)(unsafe.Pointer(l2)))
	mid, i2 := i1.split(2)

	// first half
	t.CheckItem(i1, 0, []byte("001"), nil)
	t.CheckItem(i1, 1, []byte("002"), nil)
	t.Same(l1, (*leaf)(unsafe.Pointer(i1.children[0])))

	// middle
	t.CheckSlice(mid, []byte("003"), nil)

	// second half
	t.CheckItem(i2, 0, []byte("004"), nil)
	t.Same(l2, (*leaf)(unsafe.Pointer(i2.children[1])))
}

func (t *TestSuiteInter) TestMaybeSplit() {
	cow := NewCow(1)

	n1 := cow.newLeaf(4)

	i1 := cow.newInter(4)
	i1.itemInsert(0, false, []byte("001"), nil)
	i1.itemInsert(1, false, []byte("002"), nil)
	i1.itemInsert(2, false, []byte("003"), nil)
	i1.itemInsert(3, false, []byte("004"), nil)
	i1.children = append(i1.children,
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)))

	i2 := cow.newInter(4)
	i2.children = append(i2.children, (*node)(unsafe.Pointer(i1)))

	t.True(i2.maybeSplit(0))
	t.CheckItemLen(i2, 1, 4)
	t.CheckInterLen(i2, 2, 5)

	t.CheckItem(i2, 0, []byte("003"), nil)

	// have checked split, so just check the number
	m1 := (*inter)(unsafe.Pointer(i2.children[0]))
	t.Same(m1, i1)
	t.CheckInterLen(m1, 3, 5)

	m2 := (*inter)(unsafe.Pointer(i2.children[1]))
	t.CheckInterLen(m2, 2, 5)

	t.False(i2.maybeSplit(0))
}

func (t *TestSuiteInter) TestMergeRightInter() {
	cow := NewCow(1)

	n1 := cow.newLeaf(4)
	n2 := cow.newLeaf(4)
	n3 := cow.newLeaf(4)
	n4 := cow.newLeaf(4)
	n5 := cow.newLeaf(4)

	i1 := cow.newInter(4)
	i1.itemInsert(0, false, []byte("002"), nil)
	i1.children = append(i1.children,
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n2)))

	i2 := cow.newInter(4)
	i2.itemInsert(0, false, []byte("004"), nil)
	i2.itemInsert(1, false, []byte("005"), nil)
	i2.children = append(i2.children,
		(*node)(unsafe.Pointer(n3)),
		(*node)(unsafe.Pointer(n4)),
		(*node)(unsafe.Pointer(n5)))

	i3 := cow.newInter(4)
	i3.itemInsert(0, false, []byte("003"), nil)
	i3.children = append(i3.children, (*node)(unsafe.Pointer(i1)), (*node)(unsafe.Pointer(i2)))

	// panic case
	t.Panics(func() {
		i3.children[0].typ = Invalid
		i3.mergeRight(0)
	})

	// merge it
	i3.children[0].typ = Inter
	i3.mergeRight(0)
	t.CheckItemLen(i3, 0, 4)
	t.CheckInterLen(i3, 1, 5)

	m := (*inter)(unsafe.Pointer(i3.children[0]))

	t.CheckItemLen(m, 4, 4)
	t.CheckInterLen(m, 5, 5)

	t.CheckItem(m, 0, []byte("002"), nil)
	t.CheckItem(m, 1, []byte("003"), nil)
	t.CheckItem(m, 2, []byte("004"), nil)
	t.CheckItem(m, 3, []byte("005"), nil)

	t.Same(n1, (*leaf)(unsafe.Pointer(m.children[0])))
	t.Same(n2, (*leaf)(unsafe.Pointer(m.children[1])))
	t.Same(n3, (*leaf)(unsafe.Pointer(m.children[2])))
	t.Same(n4, (*leaf)(unsafe.Pointer(m.children[3])))
	t.Same(n5, (*leaf)(unsafe.Pointer(m.children[4])))
}

func (t *TestSuiteInter) TestMergeRightLeaf() {
	cow := NewCow(1)

	l1 := cow.newLeaf(4)
	l1.itemInsert(0, false, []byte("002"), []byte("002"))
	l1.itemInsert(1, false, []byte("003"), []byte("002"))

	l2 := cow.newLeaf(4)
	l2.itemInsert(0, false, []byte("004"), []byte("002"))
	l2.itemInsert(1, false, []byte("005"), []byte("002"))

	inode := cow.newInter(4)
	inode.itemInsert(0, false, []byte("004"), nil)
	inode.children = append(inode.children, (*node)(unsafe.Pointer(l1)), (*node)(unsafe.Pointer(l2)))

	// merge it
	inode.mergeRight(0)
	t.CheckItemLen(inode, 0, 4)
	t.CheckInterLen(inode, 1, 5)

	m := (*leaf)(unsafe.Pointer(inode.children[0]))

	t.CheckItemLen(m, 4, 4)

	t.CheckItem(m, 0, []byte("002"), []byte("002"))
	t.CheckItem(m, 1, []byte("003"), []byte("002"))
	t.CheckItem(m, 2, []byte("004"), []byte("002"))
	t.CheckItem(m, 3, []byte("005"), []byte("002"))
}

func (t *TestSuiteInter) TestStealLeft() {
	cow := NewCow(1)

	n1 := cow.newLeaf(4)
	n2 := cow.newLeaf(4)

	i1 := cow.newInter(4)
	i1.itemInsert(0, false, []byte("002"), nil)
	i1.itemInsert(1, false, []byte("003"), nil)
	i1.itemInsert(2, false, []byte("004"), nil)
	i1.children = append(i1.children,
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)))

	i2 := cow.newInter(4)
	i2.itemInsert(0, false, []byte("005"), nil)
	i2.itemInsert(1, false, []byte("006"), nil)
	i2.children = append(i2.children,
		(*node)(unsafe.Pointer(n2)),
		(*node)(unsafe.Pointer(n2)),
		(*node)(unsafe.Pointer(n2)))

	i3 := cow.newInter(4)
	i3.itemInsert(0, false, []byte("004"), nil)
	i3.children = append(i3.children, (*node)(unsafe.Pointer(i1)), (*node)(unsafe.Pointer(i2)))

	i3.stealLeft(1)
	t.CheckItemLen(i3, 1, 4)
	t.CheckInterLen(i3, 2, 5)
	t.CheckItemLen(i3.children[0], 2, 4)
	t.CheckInterLen(i3.children[0], 3, 5)
	t.CheckItemLen(i3.children[1], 3, 4)
	t.CheckInterLen(i3.children[1], 4, 5)
	t.CheckItem(i3, 0, []byte("004"), nil)
	t.CheckItem(i1, 0, []byte("002"), nil)
	t.CheckItem(i1, 1, []byte("003"), nil)
	t.CheckItem(i2, 0, []byte("004"), nil)
	t.CheckItem(i2, 1, []byte("005"), nil)
	t.CheckItem(i2, 2, []byte("006"), nil)
}

func (t *TestSuiteInter) TestStealRight() {
	cow := NewCow(1)

	n1 := cow.newLeaf(4)
	n2 := cow.newLeaf(4)

	i1 := cow.newInter(4)
	i1.itemInsert(0, false, []byte("002"), nil)
	i1.itemInsert(1, false, []byte("003"), nil)
	i1.children = append(i1.children,
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)),
		(*node)(unsafe.Pointer(n1)))

	i2 := cow.newInter(4)
	i2.itemInsert(0, false, []byte("004"), nil)
	i2.itemInsert(1, false, []byte("005"), nil)
	i2.itemInsert(2, false, []byte("006"), nil)
	i2.children = append(i2.children,
		(*node)(unsafe.Pointer(n2)),
		(*node)(unsafe.Pointer(n2)),
		(*node)(unsafe.Pointer(n2)),
		(*node)(unsafe.Pointer(n2)))

	i3 := cow.newInter(4)
	i3.itemInsert(0, false, []byte("003"), nil)
	i3.children = append(i3.children, (*node)(unsafe.Pointer(i1)), (*node)(unsafe.Pointer(i2)))

	i3.stealRight(0)
	t.CheckItemLen(i3, 1, 4)
	t.CheckInterLen(i3, 2, 5)
	t.CheckItemLen(i3.children[0], 3, 4)
	t.CheckInterLen(i3.children[0], 4, 5)
	t.CheckItemLen(i3.children[1], 2, 4)
	t.CheckInterLen(i3.children[1], 3, 5)
	t.CheckItem(i3, 0, []byte("005"), nil)
	t.CheckItem(i1, 0, []byte("002"), nil)
	t.CheckItem(i1, 1, []byte("003"), nil)
	t.CheckItem(i1, 2, []byte("004"), nil)
	t.CheckItem(i2, 0, []byte("005"), nil)
	t.CheckItem(i2, 1, []byte("006"), nil)
}

func (t *TestSuiteInter) TestInsert() {
	cow := NewCow(1)

	l1 := cow.newLeaf(4)
	l1.itemInsert(0, false, []byte("002"), []byte("002"))
	l1.itemInsert(1, false, []byte("003"), []byte("002"))

	l2 := cow.newLeaf(4)
	l2.itemInsert(0, false, []byte("004"), []byte("002"))
	l2.itemInsert(1, false, []byte("005"), []byte("002"))
	l2.itemInsert(2, false, []byte("006"), []byte("002"))

	inode := cow.newInter(4)
	inode.itemInsert(0, false, []byte("004"), nil)
	inode.children = append(inode.children, (*node)(unsafe.Pointer(l1)), (*node)(unsafe.Pointer(l2)))

	// replace
	inode.insert([]byte("004"), []byte("003"))
	t.CheckItem(l2, 0, []byte("004"), []byte("003"))

	// normal insert
	inode.insert([]byte("007"), []byte("003"))
	t.CheckItemLen(inode, 1, 4)
	t.CheckInterLen(inode, 2, 5)
	t.CheckItem(l2, 3, []byte("007"), []byte("003"))

	// split insert
	inode.insert([]byte("008"), []byte("003"))
	t.CheckItemLen(inode, 2, 4)
	t.CheckInterLen(inode, 3, 5)
	t.CheckItem(inode.children[2], 2, []byte("008"), []byte("003"))
}

func (t *TestSuiteInter) TestRemove() {
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

	// panic remove
	t.Panics(func() {
		inode.remove(nil, removeInvalid)
	})

	// remove max
	t.CheckItemLen(l2, 3, 4)
	inode.remove(nil, removeMax)
	t.CheckItemLen(l2, 2, 4)
	t.CheckItem(l1, 0, []byte("001"), []byte("002"))
	t.CheckItem(l1, 1, []byte("002"), []byte("002"))
	t.CheckItem(l1, 2, []byte("003"), []byte("002"))
	t.CheckItem(l2, 0, []byte("004"), []byte("002"))
	t.CheckItem(l2, 1, []byte("005"), []byte("002"))

	// remove min
	t.CheckItemLen(l1, 3, 4)
	inode.remove(nil, removeMin)
	t.CheckItemLen(l1, 2, 4)
	t.CheckItem(l1, 0, []byte("002"), []byte("002"))
	t.CheckItem(l1, 1, []byte("003"), []byte("002"))
	t.CheckItem(l2, 0, []byte("004"), []byte("002"))
	t.CheckItem(l2, 1, []byte("005"), []byte("002"))

	// normal remove
	l2.itemInsert(2, false, []byte("006"), []byte("002"))
	t.CheckItemLen(l2, 3, 4)
	inode.remove([]byte("005"), removeItem)
	t.CheckItemLen(l2, 2, 4)
	l2.itemInsert(1, false, []byte("005"), []byte("002"))
	t.CheckItem(l1, 0, []byte("002"), []byte("002"))
	t.CheckItem(l1, 1, []byte("003"), []byte("002"))
	t.CheckItem(l2, 0, []byte("004"), []byte("002"))
	t.CheckItem(l2, 1, []byte("005"), []byte("002"))
	t.CheckItem(l2, 2, []byte("006"), []byte("002"))

	// fixup remove
	inode.remove([]byte("004"), removeItem)
	t.CheckItemLen(inode, 1, 4)
	t.CheckInterLen(inode, 2, 5)
	t.CheckItemLen(l1, 2, 4)
	t.CheckItemLen(l2, 2, 4)
	t.CheckItem(inode, 0, []byte("005"), []byte("002"))
	t.CheckItem(l1, 0, []byte("002"), []byte("002"))
	t.CheckItem(l1, 1, []byte("003"), []byte("002"))
	t.CheckItem(l2, 0, []byte("005"), []byte("002"))
	t.CheckItem(l2, 1, []byte("006"), []byte("002"))

	// left steal remove
	l1.itemInsert(0, false, []byte("001"), []byte("002"))
	inode.remove([]byte("005"), removeItem)
	t.CheckItemLen(inode, 1, 4)
	t.CheckInterLen(inode, 2, 5)
	t.CheckItemLen(l1, 2, 4)
	t.CheckItemLen(l2, 2, 4)
	t.CheckItem(inode, 0, []byte("003"), []byte("002"))
	t.CheckItem(l1, 0, []byte("001"), []byte("002"))
	t.CheckItem(l1, 1, []byte("002"), []byte("002"))
	t.CheckItem(l2, 0, []byte("003"), []byte("002"))
	t.CheckItem(l2, 1, []byte("006"), []byte("002"))

	// right steal remove
	l2.itemInsert(1, false, []byte("005"), []byte("002"))
	inode.remove([]byte("001"), removeItem)
	t.CheckItemLen(inode, 1, 4)
	t.CheckInterLen(inode, 2, 5)
	t.CheckItemLen(l1, 2, 4)
	t.CheckItemLen(l2, 2, 4)
	t.CheckItem(inode, 0, []byte("005"), []byte("002"))
	t.CheckItem(l1, 0, []byte("002"), []byte("002"))
	t.CheckItem(l1, 1, []byte("003"), []byte("002"))
	t.CheckItem(l2, 0, []byte("005"), []byte("002"))
	t.CheckItem(l2, 1, []byte("006"), []byte("002"))

	// merge remove
	inode.remove([]byte("005"), removeItem)
	t.CheckItemLen(inode, 0, 4)
	t.CheckInterLen(inode, 1, 5)
	t.CheckItemLen(l1, 3, 4)
	t.CheckItem(l1, 0, []byte("002"), []byte("002"))
	t.CheckItem(l1, 1, []byte("003"), []byte("002"))
	t.CheckItem(l1, 2, []byte("006"), []byte("002"))
}
