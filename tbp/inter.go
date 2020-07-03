package tbp

import (
	"bytes"
	"unsafe"
)

type inter struct {
	node
	children []*node
}

func (n *inter) mutableFor(cow *copyOnWrite) *inter {
	r := cow.newInter(cap(n.items))
	r.items = r.items[:len(n.items)]
	for i := range n.items {
		r.items[i] = append(r.items[i], n.items[i]...)
	}
	r.children = append(r.children, n.children...)
	return r
}

func (n *inter) mutableChild(i int) *node {
	c := n.children[i].mutableFor(n.cow)
	n.children[i] = c
	return c
}

func (n *inter) childInsert(idx int, child *node) {
	n.children = append(n.children, nil)
	copy(n.children[idx+1:], n.children[idx:])
	n.children[idx] = child
}

func (n *inter) childRemove(idx int) *node {
	r := n.children[idx]
	copy(n.children[idx:], n.children[idx+1:])
	n.children = n.children[:len(n.children)-1]
	return r
}

func (n *inter) split(i int) ([]byte, *inter) {
	middle := n.items[i]

	next := n.cow.newInter(cap(n.items))
	next.items = append(next.items, n.items[i+1:]...)
	n.items = n.items[:i]

	next.children = append(next.children, n.children[i+1:]...)
	n.children = n.children[:i+1]
	return middle, next
}

func (n *inter) splitChild(i, max int) {
	first := n.mutableChild(i)
	middle, second := first.split(max / 2)
	n.itemInsert(i, false, nil, middle)
	n.childInsert(i+1, second)
}

func (n *inter) maybeSplit(i int) bool {
	max := cap(n.items)
	if len(n.children[i].items) < max {
		return false
	}
	n.splitChild(i, max)
	return true
}

func (n *inter) stealLeft(i int) {
	right := n.mutableChild(i)
	left := n.mutableChild(i - 1)

	stolen := left.itemRemove(-1)
	right.itemInsert(0, false, nil, stolen)
	n.items[i-1] = stolen

	if right.typ == Inter {
		lin := (*inter)(unsafe.Pointer(left))
		rin := (*inter)(unsafe.Pointer(right))
		rin.childInsert(0, lin.childRemove(len(n.children)-1))
	}
}

func (n *inter) stealRight(i int) {
	left := n.mutableChild(i)
	right := n.mutableChild(i + 1)

	stolen := right.itemRemove(0)
	left.itemInsert(len(left.items), false, nil, stolen)
	n.items[i] = right.items[0]

	if left.typ == Inter {
		lin := (*inter)(unsafe.Pointer(left))
		rin := (*inter)(unsafe.Pointer(right))
		lin.childInsert(len(n.children), rin.childRemove(0))
	}
}

func (n *inter) mergeRight(i int) {
	left := n.mutableChild(i)
	switch left.typ {
	case Leaf:
		n.itemRemove(i)

		right := n.childRemove(i + 1)

		lin := (*leaf)(unsafe.Pointer(left))
		rin := (*leaf)(unsafe.Pointer(right))

		lin.items = append(lin.items, rin.items...)

		n.cow.freeLeaf(rin)
	case Inter:
		right := n.childRemove(i + 1)

		lin := (*inter)(unsafe.Pointer(left))
		rin := (*inter)(unsafe.Pointer(right))

		lin.items = append(lin.items, n.itemRemove(i))
		lin.items = append(lin.items, rin.items...)
		lin.children = append(lin.children, rin.children...)

		n.cow.freeInter(rin)
	default:
		panic("???")
	}
}

func (n *inter) stealRemove(i, min int) int {
	if i > 0 && len(n.children[i-1].items) > min {
		// left can spare one
		n.stealLeft(i)
	} else if i < len(n.items) && len(n.children[i+1].items) > min {
		// right can spare one
		n.stealRight(i)
	} else {
		// merge the right one
		if i >= len(n.items) {
			i--
		}
		n.mergeRight(i)
	}
	return i
}

func (n *inter) insert(key, val []byte) bool {
	i, found := n.itemFind(key)
	if found || (n.maybeSplit(i) && bytes.Compare(n.itemGetKey(i), key) <= 0) {
		i++
	}
	return n.children[i].insert(key, val)
}

func (n *inter) remove(key []byte, typ toRemove) ([]byte, bool) {
	var i int
	var found bool
	switch typ {
	case removeMax:
		i = len(n.items)
	case removeMin:
		i = 0
	case removeItem:
		i, found = n.itemFind(key)
	default:
		panic("invalid type")
	}
	if found {
		i++
	}

	min := cap(n.items) / 2
	if len(n.children[i].items) <= min {
		i = n.stealRemove(i, min)
	}

	r, k := n.mutableChild(i).remove(key, typ)
	if r != nil && i > 0 {
		n.items[i-1] = r
		if i == 1 {
			return r, k
		}
	}
	return nil, k
}
