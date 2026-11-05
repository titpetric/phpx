//go:build phpx

package phpx

import (
	"strconv"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Sscanf`, registerSscanf)
}

// scanSpec is one conversion of an sscanf format.
type scanSpec struct {
	verb     byte
	width    int
	suppress bool
	set      string // for %[...]
	negate   bool
}

// sscanfRun reads s by the format and answers the converted values, nil for
// the ones it never reached, and how many conversions matched.
func sscanfRun(s, format string) ([]any, int, bool) {
	var out []any
	matched := 0
	si := 0
	failed := false
	for fi := 0; fi < len(format); fi++ {
		c := format[fi]
		switch {
		case c == '%':
			fi++
			if fi >= len(format) {
				break
			}
			if format[fi] == '%' {
				if !failed && si < len(s) && s[si] == '%' {
					si++
				} else {
					failed = true
				}
				continue
			}
			spec := scanSpec{}
			if format[fi] == '*' {
				spec.suppress = true
				fi++
			}
			for fi < len(format) && format[fi] >= '0' && format[fi] <= '9' {
				spec.width = spec.width*10 + int(format[fi]-'0')
				fi++
			}
			for fi < len(format) && (format[fi] == 'l' || format[fi] == 'h' || format[fi] == 'L') {
				fi++
			}
			if fi >= len(format) {
				break
			}
			spec.verb = format[fi]
			if spec.verb == '[' {
				end := fi + 1
				if end < len(format) && format[end] == '^' {
					spec.negate = true
					end++
				}
				first := end
				if end < len(format) && format[end] == ']' {
					end++
				}
				for end < len(format) && format[end] != ']' {
					end++
				}
				spec.set = expandScanSet(format[first:min(end, len(format))])
				fi = end
			}
			var val any
			if !failed {
				var ok bool
				val, si, ok = scanOne(s, si, spec)
				if !ok {
					failed = true
					val = nil
				} else if !spec.suppress && spec.verb != 'n' {
					matched++
				}
			}
			if !spec.suppress {
				out = append(out, val)
			}
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			for si < len(s) && strings.IndexByte(" \t\n\r", s[si]) >= 0 {
				si++
			}
		default:
			if !failed && si < len(s) && s[si] == c {
				si++
			} else {
				failed = true
			}
		}
	}
	return out, matched, si >= len(s) || len(out) > 0
}

func expandScanSet(set string) string {
	var b strings.Builder
	for i := 0; i < len(set); i++ {
		if i+2 < len(set) && set[i+1] == '-' {
			for c := int(set[i]); c <= int(set[i+2]); c++ {
				b.WriteByte(byte(c))
			}
			i += 2
			continue
		}
		b.WriteByte(set[i])
	}
	return b.String()
}

func scanOne(s string, si int, spec scanSpec) (any, int, bool) {
	skipSpace := func() {
		for si < len(s) && strings.IndexByte(" \t\n\r", s[si]) >= 0 {
			si++
		}
	}
	limit := func(start int) int {
		if spec.width > 0 {
			return min(len(s), start+spec.width)
		}
		return len(s)
	}
	switch spec.verb {
	case 'n':
		return int64(si), si, true
	case 'c':
		w := max(spec.width, 1)
		if si+w > len(s) {
			return nil, si, false
		}
		return s[si : si+w], si + w, true
	case 's':
		skipSpace()
		start, end := si, limit(si)
		for si < end && strings.IndexByte(" \t\n\r", s[si]) < 0 {
			si++
		}
		if si == start {
			return nil, si, false
		}
		return s[start:si], si, true
	case '[':
		start, end := si, limit(si)
		for si < end && (strings.IndexByte(spec.set, s[si]) >= 0) != spec.negate {
			si++
		}
		if si == start {
			return nil, si, false
		}
		return s[start:si], si, true
	case 'd', 'i', 'u', 'x', 'X', 'o':
		skipSpace()
		start, end := si, limit(si)
		if si < end && (s[si] == '-' || s[si] == '+') {
			si++
		}
		base := 10
		switch spec.verb {
		case 'x', 'X':
			base = 16
			if si+1 < end && s[si] == '0' && (s[si+1] == 'x' || s[si+1] == 'X') {
				si += 2
			}
		case 'o':
			base = 8
		}
		digits := si
		for si < end && digitOf(s[si], base) {
			si++
		}
		if si == digits {
			return nil, start, false
		}
		text := s[start:si]
		text = strings.Replace(strings.Replace(text, "0x", "", 1), "0X", "", 1)
		n, err := strconv.ParseInt(text, base, 64)
		if err != nil {
			return nil, start, false
		}
		return n, si, true
	case 'f', 'e', 'E', 'g':
		skipSpace()
		start, end := si, limit(si)
		if si < end && (s[si] == '-' || s[si] == '+') {
			si++
		}
		seen := false
		for si < end && (s[si] >= '0' && s[si] <= '9' || s[si] == '.' || s[si] == 'e' || s[si] == 'E' || ((s[si] == '-' || s[si] == '+') && (s[si-1] == 'e' || s[si-1] == 'E'))) {
			seen = true
			si++
		}
		if !seen {
			return nil, start, false
		}
		f, err := strconv.ParseFloat(s[start:si], 64)
		if err != nil {
			return nil, start, false
		}
		return f, si, true
	}
	return nil, si, false
}

func digitOf(c byte, base int) bool {
	switch base {
	case 8:
		return c >= '0' && c <= '7'
	case 16:
		return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
	}
	return c >= '0' && c <= '9'
}

func registerSscanf(rt *runner.Runtime) {
	rt.RegisterFunc("sscanf", func(s, format string, refs ...any) any {
		vals, matched, _ := sscanfRun(s, format)
		if len(refs) == 0 {
			if len(vals) > 0 && matched == 0 && s == "" {
				return int64(-1)
			}
			out := model.NewArray()
			for _, v := range vals {
				out.Append(v)
			}
			return out
		}
		for i, ref := range refs {
			if set, ok := ref.(func(any)); ok && i < len(vals) {
				set(vals[i])
			}
		}
		if s == "" {
			return int64(-1)
		}
		return int64(matched)
	})
	rt.RegisterFunc("str_ireplace", func(search, replace, subject any, rest ...any) any {
		total := int64(0)
		one := func(text string) string {
			searches := phpxStringsOf(search)
			replaces := phpxStringsOf(replace)
			listRepl := model.IsCollection(replace)
			for i, from := range searches {
				to := ""
				switch {
				case !listRepl:
					to = phpval.String(replace)
				case i < len(replaces):
					to = replaces[i]
				}
				if from == "" {
					continue
				}
				lower, lowerFrom := strings.ToLower(text), strings.ToLower(from)
				var b strings.Builder
				pos := 0
				for {
					j := strings.Index(lower[pos:], lowerFrom)
					if j < 0 {
						break
					}
					b.WriteString(text[pos : pos+j])
					b.WriteString(to)
					pos += j + len(from)
					total++
				}
				b.WriteString(text[pos:])
				text = b.String()
				lower = strings.ToLower(text)
			}
			return text
		}
		var out any
		if model.IsCollection(subject) {
			res := model.NewArray()
			model.RangeValues(subject, func(k, v any) bool {
				res.Set(k, one(phpval.String(v)))
				return true
			})
			out = res
		} else {
			out = one(phpval.String(subject))
		}
		if len(rest) > 0 {
			if set, ok := rest[0].(func(any)); ok {
				set(total)
			}
		}
		return out
	})
}

func phpxStringsOf(v any) []string {
	var out []string
	if model.IsCollection(v) {
		model.RangeValues(v, func(_, item any) bool {
			out = append(out, phpval.String(item))
			return true
		})
		return out
	}
	return []string{phpval.String(v)}
}
