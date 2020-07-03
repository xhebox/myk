package tbp

import (
	"unsafe"
)

type TreeOption struct {
	NodeSize    int
	CopyContext *copyOnWrite
}

type Tree struct {
	root *node
	size int
	cow  *copyOnWrite
}

func NewTree(f ...func(*TreeOption)) *Tree {
	var t TreeOption
	for i := range f {
		f[i](&t)
	}
	if t.NodeSize == 0 {
		t.NodeSize = 64
	}
	if t.CopyContext == nil {
		t.CopyContext = NewCow(64)
	}
	return &Tree{
		root: (*node)(unsafe.Pointer(t.CopyContext.newLeaf(t.NodeSize))),
		cow:  t.CopyContext,
	}
}

func (t *Tree) Clone(cow *copyOnWrite) *Tree {
	r := *t
	r.cow = cow
	return &r
}

func (t *Tree) Insert(key, val []byte) {
	if len(key) == 0 || len(val) == 0 {
		panic("empty value is not allowed")
	}
	t.root = t.root.mutableFor(t.cow)
	if len(t.root.items) >= cap(t.root.items) {
		m := t.cow.newInter(cap(t.root.items))
		m.children = append(m.children, t.root)
		m.splitChild(0, cap(t.root.items))
		t.root = (*node)(unsafe.Pointer(m))
	}
	if t.root.insert(key, val) {
		t.size++
	}
}

func (t *Tree) RemoveMin() {
	t.root.remove(nil, removeMin)
}

func (t *Tree) RemoveMax() {
	t.root.remove(nil, removeMax)
}

func (t *Tree) Remove(key []byte) {
	if len(key) == 0 {
		panic("empty value is not allowed")
	}
	t.root = t.root.mutableFor(t.cow)
	if t.root.typ == Inter && len(t.root.items) <= 1 {
		n := (*inter)(unsafe.Pointer(t.root))
		n.mergeRight(0)
		t.root = n.children[0]
	}
	_, k := t.root.remove(key, removeItem)
	if k {
		t.size--
	}
}

func (t *Tree) Reset() {
	t.root = nil
}

func (t *Tree) Size() int {
	return t.size
}

func (t *Tree) Get(k []byte, typ toRemove) []byte {
	var i int
	var found bool

	n := t.root
	for n.typ != Leaf {
		switch typ {
		case removeMax:
			i = len(n.items)
		case removeMin:
			i = 0
		case removeItem:
			i, found = n.itemFind(k)
		default:
			panic("invalid type")
		}
		if found {
			i++
		}
		n = (*inter)(unsafe.Pointer(n)).children[i]
	}
	switch typ {
	case removeMax:
		i = len(n.items) - 1
	case removeMin:
		i = 0
	case removeItem:
		i, _ = n.itemFind(k)
	}
	return n.itemGetVal(i)
}

func (t *Tree) Iter(k, stop []byte, reverse, inclusive bool) Iterator {
	typ := removeItem
	if len(k) == 0 {
		if reverse {
			typ = removeMax
		} else {
			typ = removeMin
		}
	}

	nstack := make([]*node, 0, 8)
	istack := make([]int, 0, 8)

	n := t.root
	for n.typ != Leaf {
		var i int
		var found bool
		switch typ {
		case removeMax:
			i = len(n.items)
		case removeMin:
			i = 0
		case removeItem:
			i, found = n.itemFind(k)
		}
		if found {
			i++
		}
		nstack = append(nstack, n)
		istack = append(istack, i)
		n = (*inter)(unsafe.Pointer(n)).children[i]
	}

	var i int
	var found bool
	switch typ {
	case removeMax:
		i = len(n.items) - 1
	case removeMin:
		i = 0
	case removeItem:
		i, found = n.itemFind(k)
		if reverse && !found {
			// this is a b+ tree, so i >= 0, not found, then i > 0
			i--
		}
	}

	if reverse {
		return &DescendIterator{
			nstack:    nstack,
			istack:    istack,
			focus:     n,
			idx:       i,
			stop:      append([]byte{}, stop...),
			inclusive: inclusive,
		}
	} else {
		return &AscendIterator{
			nstack:    nstack,
			istack:    istack,
			focus:     n,
			idx:       i,
			stop:      append([]byte{}, stop...),
			inclusive: inclusive,
		}
	}
}
