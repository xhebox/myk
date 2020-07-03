package tbp

import (
	"unsafe"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuiteUtil struct {
	suite.Suite
	*require.Assertions
}

func (t *TestSuiteUtil) BeforeTest(suiteName, testName string) {
	t.Assertions = t.Suite.Require()
}

func (t *TestSuiteUtil) CheckType(n interface{}, typ nodeType) {
	switch v := n.(type) {
	case *node:
		t.Equal(typ, v.typ)
	case *leaf:
		t.Equal(typ, v.typ)
	case *inter:
		t.Equal(typ, v.typ)
	default:
		t.Fail("invalid type")
	}
}

func (t *TestSuiteUtil) CheckItemLen(n interface{}, l, c int) {
	switch v := n.(type) {
	case *node:
		t.Equal(l, len(v.items))
		t.Equal(c, cap(v.items))
	case *leaf:
		t.Equal(l, len(v.items))
		t.Equal(c, cap(v.items))
	case *inter:
		t.Equal(l, len(v.items))
		t.Equal(c, cap(v.items))
	default:
		t.Fail("invalid type")
	}
}

func (t *TestSuiteUtil) CheckInterLen(n interface{}, l, c int) {
	switch v := n.(type) {
	case *inter:
		t.Equal(l, len(v.children))
		t.Equal(c, cap(v.children))
	case *node:
		if v.typ != Inter {
			t.Fail("invalid type")
		}
		m := (*inter)(unsafe.Pointer(v))
		t.Equal(l, len(m.children))
		t.Equal(c, cap(m.children))
	default:
		t.Fail("invalid type")
	}
}

func (t *TestSuiteUtil) CheckSlice(n, key, val []byte) {
	buf := make([]byte, 4+len(key)+len(val))
	putUint32(buf, uint32(len(key)))
	copy(buf[4:], key)
	copy(buf[4+len(key):], val)
	t.Equal(buf, n)
}

func (t *TestSuiteUtil) CheckItem(n interface{}, i int, key, val []byte) {
	switch v := n.(type) {
	case *node:
		t.CheckSlice(v.items[i], key, val)
	case *leaf:
		t.CheckSlice(v.items[i], key, val)
	case *inter:
		t.CheckSlice(v.items[i], key, val)
	default:
		t.Fail("invalid type")
	}
}
