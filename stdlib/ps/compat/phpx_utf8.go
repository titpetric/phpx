//go:build phpx

package compat

import (
	"strings"
	"unicode/utf8"

	"github.com/titpetric/phpx/runner"
)

// pregErrorMessages are the texts preg_last_error_msg() answers.
var pregErrorMessages = []string{
	"No error", "Internal error", "Backtrack limit exhausted", "Recursion limit exhausted",
	"Malformed UTF-8 characters, possibly incorrectly encoded",
	"The offset did not correspond to the beginning of a valid UTF-8 code point", "JIT stack limit exhausted",
}

// phpxBadUTF8 reports whether a pattern with the u modifier is run on a subject
// that is not valid UTF-8, which php refuses with PREG_BAD_UTF8_ERROR, and
// records the outcome for preg_last_error().
func (c *regexpCache) phpxBadUTF8(pattern, subject string) bool {
	_, flags := splitPCRE(pattern)
	bad := strings.ContainsRune(flags, 'u') && !utf8.ValidString(subject)
	if bad {
		c.lastError.Store(4)
	} else {
		c.lastError.Store(0)
	}
	return bad
}

// phpxRegisterErrors installs preg_last_error and preg_last_error_msg over the
// state the cache keeps.
func (c *regexpCache) phpxRegisterErrors(rt *runner.Runtime) {
	rt.RegisterFunc("preg_last_error", func() int64 { return c.lastError.Load() })
	rt.RegisterFunc("preg_last_error_msg", func() string {
		n := c.lastError.Load()
		if n < 0 || int(n) >= len(pregErrorMessages) {
			return pregErrorMessages[1]
		}
		return pregErrorMessages[n]
	})
}

// phpxUCP rewrites the shorthand classes of a pattern with the u modifier to
// the Unicode properties php's PCRE2_UCP gives them: \w is a letter, a number
// or an underscore, \d a decimal digit, \s any space. It reports whether the
// pattern holds a word boundary, which RE2 judges by ASCII alone.
func phpxUCP(body string) (string, bool) {
	boundary := false
	var b strings.Builder
	b.Grow(len(body) + 16)
	inClass := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			n := body[i+1]
			i++
			switch n {
			case 'w':
				b.WriteString(map[bool]string{false: `[\p{L}\p{N}_]`, true: `\p{L}\p{N}_`}[inClass])
			case 'W':
				if inClass {
					b.WriteString(`\W`)
				} else {
					b.WriteString(`[^\p{L}\p{N}_]`)
				}
			case 'd':
				b.WriteString(`\p{Nd}`)
			case 'D':
				b.WriteString(`\P{Nd}`)
			case 's':
				b.WriteString(map[bool]string{false: `[\s\x{85}\x{a0}\p{Z}]`, true: `\s\x{85}\x{a0}\p{Z}`}[inClass])
			case 'S':
				if inClass {
					b.WriteString(`\S`)
				} else {
					b.WriteString(`[^\s\x{85}\x{a0}\p{Z}]`)
				}
			case 'b', 'B':
				if !inClass {
					boundary = true
				}
				b.WriteByte('\\')
				b.WriteByte(n)
			default:
				b.WriteByte('\\')
				b.WriteByte(n)
			}
		case c == '[' && !inClass:
			inClass = true
			b.WriteByte(c)
		case c == ']' && inClass:
			inClass = false
			b.WriteByte(c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String(), boundary
}
