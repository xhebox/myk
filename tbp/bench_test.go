package tbp

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

func BenchmarkPutDerive(b *testing.B) {
	buf := make([][128]byte, b.N)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewTree()
	j := 0

	b.ResetTimer()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
		if j == 4096 {
			p = p.Clone(NewCow(64))
			j = 0
		}
		j++
	}
}

func BenchmarkPut(b *testing.B) {
	buf := make([][128]byte, b.N)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewTree()
	b.ResetTimer()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}
}

func BenchmarkPutRandom(b *testing.B) {
	buf := make([][128]byte, b.N)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(rand.Int()))
	}

	p := NewTree()
	b.ResetTimer()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}
}

func BenchmarkGet(b *testing.B) {
	buf := make([][128]byte, b.N)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewTree()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}

	b.ResetTimer()
	for i := range buf {
		p.Get(buf[i][:16], removeItem)
	}
}

func BenchmarkGetRandom(b *testing.B) {
	buf := make([][128]byte, b.N)
	for i := range buf {
		binary.BigEndian.PutUint32(buf[i][:], uint32(rand.Int()))
	}

	p := NewTree()

	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}

	b.ResetTimer()
	for i := range buf {
		p.Get(buf[i][:16], removeItem)
	}
}

func BenchmarkPutLarge(b *testing.B) {
	buf := make([][128]byte, 10000000)
	for i := range buf {
		binary.LittleEndian.PutUint32(buf[i][:], uint32(i))
	}

	p := NewTree()

	b.ResetTimer()
	for i := range buf {
		p.Insert(buf[i][:16], buf[i][16:])
	}
}
