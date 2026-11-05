//go:build phpx

package phpx

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Unpack`, registerPhpxUnpack)
}

// registerPhpxUnpack installs unpack() for the formats libraries use to read
// bytes: C c n v N V J P a A Z H h, with a repeat count or *, and a name.
func registerPhpxUnpack(rt *runner.Runtime) {
	rt.RegisterFunc("unpack", func(format, data string, offset ...int64) (any, error) {
		if len(offset) > 0 && offset[0] > 0 && int(offset[0]) <= len(data) {
			data = data[offset[0]:]
		}
		out := model.NewArray()
		for _, part := range strings.Split(format, "/") {
			if part == "" {
				continue
			}
			code := part[0]
			rest := part[1:]
			count, star := 1, false
			digits := 0
			for digits < len(rest) && rest[digits] >= '0' && rest[digits] <= '9' {
				digits++
			}
			hasCount := digits > 0
			if digits > 0 {
				count, _ = strconv.Atoi(rest[:digits])
			}
			rest = rest[digits:]
			if strings.HasPrefix(rest, "*") {
				star, rest = true, rest[1:]
			}
			name := rest
			size := unpackSizes[code]
			switch code {
			case 'a', 'A', 'Z', 'H', 'h':
				n := count
				if code == 'H' || code == 'h' {
					if star {
						n = len(data) * 2
					}
					if n > len(data)*2 {
						return false, nil
					}
					b := data[:(n+1)/2]
					data = data[(n+1)/2:]
					s := hex.EncodeToString([]byte(b))
					if code == 'h' {
						r := []byte(s)
						for i := 0; i+1 < len(r); i += 2 {
							r[i], r[i+1] = r[i+1], r[i]
						}
						s = string(r)
					}
					unpackSet(out, name, 1, 1, s[:n])
					continue
				}
				if star || (!hasCount && false) {
					n = len(data)
				}
				if n > len(data) {
					return false, nil
				}
				s := data[:n]
				data = data[n:]
				switch code {
				case 'A':
					s = strings.TrimRight(s, " \t\r\n\x00")
				case 'Z':
					if i := strings.IndexByte(s, 0); i >= 0 {
						s = s[:i]
					}
				}
				unpackSet(out, name, 1, 1, s)
			default:
				if size == 0 {
					return nil, fmt.Errorf("unpack(): Type %c: unknown format code", code)
				}
				if star {
					count = len(data) / size
				}
				for i := 0; i < count; i++ {
					if len(data) < size {
						return false, nil
					}
					b := []byte(data[:size])
					data = data[size:]
					var v any
					switch code {
					case 's', 'S', 'l', 'L', 'q', 'Q', 'i', 'I', 'f', 'd', 'e', 'E', 'g', 'G':
						v = unpackNative(code, b)
					case 'C':
						v = int64(b[0])
					case 'c':
						v = int64(int8(b[0]))
					case 'n':
						v = int64(binary.BigEndian.Uint16(b))
					case 'v':
						v = int64(binary.LittleEndian.Uint16(b))
					case 'N':
						v = int64(binary.BigEndian.Uint32(b))
					case 'V':
						v = int64(binary.LittleEndian.Uint32(b))
					case 'J':
						v = int64(binary.BigEndian.Uint64(b))
					case 'P':
						v = int64(binary.LittleEndian.Uint64(b))
					}
					unpackSet(out, name, i+1, count, v)
				}
			}
		}
		return out, nil
	})
}

// unpackSet stores one value under php's key: the name, the name and the
// position when it repeats, or the 1-based position with no name.
func unpackSet(out *model.Array, name string, pos, count int, v any) {
	switch {
	case name == "":
		out.Set(int64(pos), v)
	case count > 1:
		out.Set(name+strconv.Itoa(pos), v)
	default:
		out.Set(name, v)
	}
}

// unpackSizes is the byte width of each fixed-size format code.
var unpackSizes = map[byte]int{
	'C': 1, 'c': 1, 'n': 2, 'v': 2, 'N': 4, 'V': 4, 'J': 8, 'P': 8,
	's': 2, 'S': 2, 'l': 4, 'L': 4, 'q': 8, 'Q': 8, 'i': 4, 'I': 4,
	'f': 4, 'd': 8, 'e': 8, 'E': 8, 'g': 4, 'G': 4,
}

// unpackNative reads a machine-order integer or a float of either order.
func unpackNative(code byte, b []byte) any {
	switch code {
	case 's':
		return int64(int16(binary.LittleEndian.Uint16(b)))
	case 'S':
		return int64(binary.LittleEndian.Uint16(b))
	case 'l', 'i':
		return int64(int32(binary.LittleEndian.Uint32(b)))
	case 'L', 'I':
		return int64(binary.LittleEndian.Uint32(b))
	case 'q', 'Q':
		return int64(binary.LittleEndian.Uint64(b))
	case 'f':
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(b)))
	case 'g':
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(b)))
	case 'G':
		return float64(math.Float32frombits(binary.BigEndian.Uint32(b)))
	case 'd', 'e':
		return math.Float64frombits(binary.LittleEndian.Uint64(b))
	case 'E':
		return math.Float64frombits(binary.BigEndian.Uint64(b))
	}
	return nil
}
