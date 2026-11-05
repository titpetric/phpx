//go:build phpx

package phpx

import (
	crand "crypto/rand"
	"encoding/binary"
	"math"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Mtrand`, registerMtRand)
}

const (
	mtN = 624
	mtM = 397
	// mtRandPHP is MT_RAND_PHP: the pre-7.1 twist and range scaling.
	mtRandPHP = 1
)

// mtState is php's Mersenne Twister, state and all, so a seeded sequence is the
// sequence php prints for the same seed.
type mtState struct {
	state  [mtN]uint32
	next   int
	left   int
	legacy bool
	seeded bool
}

func (m *mtState) seed(seed uint32) {
	m.state[0] = seed
	for i := 1; i < mtN; i++ {
		m.state[i] = 1812433253*(m.state[i-1]^(m.state[i-1]>>30)) + uint32(i)
	}
	m.reload()
	m.seeded = true
}

func (m *mtState) twist(mv, u, v uint32) uint32 {
	mix := (u & 0x80000000) | (v & 0x7fffffff)
	lo := v & 1
	if m.legacy {
		lo = u & 1
	}
	return mv ^ (mix >> 1) ^ (-lo & 0x9908b0df)
}

func (m *mtState) reload() {
	s := &m.state
	p := 0
	for i := mtN - mtM; i > 0; i-- {
		s[p] = m.twist(s[p+mtM], s[p], s[p+1])
		p++
	}
	for i := mtM - 1; i > 0; i-- {
		s[p] = m.twist(s[p+mtM-mtN], s[p], s[p+1])
		p++
	}
	s[p] = m.twist(s[p+mtM-mtN], s[p], s[0])
	m.left = mtN
	m.next = 0
}

// uint32 is php_mt_rand: the next tempered 32-bit word.
func (m *mtState) uint32() uint32 {
	if !m.seeded {
		var b [4]byte
		_, _ = crand.Read(b[:])
		m.seed(binary.LittleEndian.Uint32(b[:]))
	}
	if m.left == 0 {
		m.reload()
	}
	m.left--
	s1 := m.state[m.next]
	m.next++
	s1 ^= s1 >> 11
	s1 ^= (s1 << 7) & 0x9d2c5680
	s1 ^= (s1 << 15) & 0xefc60000
	return s1 ^ (s1 >> 18)
}

func (m *mtState) uint64() uint64 {
	// php writes (php_mt_rand() << 32) | php_mt_rand() and its compiler draws
	// the right operand first, so the low word is the first draw.
	lo := uint64(m.uint32())
	return uint64(m.uint32())<<32 | lo
}

// rangeU32 is php_mt_rand_range32: uniform on [0, umax] by rejection.
func (m *mtState) rangeU32(umax uint32) uint32 {
	result := m.uint32()
	if umax == math.MaxUint32 {
		return result
	}
	umax++
	if umax&(umax-1) == 0 {
		return result & (umax - 1)
	}
	limit := math.MaxUint32 - (math.MaxUint32 % umax) - 1
	for result > limit {
		result = m.uint32()
	}
	return result % umax
}

func (m *mtState) rangeU64(umax uint64) uint64 {
	result := m.uint64()
	if umax == math.MaxUint64 {
		return result
	}
	umax++
	if umax&(umax-1) == 0 {
		return result & (umax - 1)
	}
	limit := math.MaxUint64 - (math.MaxUint64 % umax) - 1
	for result > limit {
		result = m.uint64()
	}
	return result % umax
}

// between is php_mt_rand_range: a uniform integer on [min, max].
func (m *mtState) between(min, max int64) int64 {
	umax := uint64(max) - uint64(min)
	if umax > math.MaxUint32 {
		return int64(m.rangeU64(umax) + uint64(min))
	}
	return int64(uint64(m.rangeU32(uint32(umax))) + uint64(min))
}

// legacyBetween is the RAND_RANGE_BADSCALING of MT_RAND_PHP.
func (m *mtState) legacyBetween(min, max int64) int64 {
	n := int64(m.uint32() >> 1)
	return min + int64(float64(max-min+1)*(float64(n)/(float64(math.MaxInt32)+1.0)))
}

func (m *mtState) rand(args []any) (int64, error) {
	if len(args) == 0 {
		return int64(m.uint32() >> 1), nil
	}
	if len(args) != 2 {
		return 0, &valueError{message: "mt_rand() expects exactly 2 arguments, 1 given"}
	}
	lo, hi := phpval.Int(args[0]), phpval.Int(args[1])
	if hi < lo {
		return 0, &valueError{message: "mt_rand(): Argument #2 ($max) must be greater than or equal to argument #1 ($min)"}
	}
	if m.legacy {
		return m.legacyBetween(lo, hi), nil
	}
	return m.between(lo, hi), nil
}

func registerMtRand(rt *runner.Runtime) {
	m := &mtState{}
	srand := func(args ...any) {
		m.legacy = len(args) > 1 && phpval.Int(args[1]) == mtRandPHP
		if len(args) == 0 || args[0] == nil {
			m.seeded = false
			m.uint32()
			return
		}
		m.seed(uint32(phpval.Int(args[0])))
	}
	rt.RegisterFunc("mt_srand", srand)
	rt.RegisterFunc("srand", srand)
	rt.RegisterFunc("mt_rand", func(args ...any) (int64, error) { return m.rand(args) })
	rt.RegisterFunc("rand", func(args ...any) (int64, error) {
		if len(args) == 2 && phpval.Int(args[1]) < phpval.Int(args[0]) {
			// rand() takes its bounds either way round.
			return m.rand([]any{args[1], args[0]})
		}
		return m.rand(args)
	})
	rt.RegisterFunc("str_shuffle", func(s string) string {
		b := []byte(s)
		for left := len(b) - 1; left > 0; left-- {
			j := int(m.between(0, int64(left)))
			b[left], b[j] = b[j], b[left]
		}
		return string(b)
	})
	rt.RegisterFunc("shuffle", func(arr any) bool {
		return shuffleInPlace(arr, m)
	})
	rt.RegisterFunc("array_rand", func(arr any, num ...any) (any, error) {
		return arrayRand(arr, m, num)
	})
}

// shuffleInPlace is php's shuffle: Fisher-Yates from the end, over the values,
// leaving a list.
func shuffleInPlace(arr any, m *mtState) bool {
	if !model.IsCollection(arr) {
		return false
	}
	var vals []any
	model.RangeValues(arr, func(_, v any) bool {
		vals = append(vals, v)
		return true
	})
	for left := len(vals) - 1; left > 0; left-- {
		j := int(m.between(0, int64(left)))
		vals[left], vals[j] = vals[j], vals[left]
	}
	return replaceArray(arr, vals)
}

// replaceArray rewrites an array in place as a list of vals; a Go-native value
// has no in-place rewrite and answers false.
func replaceArray(arr any, vals []any) bool {
	a, ok := arr.(*model.Array)
	if !ok {
		return false
	}
	a.Clear()
	for _, v := range vals {
		a.Append(v)
	}
	return true
}

// arrayRand is php's array_rand: one key, or n keys in the array's own order.
func arrayRand(arr any, m *mtState, num []any) (any, error) {
	if !model.IsCollection(arr) {
		return nil, &valueError{message: "array_rand(): Argument #1 ($array) must be of type array"}
	}
	var keys []any
	model.RangeValues(arr, func(k, _ any) bool {
		keys = append(keys, k)
		return true
	})
	if len(keys) == 0 {
		return nil, &valueError{message: "array_rand(): Argument #1 ($array) cannot be empty"}
	}
	n := int64(1)
	if len(num) > 0 {
		n = phpval.Int(num[0])
	}
	if n < 1 || n > int64(len(keys)) {
		return nil, &valueError{message: "array_rand(): Argument #2 ($num) must be between 1 and the number of elements in argument #1 ($array)"}
	}
	if n == 1 {
		return keys[m.between(0, int64(len(keys))-1)], nil
	}
	// php selects n of the keys by a bitset walk, then answers them in order.
	out := model.NewArray()
	need := int(n)
	negative := false
	if need > len(keys)/2 {
		negative = true
		need = len(keys) - need
	}
	picked := make([]bool, len(keys))
	for i := 0; i < need; {
		j := int(m.between(0, int64(len(keys))-1))
		if !picked[j] {
			picked[j] = true
			i++
		}
	}
	for i, k := range keys {
		if picked[i] != negative {
			out.Append(k)
		}
	}
	return out, nil
}
