package tbp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTree(t *testing.T) {
	suite.Run(t, new(TestSuiteTree))
}

type TestSuiteTree struct {
	TestSuiteUtil
}

func (t *TestSuiteTree) TestInsertAndIter() {
	data := [][]byte{
		[]byte("000"),
		[]byte("002"),
		[]byte("003"),
		[]byte("004"),

		[]byte("005"),
		[]byte("006"),
		[]byte("007"),
		[]byte("008"),

		[]byte("009"),
		[]byte("010"),
		[]byte("011"),
		[]byte("012"),

		[]byte("013"),
		[]byte("014"),
		[]byte("015"),
		[]byte("016"),

		[]byte("018"),
		[]byte("019"),
		[]byte("020"),
		[]byte("021"),
	}

	tree := NewTree(func(o *TreeOption) {
		o.NodeSize = 4
	})

	randData := append([][]byte{}, data...)

	rand.Shuffle(len(data), func(i, j int) {
		randData[i], randData[j] = randData[j], randData[i]
	})

	for i := range randData {
		tree.Insert(randData[i], []byte("002"))
	}
	t.Panics(func() {
		tree.Insert(nil, nil)
	})
	t.Panics(func() {
		tree.Insert([]byte("002"), nil)
	})

	j := 0
	it := tree.Iter(nil, nil, false, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		if bytes.Compare(key, it.GetKey()) != 0 {
			t.Fail("inconsistent key")
		}
		if bytes.Compare(val, it.GetVal()) != 0 {
			t.Fail("inconsistent val")
		}
		j++
		it.Next()
	}

	j = len(data) - 1
	it = tree.Iter(nil, []byte("001"), true, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		if bytes.Compare(key, it.GetKey()) != 0 {
			t.Fail("inconsistent key")
		}
		if bytes.Compare(val, it.GetVal()) != 0 {
			t.Fail("inconsistent val")
		}
		j--
		it.Next()
	}

	j = 5
	it = tree.Iter([]byte("006"), []byte("018"), false, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j++
		it.Next()
	}

	j = 4
	it = tree.Iter([]byte("005"), []byte("017"), false, true)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j++
		it.Next()
	}

	j = len(data) - 1
	it = tree.Iter([]byte("021"), nil, true, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j--
		it.Next()
	}

	j = len(data) - 5
	it = tree.Iter([]byte("017"), []byte("001"), true, true)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j--
		it.Next()
	}
}

func (t *TestSuiteTree) TestReset() {
	tree := NewTree()

	tree.Insert([]byte("004"), []byte("002"))

	tree.Reset()
	t.Nil(tree.root)
}

func (t *TestSuiteTree) TestRemove() {
	data := [][]byte{
		[]byte("000"),
		[]byte("002"),
		[]byte("003"),
		[]byte("004"),

		[]byte("005"),
		[]byte("006"),
		[]byte("007"),
	}

	tree := NewTree(func(o *TreeOption) {
		o.NodeSize = 4
	})

	for i := range data {
		tree.Insert(data[i], []byte("002"))
	}

	tree.RemoveMax()
	tree.RemoveMin()
	tree.Remove([]byte("004"))
	t.Panics(func() {
		tree.Remove(nil)
	})
}

func (t *TestSuiteTree) TestClone() {
	data := [][]byte{
		[]byte("000"),
		[]byte("002"),
		[]byte("003"),
		[]byte("004"),

		[]byte("005"),
		[]byte("006"),
		[]byte("007"),
	}

	tree1 := NewTree(func(o *TreeOption) {
		o.NodeSize = 4
	})

	for i := range data {
		tree1.Insert(data[i], []byte("002"))
	}

	tree2 := tree1.Clone(NewCow(64))
	tree2.Remove([]byte("007"))
	t.Equal(len(data)-1, tree2.Size())

	j := 0
	it := tree1.Iter(nil, nil, false, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j++
		it.Next()
	}
	if j != len(data) {
		t.Fail("tree2 affected tree1")
	}

	j = 0
	it = tree2.Iter(nil, nil, false, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		it.Next()
		j++
	}
	if j != len(data)-1 {
		t.Fail("tree2 fails to delete")
	}

	tree3 := tree2.Clone(NewCow(64))
	tree3.Insert([]byte("007"), []byte("002"))

	j = 0
	it = tree3.Iter(nil, nil, false, false)
	for it.Valid() {
		key, val := it.Get()
		if bytes.Compare(key, data[j]) != 0 || bytes.Compare(val, []byte("002")) != 0 {
			t.T().Fatalf("inconsistent kv[%d]: [%s|%s]=[%s]", j, key, val, data[j])
		}
		j++
		it.Next()
	}
	if j != len(data) {
		t.Fail("tree2 affected tree3")
	}

	t.Equal(len(data), tree3.Size())
}

func (t *TestSuiteTree) TestGet() {
	data := [][]byte{
		[]byte("000"),
		[]byte("002"),
		[]byte("003"),
		[]byte("004"),

		[]byte("005"),
		[]byte("006"),
		[]byte("007"),
	}

	tree := NewTree(func(o *TreeOption) {
		o.NodeSize = 4
	})

	for i := range data {
		tree.Insert(data[i], data[i])
	}

	t.Equal(data[0], tree.Get(nil, removeMin))
	t.Equal(data[len(data)-1], tree.Get(nil, removeMax))
	t.Equal([]byte("004"), tree.Get([]byte("004"), removeItem))
}

func (t *TestSuiteTree) TestRandom() {
	buf := make([][128]byte, 1024*256)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(rand.Int()))
	}

	p := NewTree()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}

	for i := range buf {
		t.Equal(p.Get(buf[i][:16], removeItem), buf[i][16:])
	}
}
