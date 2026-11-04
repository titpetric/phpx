package compat

import (
	"regexp"
	"strings"
)

// phpxEscapes rewrites the PCRE escapes RE2 has no spelling for into ones it
// has: \R (any line break) and \h and \H (horizontal space), outside and inside
// a character class. The pattern is otherwise left as it came.
func phpxEscapes(body string) string {
	body = phpxPosixClasses(body)
	body = phpxPossessive(body)
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

// possessive drops the possessive marker after a character class: `[a-z]++` and
// `[a-z]*+` read as the greedy forms, which match the same text where a class
// is followed by something it cannot also match. Neither engine has the syntax.
//
// A group takes the same reading: `(?:ab)++` is `(?:ab)+`. Backtracking into
// the group is the only difference and a pattern written for the possessive
// form does not rely on it failing.
var possessive = strings.NewReplacer("]++", "]+", "]*+", "]*", "]?+", "]?", ")++", ")+", ")*+", ")*", ")?+", ")?")

// possessiveCount is a counted quantifier made possessive: `{2}+`, `{1,3}+`.
var possessiveCount = regexp.MustCompile(`(\{\d+(?:,\d*)?\})\+`)

func phpxPossessive(body string) string {
	if !strings.Contains(body, "+") {
		return body
	}
	body = possessive.Replace(body)
	return possessiveCount.ReplaceAllString(body, "$1")
}
