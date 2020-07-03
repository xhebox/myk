package tbp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestLeaf(t *testing.T) {
	suite.Run(t, new(TestSuiteLeaf))
}

type TestSuiteLeaf struct {
	TestSuiteUtil
}

func (t *TestSuiteLeaf) TestSplit() {
	cow := NewCow(1)

	first := cow.newLeaf(4)
	first.itemInsert(0, false, []byte("004"), []byte("002"))
	first.itemInsert(0, false, []byte("003"), []byte("002"))
	first.itemInsert(0, false, []byte("002"), []byte("002"))
	first.itemInsert(0, false, []byte("001"), []byte("002"))
	middle, second := first.split(2)

	// first half
	t.CheckItemLen(first, 2, 4)
	t.CheckItem(first, 0, []byte("001"), []byte("002"))
	t.CheckItem(first, 1, []byte("002"), []byte("002"))

	// second half
	t.CheckItemLen(second, 2, 4)
	t.CheckItem(second, 0, []byte("003"), []byte("002"))
	t.CheckItem(second, 1, []byte("004"), []byte("002"))

	// middle and pointers
	t.CheckSlice(middle, []byte("003"), []byte("002"))
}

func (t *TestSuiteLeaf) TestInsert() {
	cow := NewCow(1)

	first := cow.newLeaf(4)
	first.insert([]byte("004"), []byte("002"))
	first.insert([]byte("02"), []byte("002"))
	first.insert([]byte("035"), []byte("002"))
	first.insert([]byte("025"), []byte("002"))

	t.CheckItemLen(first, 4, 4)
	t.CheckItem(first, 0, []byte("004"), []byte("002"))
	t.CheckItem(first, 1, []byte("02"), []byte("002"))
	t.CheckItem(first, 2, []byte("025"), []byte("002"))
	t.CheckItem(first, 3, []byte("035"), []byte("002"))
}

func (t *TestSuiteLeaf) TestRemove() {
	cow := NewCow(1)

	first := cow.newLeaf(4)
	first.insert([]byte("004"), []byte("002"))
	first.insert([]byte("02"), []byte("002"))
	first.insert([]byte("035"), []byte("002"))
	first.insert([]byte("025"), []byte("002"))

	first.remove(nil, removeMax)
	first.remove(nil, removeMin)
	first.remove([]byte("02"), removeItem)
	first.remove([]byte("05"), removeItem)

	t.CheckItemLen(first, 1, 4)
	t.CheckItem(first, 0, []byte("025"), []byte("002"))

	t.Panics(func() {
		first.remove(nil, removeInvalid)
	})
}
