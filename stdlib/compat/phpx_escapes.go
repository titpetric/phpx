package compat

import (
	"strings"
)

// phpxEscapes rewrites the PCRE escapes RE2 has no spelling for into ones it
// has: \R (any line break) and \h and \H (horizontal space), outside and inside
// a character class. The pattern is otherwise left as it came.
func phpxEscapes(body string, extended bool) string {
	body = phpxPosixClasses(body)
	body = phpxPossessive(body, extended)
	if !strings.Contains(body, `\R`) && !strings.Contains(body, `\h`) && !strings.Contains(body, `\H`) {
		return body
	}
	var b strings.Builder
	b.Grow(len(body) + 16)
	inClass := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			n := body[i+1]
			i++
			switch {
			case n == 'R' && !inClass:
				b.WriteString(`(?:\r\n|\n|\r)`)
			case n == 'h' && inClass:
				b.WriteString(` \t`)
			case n == 'h':
				b.WriteString(`[ \t]`)
			case n == 'H' && !inClass:
				b.WriteString(`[^ \t]`)
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
	return b.String()
}

// posixClasses spell the POSIX bracket classes as the ranges a character class
// holds. The backtracking engine has no [:name:] syntax, and a pattern with a
// lookbehind reaches it.
var posixClasses = strings.NewReplacer(
	"[:alpha:]", `a-zA-Z`, "[:digit:]", `0-9`, "[:alnum:]", `a-zA-Z0-9`,
	"[:upper:]", `A-Z`, "[:lower:]", `a-z`, "[:space:]", `\s`, "[:blank:]", ` \t`,
	"[:word:]", `\w`, "[:xdigit:]", `0-9A-Fa-f`, "[:cntrl:]", `\x00-\x1f\x7f`,
	"[:print:]", `\x20-\x7e`, "[:graph:]", `\x21-\x7e`,
	"[:punct:]", `!-\/:-@\[-`+"`"+`{-~`,
)

// phpxPosixClasses rewrites every [:name:] a pattern holds.
func phpxPosixClasses(body string) string {
	if !strings.Contains(body, "[:") {
		return body
	}
	return posixClasses.Replace(body)
}

// phpxPossessive rewrites a possessive quantifier, `\w*+`, `[a-z]++`,
// `(?:ab)?+` and `{2}+`, to the atomic group that means the same, `(?>\w*)`,
// which the backtracking engine has and the linear one does not (the pattern
// goes to the backtracking engine because of it).
//
// The scan keeps the start of the last atom: an escape, a class, a literal, or
// a group from its opening parenthesis to its closing one. It skips an escaped
// character and the inside of a class, so a literal `\++` and `[+]+` keep their
// plus, and with the x flag a `#` comment to the end of the line.
func phpxPossessive(body string, extended bool) string {
	if !strings.Contains(body, "+") {
		return body
	}
	out := make([]byte, 0, len(body)+16)
	var groups []int
	atom := -1
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			atom = len(out)
			out = append(out, c)
			i++
			out = append(out, body[i])
			if (body[i] == 'p' || body[i] == 'P') && i+1 < len(body) && body[i+1] == '{' {
				for i+1 < len(body) && body[i] != '}' {
					i++
					out = append(out, body[i])
				}
			}
		case c == '[':
			atom = len(out)
			out = append(out, c)
			if i+1 < len(body) && body[i+1] == '^' {
				i++
				out = append(out, '^')
			}
			if i+1 < len(body) && body[i+1] == ']' {
				i++
				out = append(out, ']')
			}
			for i+1 < len(body) && body[i+1] != ']' {
				i++
				out = append(out, body[i])
				if body[i] == '\\' && i+1 < len(body) {
					i++
					out = append(out, body[i])
				}
			}
			if i+1 < len(body) {
				i++
				out = append(out, ']')
			}
		case c == '#' && extended:
			for i < len(body) && body[i] != '\n' {
				out = append(out, body[i])
				i++
			}
			if i < len(body) {
				out = append(out, '\n')
			}
		case c == '(':
			groups = append(groups, len(out))
			atom = -1
			out = append(out, c)
		case c == ')':
			out = append(out, c)
			if n := len(groups); n > 0 {
				atom = groups[n-1]
				groups = groups[:n-1]
			}
		case c == '*' || c == '+' || c == '?' || c == '{' && phpxCountAt(body, i) > 0:
			if c == '?' && i > 0 && body[i-1] == '(' {
				out = append(out, c)
				continue
			}
			end := i
			if c == '{' {
				end = i + phpxCountAt(body, i) - 1
			}
			out = append(out, body[i:end+1]...)
			i = end
			if i+1 < len(body) && body[i+1] == '+' && atom >= 0 {
				i++
				quantified := append([]byte("(?>"), out[atom:]...)
				quantified = append(quantified, ')')
				out = append(out[:atom], quantified...)
			}
		default:
			atom = len(out)
			out = append(out, c)
		}
	}
	return string(out)
}

// phpxCountAt answers the length of a counted quantifier, `{2}` or `{1,3}`,
// that starts at body[i], or 0 when the brace starts none.
func phpxCountAt(body string, i int) int {
	j := i + 1
	digits := 0
	for j < len(body) && body[j] >= '0' && body[j] <= '9' {
		j++
		digits++
	}
	if digits == 0 {
		return 0
	}
	if j < len(body) && body[j] == ',' {
		j++
		for j < len(body) && body[j] >= '0' && body[j] <= '9' {
			j++
		}
	}
	if j < len(body) && body[j] == '}' {
		return j - i + 1
	}
	return 0
}
