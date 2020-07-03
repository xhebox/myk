package tbp

import (
	"bytes"
	"unsafe"
)

type Iterator interface {
	GetKey() []byte
	GetVal() []byte
	Get() ([]byte, []byte)
	Valid() bool
	Next()
}

type AscendIterator struct {
	nstack    []*node
	istack    []int
	focus     *node
	idx       int
	stop      []byte
	inclusive bool
}

func (it *AscendIterator) GetKey() []byte {
	return it.focus.itemGetKey(it.idx)
}

func (it *AscendIterator) GetVal() []byte {
	return it.focus.itemGetVal(it.idx)
}

func (it *AscendIterator) Get() ([]byte, []byte) {
	return it.focus.itemGet(it.idx)
}

func (it *AscendIterator) Valid() bool {
	if len(it.stop) == 0 {
		return it.idx != -1
	} else {
		if it.inclusive {
			return it.idx != -1 && bytes.Compare(it.GetKey(), it.stop) <= 0
		} else {
			return it.idx != -1 && bytes.Compare(it.GetKey(), it.stop) < 0
		}
	}
}

func (it *AscendIterator) Next() {
	it.idx++
	if it.idx < len(it.focus.items) {
		return
	}
	for len(it.nstack) != 0 {
		it.focus, it.nstack = it.nstack[len(it.nstack)-1], it.nstack[:len(it.nstack)-1]
		it.idx, it.istack = it.istack[len(it.istack)-1], it.istack[:len(it.istack)-1]
		it.idx++
		if it.idx <= len(it.focus.items) {
			for it.focus.typ != Leaf {
				it.nstack = append(it.nstack, it.focus)
				it.istack = append(it.istack, it.idx)
				it.focus = (*inter)(unsafe.Pointer(it.focus)).children[it.idx]
				it.idx = 0
			}
			return
		}
	}
	it.idx = -1
	return
}

type DescendIterator struct {
	nstack    []*node
	istack    []int
	focus     *node
	stop      []byte
	idx       int
	inclusive bool
}

func (it *DescendIterator) GetKey() []byte {
	return it.focus.itemGetKey(it.idx)
}

func (it *DescendIterator) GetVal() []byte {
	return it.focus.itemGetVal(it.idx)
}

func (it *DescendIterator) Get() ([]byte, []byte) {
	return it.focus.itemGet(it.idx)
}

func (it *DescendIterator) Valid() bool {
	if len(it.stop) == 0 {
		return it.idx != -1
	} else {
		if it.inclusive {
			return it.idx != -1 && bytes.Compare(it.GetKey(), it.stop) >= 0
		} else {
			return it.idx != -1 && bytes.Compare(it.GetKey(), it.stop) > 0
		}
	}
}

func (it *DescendIterator) Next() {
	it.idx--
	if it.idx >= 0 {
		return
	}
	for len(it.nstack) != 0 {
		it.focus, it.nstack = it.nstack[len(it.nstack)-1], it.nstack[:len(it.nstack)-1]
		it.idx, it.istack = it.istack[len(it.istack)-1], it.istack[:len(it.istack)-1]
		it.idx--
		if it.idx >= 0 {
			for it.focus.typ != Leaf {
				it.nstack = append(it.nstack, it.focus)
				it.istack = append(it.istack, it.idx)
				it.focus = (*inter)(unsafe.Pointer(it.focus)).children[it.idx]
				it.idx = len(it.focus.items)
			}
			it.idx--
			return
		}
	}
	it.idx = -1
	return
}
