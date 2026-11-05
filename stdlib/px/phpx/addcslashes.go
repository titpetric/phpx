//go:build phpx

package phpx

import "strings"

// addCSlashes is addcslashes(): the characters of the list, a range written
// "a..z" among them, get a backslash, and the unprintable ones are written as
// C does (\n, \t, \r, \a, \v, \b, \f, else \ooo).
func addCSlashes(s, chars string) string {
	var set [256]bool
	for i := 0; i < len(chars); i++ {
		if i+3 < len(chars) && chars[i+1] == '.' && chars[i+2] == '.' && chars[i+3] >= chars[i] {
			for c := int(chars[i]); c <= int(chars[i+3]); c++ {
				set[c] = true
			}
			i += 3
			continue
		}
		set[chars[i]] = true
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !set[c] {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('\\')
		if c < 32 || c > 126 {
			switch c {
			case '\n':
				b.WriteByte('n')
			case '\t':
				b.WriteByte('t')
			case '\r':
				b.WriteByte('r')
			case 7:
				b.WriteByte('a')
			case 11:
				b.WriteByte('v')
			case 8:
				b.WriteByte('b')
			case 12:
				b.WriteByte('f')
			default:
				b.WriteByte('0' + c>>6)
				b.WriteByte('0' + (c>>3)&7)
				b.WriteByte('0' + c&7)
			}
			continue
		}
		b.WriteByte(c)
	}
	return b.String()
}
