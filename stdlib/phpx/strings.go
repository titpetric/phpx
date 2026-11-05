//go:build phpx

package phpx

import (
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxStrings)
}

// registerPhpxStrings installs the string and number readers with at most two
// arguments that vendor code calls.
func registerPhpxStrings(rt *runner.Runtime) {
	rt.RegisterFunc("explode", func(separator, str string, limit ...int64) ([]string, error) {
		if separator == "" {
			return nil, &valueError{message: "explode(): Argument #1 ($separator) cannot be empty"}
		}
		parts := strings.Split(str, separator)
		if len(limit) == 0 {
			return parts, nil
		}
		n := limit[0]
		switch {
		case n < 0:
			cut := int64(len(parts)) + n
			if cut < 0 {
				cut = 0
			}
			return parts[:cut], nil
		case n == 0:
			n = 1
		}
		if int64(len(parts)) > n {
			parts = append(parts[:n-1:n-1], strings.Join(parts[n-1:], separator))
		}
		return parts, nil
	})
	rt.RegisterFunc("strcspn", func(s, mask string, rest ...int64) int64 { return phpxSpan(s, mask, false, rest) })
	rt.RegisterFunc("octdec", func(s string) any {
		n, err := strconv.ParseUint(s, 8, 64)
		if err != nil {
			return int64(0)
		}
		if n > math.MaxInt64 {
			return float64(n)
		}
		return int64(n)
	})
	rt.RegisterFunc("fmod", func(x, y float64) float64 { return math.Mod(x, y) })
	rt.RegisterFunc("mb_strwidth", func(s string) int64 {
		width := int64(0)
		for _, r := range s {
			width++
			if r >= 0x1100 && (r <= 0x115f || r >= 0x2e80 && r <= 0xa4cf || r >= 0xac00 && r <= 0xd7a3 ||
				r >= 0xf900 && r <= 0xfaff || r >= 0xfe30 && r <= 0xfe6f || r >= 0xff00 && r <= 0xff60 || r >= 0xffe0 && r <= 0xffe6) {
				width++
			}
		}
		return width
	})
	rt.RegisterFunc("nl2br", func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '\r' || c == '\n' {
				b.WriteString("<br />")
				b.WriteByte(c)
				if i+1 < len(s) && (s[i+1] == '\r' || s[i+1] == '\n') && s[i+1] != c {
					i++
					b.WriteByte(s[i])
				}
				continue
			}
			b.WriteByte(c)
		}
		return b.String()
	})
	rt.RegisterFunc("mb_check_encoding", func(s string) bool { return utf8.ValidString(s) })
}

func init() {
	runner.RegisterBinding(registerPhpxEdit)
}

// registerPhpxEdit installs the edit distance between two byte strings.
func registerPhpxEdit(rt *runner.Runtime) {
	rt.RegisterFunc("levenshtein", func(a, b string) int64 {
		prev := make([]int, len(b)+1)
		for j := range prev {
			prev[j] = j
		}
		for i := 1; i <= len(a); i++ {
			cur := make([]int, len(b)+1)
			cur[0] = i
			for j := 1; j <= len(b); j++ {
				cost := 1
				if a[i-1] == b[j-1] {
					cost = 0
				}
				cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			}
			prev = cur
		}
		return int64(prev[len(b)])
	})
}

// phpxSpan is strspn (inside set) and strcspn (outside it) over the window an
// optional offset and length cut out of the subject; a negative offset counts
// from the end and a negative length stops that far from it.
func phpxSpan(s, mask string, inside bool, window []int64) int64 {
	start, end := int64(0), int64(len(s))
	if len(window) > 0 {
		start = window[0]
		if start < 0 {
			start += end
			if start < 0 {
				start = 0
			}
		}
		if start > end {
			return 0
		}
	}
	if len(window) > 1 {
		length := window[1]
		if length < 0 {
			end += length
		} else if start+length < end {
			end = start + length
		}
		if end < start {
			return 0
		}
	}
	n := int64(0)
	for i := start; i < end; i++ {
		if (strings.IndexByte(mask, s[i]) >= 0) != inside {
			break
		}
		n++
	}
	return n
}
