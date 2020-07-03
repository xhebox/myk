package tbp

import ()

type leaf struct {
	node
}

func (n *leaf) mutableFor(cow *copyOnWrite) *leaf {
	r := cow.newLeaf(cap(n.items))
	r.items = r.items[:len(n.items)]
	for i := range n.items {
		r.items[i] = append(r.items[i], n.items[i]...)
	}
	return r
}

func (n *leaf) split(i int) ([]byte, *leaf) {
	middle := n.items[i]

	next := n.cow.newLeaf(cap(n.items))
	next.items = append(next.items, n.items[i:]...)
	n.items = n.items[:i]

	return middle, next
}

func (n *leaf) insert(key, val []byte) bool {
	i, found := n.itemFind(key)
	n.itemInsert(i, found, key, val)
	return !found
}

func (n *leaf) remove(key []byte, typ toRemove) ([]byte, bool) {
	if len(n.items) == 0 {
		return nil, false
	}

	k := true
	r := false
	switch typ {
	case removeMax:
		n.itemRemove(-1)
	case removeMin:
		n.itemRemove(0)
		r = true
	case removeItem:
		i, found := n.itemFind(key)
		if found {
			n.itemRemove(i)
			r = i == 0
		}
		k = found
	default:
		panic("invalid type")
	}
	if r {
		return n.items[0], k
	} else {
		return nil, k
	}
}
