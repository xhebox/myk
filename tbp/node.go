package tbp

import (
	"bytes"
	"sort"
	"unsafe"
)

type nodeType = int

const (
	Inter nodeType = iota
	Leaf
	Invalid
)

type node struct {
	typ   nodeType
	cow   *copyOnWrite
	items [][]byte
}

// key/val code
func putUint32(b []byte, v uint32) {
	_ = b[3] // BCE
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func getUint32(b []byte) uint32 {
	_ = b[3] // BCE
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (n *node) itemGetKey(idx int) []byte {
	klen := int(getUint32(n.items[idx]))
	return n.items[idx][4 : 4+klen]
}

func (n *node) itemGetVal(idx int) []byte {
	klen := int(getUint32(n.items[idx]))
	return n.items[idx][4+klen:]
}

func (n *node) itemGet(idx int) ([]byte, []byte) {
	klen := int(getUint32(n.items[idx]))
	return n.items[idx][4 : 4+klen], n.items[idx][4+klen:]
}

func (n *node) itemFind(k []byte) (int, bool) {
	i := sort.Search(len(n.items), func(i int) bool {
		return bytes.Compare(k, n.itemGetKey(i)) < 0
	})
	if i > 0 && bytes.Compare(n.itemGetKey(i-1), k) == 0 {
		return i - 1, true
	}
	return i, false
}

func (n *node) itemInsert(idx int, found bool, key, val []byte) {
	if !found {
		n.items = append(n.items, nil)
		copy(n.items[idx+1:], n.items[idx:])
		if len(key) != 0 {
			buf := make([]byte, 4+len(key)+len(val))
			putUint32(buf, uint32(len(key)))
			copy(buf[4:], key)
			copy(buf[4+len(key):], val)
			n.items[idx] = buf
		} else {
			// internal exchange item
			n.items[idx] = val
		}
		return
	}

	buf := n.items[idx]
	n.items[idx] = append(buf[:4+len(key):cap(buf)], val...)
}

func (n *node) itemRemove(i int) []byte {
	if i == -1 {
		i = len(n.items) - 1
	}
	r := n.items[i]
	copy(n.items[i:], n.items[i+1:])
	n.items = n.items[:len(n.items)-1]
	return r
}

// dispatch code
func (n *node) mutableFor(cow *copyOnWrite) *node {
	if n.cow == cow {
		return n
	}

	switch n.typ {
	case Inter:
		m := (*inter)(unsafe.Pointer(n)).mutableFor(cow)
		return (*node)(unsafe.Pointer(m))
	case Leaf:
		m := (*leaf)(unsafe.Pointer(n)).mutableFor(cow)
		return (*node)(unsafe.Pointer(m))
	default:
		panic("???")
	}
}

func (n *node) split(i int) ([]byte, *node) {
	switch n.typ {
	case Inter:
		middle, newnode := (*inter)(unsafe.Pointer(n)).split(i)
		return middle, (*node)(unsafe.Pointer(newnode))
	case Leaf:
		middle, newnode := (*leaf)(unsafe.Pointer(n)).split(i)
		return middle, (*node)(unsafe.Pointer(newnode))
	default:
		panic("???")
	}
}

func (n *node) insert(key, val []byte) bool {
	switch n.typ {
	case Leaf:
		return (*leaf)(unsafe.Pointer(n)).insert(key, val)
	case Inter:
		return (*inter)(unsafe.Pointer(n)).insert(key, val)
	default:
		panic("???")
	}
}

type toRemove = int

const (
	removeItem toRemove = iota
	removeMin
	removeMax
	removeInvalid
)

func (n *node) remove(key []byte, typ toRemove) ([]byte, bool) {
	switch n.typ {
	case Leaf:
		return (*leaf)(unsafe.Pointer(n)).remove(key, typ)
	case Inter:
		return (*inter)(unsafe.Pointer(n)).remove(key, typ)
	default:
		panic("???")
	}
}
