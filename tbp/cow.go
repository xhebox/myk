package tbp

import (
	"sync"
)

type copyOnWrite struct {
	imu    sync.Mutex
	inters []*inter

	lmu    sync.Mutex
	leaves []*leaf
}

func NewCow(max int) *copyOnWrite {
	return &copyOnWrite{
		inters: make([]*inter, 0, max),
		leaves: make([]*leaf, 0, max),
	}
}

func (c *copyOnWrite) newInter(max int) *inter {
	c.imu.Lock()
	defer c.imu.Unlock()

	if len(c.inters) > 0 {
		var ret *inter
		ret, c.inters = c.inters[len(c.inters)-1], c.inters[:len(c.inters)-1]
		if cap(ret.items) < max {
			ret.items = make([][]byte, 0, max)
			ret.children = make([]*node, 0, max+1)
		}
		ret.cow = c
		return ret
	}

	return &inter{
		node: node{
			typ:   Inter,
			items: make([][]byte, 0, max),
			cow:   c,
		},
		children: make([]*node, 0, max+1),
	}
}

func (c *copyOnWrite) freeInter(n *inter) {
	c.imu.Lock()
	defer c.imu.Unlock()

	if len(c.inters) < cap(c.inters) {
		n.items = n.items[:0]
		n.children = n.children[:0]
		n.cow = nil
		c.inters = append(c.inters, n)
	}
}

func (c *copyOnWrite) newLeaf(max int) *leaf {
	c.lmu.Lock()
	defer c.lmu.Unlock()

	if len(c.leaves) > 0 {
		var ret *leaf
		ret, c.leaves = c.leaves[len(c.leaves)-1], c.leaves[:len(c.leaves)-1]
		if cap(ret.items) < max {
			ret.items = make([][]byte, 0, max)
		}
		ret.cow = c
		return ret
	}

	return &leaf{
		node: node{
			typ:   Leaf,
			items: make([][]byte, 0, max),
			cow:   c,
		},
	}
}

func (c *copyOnWrite) freeLeaf(n *leaf) {
	c.lmu.Lock()
	defer c.lmu.Unlock()

	if len(c.leaves) < cap(c.leaves) {
		n.items = n.items[:0]
		n.cow = nil
		c.leaves = append(c.leaves, n)
	}
}
