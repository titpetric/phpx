//go:build phpx

package phpx

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPack)
}

func registerPack(rt *runner.Runtime) {
	rt.RegisterFunc("pack", func(format string, args ...any) (any, error) {
		return pack(format, args)
	})
}

// pack builds a binary string from php's format codes: a A Z H h, the integer
// codes c C s S n v l L N V q Q J P i I, the float codes f g G d e E, x (a NUL),
// X (back up one byte) and @ (pad to an offset), each with a repeat count or *.
func pack(format string, args []any) (string, error) {
	var out []byte
	next := 0
	take := func(code byte) (any, error) {
		if next >= len(args) {
			return nil, fmt.Errorf("pack(): Type %c: too few arguments", code)
		}
		v := args[next]
		next++
		return v, nil
	}
	for i := 0; i < len(format); {
		code := format[i]
		i++
		count, star, hasCount := 1, false, false
		if i < len(format) && format[i] == '*' {
			star = true
			i++
		} else {
			start := i
			for i < len(format) && format[i] >= '0' && format[i] <= '9' {
				i++
			}
			if i > start {
				count, _ = strconv.Atoi(format[start:i])
				hasCount = true
			}
		}
		switch code {
		case 'a', 'A', 'Z':
			v, err := take(code)
			if err != nil {
				return "", err
			}
			s := phpval.String(v)
			n := count
			if star {
				n = len(s)
				if code == 'Z' {
					n++
				}
			}
			pad := byte(0)
			if code == 'A' {
				pad = ' '
			}
			b := make([]byte, n)
			for j := range b {
				b[j] = pad
			}
			limit := n
			if code == 'Z' && n > 0 {
				limit = n - 1
			}
			copy(b, s[:min(len(s), limit)])
			out = append(out, b...)
		case 'H', 'h':
			v, err := take(code)
			if err != nil {
				return "", err
			}
			s := phpval.String(v)
			if star || !hasCount {
				if star {
					count = len(s)
				}
			}
			if count > len(s) {
				count = len(s)
			}
			s = s[:count]
			if len(s)%2 == 1 {
				s += "0"
			}
			if code == 'h' {
				r := []byte(s)
				for j := 0; j+1 < len(r); j += 2 {
					r[j], r[j+1] = r[j+1], r[j]
				}
				s = string(r)
			}
			b, err := hex.DecodeString(s)
			if err != nil {
				return "", fmt.Errorf("pack(): Type %c: illegal hex digit", code)
			}
			out = append(out, b...)
		case 'x':
			if star {
				count = 0
			}
			out = append(out, make([]byte, count)...)
		case 'X':
			if star {
				count = 0
			}
			out = out[:max(0, len(out)-count)]
		case '@':
			if count > len(out) {
				out = append(out, make([]byte, count-len(out))...)
			} else {
				out = out[:count]
			}
		case 'c', 'C', 's', 'S', 'n', 'v', 'l', 'L', 'N', 'V', 'q', 'Q', 'J', 'P', 'i', 'I',
			'f', 'g', 'G', 'd', 'e', 'E':
			if star {
				count = len(args) - next
			}
			for j := 0; j < count; j++ {
				v, err := take(code)
				if err != nil {
					return "", err
				}
				out = packOne(out, code, v)
			}
		default:
			return "", fmt.Errorf("pack(): Type %c: unknown format code", code)
		}
	}
	if next < len(args) {
		return "", fmt.Errorf("pack(): %d arguments unused", len(args)-next)
	}
	return string(out), nil
}

// packOne appends one value in the width and byte order the code names.
func packOne(out []byte, code byte, v any) []byte {
	switch code {
	case 'f', 'g', 'G', 'd', 'e', 'E':
		f := phpval.Float(v)
		switch code {
		case 'f', 'g':
			return binary.LittleEndian.AppendUint32(out, math.Float32bits(float32(f)))
		case 'G':
			return binary.BigEndian.AppendUint32(out, math.Float32bits(float32(f)))
		case 'E':
			return binary.BigEndian.AppendUint64(out, math.Float64bits(f))
		}
		return binary.LittleEndian.AppendUint64(out, math.Float64bits(f))
	}
	n := uint64(phpval.Int(v))
	switch code {
	case 'c', 'C':
		return append(out, byte(n))
	case 'n':
		return binary.BigEndian.AppendUint16(out, uint16(n))
	case 's', 'S', 'v':
		return binary.LittleEndian.AppendUint16(out, uint16(n))
	case 'N':
		return binary.BigEndian.AppendUint32(out, uint32(n))
	case 'l', 'L', 'V', 'i', 'I':
		return binary.LittleEndian.AppendUint32(out, uint32(n))
	case 'J':
		return binary.BigEndian.AppendUint64(out, n)
	}
	return binary.LittleEndian.AppendUint64(out, n)
}
