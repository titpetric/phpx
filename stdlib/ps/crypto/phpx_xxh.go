//go:build phpx

package crypto

import (
	"encoding/binary"
	"hash"
	"math/bits"

	"github.com/titpetric/phpx/stdlib/ps/crypto/phpx_xxh3"
)

func init() {
	digests["xxh32"] = func() hash.Hash { return &xxhBuffer{size: 4, sum: xxh32Sum} }
	digests["xxh64"] = func() hash.Hash { return &xxhBuffer{size: 8, sum: xxh64Sum} }
	digests["xxh3"] = func() hash.Hash { return &xxhBuffer{size: 8, sum: xxh3Sum} }
	digests["xxh128"] = func() hash.Hash { return &xxhBuffer{size: 16, sum: xxh128Sum} }
}

// xxhBuffer is a hash.Hash that keeps what it is given and computes the digest
// at Sum: the xxh family here is a one-shot function over the whole input,
// which is all hash(), hash_init() and hash_file() need.
type xxhBuffer struct {
	data []byte
	size int
	sum  func([]byte) []byte
}

func (h *xxhBuffer) Write(p []byte) (int, error) { h.data = append(h.data, p...); return len(p), nil }
func (h *xxhBuffer) Sum(b []byte) []byte         { return append(b, h.sum(h.data)...) }
func (h *xxhBuffer) Reset()                      { h.data = nil }
func (h *xxhBuffer) Size() int                   { return h.size }
func (h *xxhBuffer) BlockSize() int              { return 32 }

func xxh3Sum(data []byte) []byte {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], xxh3.Hash(data))
	return out[:]
}

func xxh128Sum(data []byte) []byte {
	sum := xxh3.Hash128(data)
	var out [16]byte
	binary.BigEndian.PutUint64(out[:8], sum.Hi)
	binary.BigEndian.PutUint64(out[8:], sum.Lo)
	return out[:]
}

const (
	xxP32a, xxP32b, xxP32c, xxP32d, xxP32e = 2654435761, 2246822519, 3266489917, 668265263, 374761393
	xxP64a, xxP64b, xxP64c, xxP64d, xxP64e = 11400714785074694791, 14029467366897019727, 1609587929392839161, 9650029242287828579, 2870177450012600261
)

func xxh32Sum(data []byte) []byte {
	var h uint32
	n := len(data)
	p := data
	if n >= 16 {
		p1, p2 := uint32(xxP32a), uint32(xxP32b)
		v1, v2, v3, v4 := p1+p2, p2, uint32(0), -p1
		for len(p) >= 16 {
			v1 = bits.RotateLeft32(v1+binary.LittleEndian.Uint32(p)*xxP32b, 13) * xxP32a
			v2 = bits.RotateLeft32(v2+binary.LittleEndian.Uint32(p[4:])*xxP32b, 13) * xxP32a
			v3 = bits.RotateLeft32(v3+binary.LittleEndian.Uint32(p[8:])*xxP32b, 13) * xxP32a
			v4 = bits.RotateLeft32(v4+binary.LittleEndian.Uint32(p[12:])*xxP32b, 13) * xxP32a
			p = p[16:]
		}
		h = bits.RotateLeft32(v1, 1) + bits.RotateLeft32(v2, 7) + bits.RotateLeft32(v3, 12) + bits.RotateLeft32(v4, 18)
	} else {
		h = xxP32e
	}
	h += uint32(n)
	for len(p) >= 4 {
		h = bits.RotateLeft32(h+binary.LittleEndian.Uint32(p)*xxP32c, 17) * xxP32d
		p = p[4:]
	}
	for _, b := range p {
		h = bits.RotateLeft32(h+uint32(b)*xxP32e, 11) * xxP32a
	}
	h ^= h >> 15
	h *= xxP32b
	h ^= h >> 13
	h *= xxP32c
	h ^= h >> 16
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], h)
	return out[:]
}

func xxh64Round(acc, input uint64) uint64 {
	return bits.RotateLeft64(acc+input*xxP64b, 31) * xxP64a
}

func xxh64Merge(acc, v uint64) uint64 {
	return (acc^xxh64Round(0, v))*xxP64a + xxP64d
}

func xxh64Sum(data []byte) []byte {
	var h uint64
	n := len(data)
	p := data
	if n >= 32 {
		p1, p2 := uint64(xxP64a), uint64(xxP64b)
		v1, v2, v3, v4 := p1+p2, p2, uint64(0), -p1
		for len(p) >= 32 {
			v1 = xxh64Round(v1, binary.LittleEndian.Uint64(p))
			v2 = xxh64Round(v2, binary.LittleEndian.Uint64(p[8:]))
			v3 = xxh64Round(v3, binary.LittleEndian.Uint64(p[16:]))
			v4 = xxh64Round(v4, binary.LittleEndian.Uint64(p[24:]))
			p = p[32:]
		}
		h = bits.RotateLeft64(v1, 1) + bits.RotateLeft64(v2, 7) + bits.RotateLeft64(v3, 12) + bits.RotateLeft64(v4, 18)
		h = xxh64Merge(h, v1)
		h = xxh64Merge(h, v2)
		h = xxh64Merge(h, v3)
		h = xxh64Merge(h, v4)
	} else {
		h = xxP64e
	}
	h += uint64(n)
	for len(p) >= 8 {
		h ^= xxh64Round(0, binary.LittleEndian.Uint64(p))
		h = bits.RotateLeft64(h, 27)*xxP64a + xxP64d
		p = p[8:]
	}
	if len(p) >= 4 {
		h ^= uint64(binary.LittleEndian.Uint32(p)) * xxP64a
		h = bits.RotateLeft64(h, 23)*xxP64b + xxP64c
		p = p[4:]
	}
	for _, b := range p {
		h ^= uint64(b) * xxP64e
		h = bits.RotateLeft64(h, 11) * xxP64a
	}
	h ^= h >> 33
	h *= xxP64b
	h ^= h >> 29
	h *= xxP64c
	h ^= h >> 32
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], h)
	return out[:]
}
