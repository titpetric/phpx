package compat

import (
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxNamedGroups rewrites the named capture groups of a PCRE body to plain
// ones and answers the names by group number, 1-based, "" for an unnamed group.
// PCRE numbers every group left to right; the backtracking engine numbers the
// unnamed ones first and the named after, so a pattern that mixes the two would
// shift its numbers. A back reference by name is rewritten to its number. A
// body with no named group comes back as it was and names is nil.
func phpxNamedGroups(body string, extended bool) (string, []string) {
	if !strings.Contains(body, "(?<") && !strings.Contains(body, "(?P") && !strings.Contains(body, "(?'") && !strings.Contains(body, `\k`) && !strings.Contains(body, `\g{`) {
		return body, nil
	}
	var out strings.Builder
	out.Grow(len(body))
	names := []string{""}
	named := false
	inClass := false
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			if ref, width, ok := phpxNameRef(body[i:], names); ok {
				out.WriteString(ref)
				i += width - 1
				continue
			}
			out.WriteByte(c)
			i++
			out.WriteByte(body[i])
		case inClass:
			if c == ']' {
				inClass = false
			}
			out.WriteByte(c)
		case c == '[':
			inClass = true
			out.WriteByte(c)
			if i+1 < len(body) && body[i+1] == '^' {
				i++
				out.WriteByte('^')
			}
			if i+1 < len(body) && body[i+1] == ']' {
				i++
				out.WriteByte(']')
			}
		case c == '#' && extended:
			for i < len(body) && body[i] != '\n' {
				out.WriteByte(body[i])
				i++
			}
			if i < len(body) {
				out.WriteByte('\n')
			}
		case c == '(':
			rest := body[i+1:]
			switch {
			case strings.HasPrefix(rest, "?<") && len(rest) > 2 && rest[2] != '=' && rest[2] != '!':
				if end := strings.IndexByte(rest, '>'); end > 0 {
					names = append(names, rest[2:end])
					named = true
					out.WriteByte('(')
					i += 1 + end
					continue
				}
			case strings.HasPrefix(rest, "?P<"):
				if end := strings.IndexByte(rest, '>'); end > 0 {
					names = append(names, rest[3:end])
					named = true
					out.WriteByte('(')
					i += 1 + end
					continue
				}
			case strings.HasPrefix(rest, "?'"):
				if end := strings.IndexByte(rest[2:], '\''); end > 0 {
					names = append(names, rest[2:2+end])
					named = true
					out.WriteByte('(')
					i += 2 + end + 1
					continue
				}
			case strings.HasPrefix(rest, "?P="):
				if end := strings.IndexByte(rest, ')'); end > 0 {
					if n := phpxGroupNumber(names, rest[3:end]); n > 0 {
						out.WriteString(`(?:\` + strconv.Itoa(n) + `)`)
						i += 1 + end
						continue
					}
				}
			case strings.HasPrefix(rest, "?"), strings.HasPrefix(rest, "*"):
				// A non-capturing group, a lookaround, an inline flag or a verb.
			default:
				names = append(names, "")
			}
			out.WriteByte(c)
		default:
			out.WriteByte(c)
		}
	}
	if !named {
		return out.String(), nil
	}
	return out.String(), names
}

// phpxNameRef reads `\k<name>`, `\k{name}`, `\k'name'` and `\g{name}` at the
// start of s and answers the numbered back reference.
func phpxNameRef(s string, names []string) (string, int, bool) {
	var open, closer byte
	switch {
	case strings.HasPrefix(s, `\k<`):
		open, closer = '<', '>'
	case strings.HasPrefix(s, `\k{`), strings.HasPrefix(s, `\g{`):
		open, closer = '{', '}'
	case strings.HasPrefix(s, `\k'`):
		open, closer = '\'', '\''
	default:
		return "", 0, false
	}
	end := strings.IndexByte(s[3:], closer)
	if end < 0 {
		return "", 0, false
	}
	name := s[3 : 3+end]
	if open == '{' && s[1] == 'g' {
		if n, err := strconv.Atoi(name); err == nil {
			return `\` + strconv.Itoa(n), 3 + end + 1, true
		}
	}
	n := phpxGroupNumber(names, name)
	if n == 0 {
		return "", 0, false
	}
	return `(?:\` + strconv.Itoa(n) + `)`, 3 + end + 1, true
}

func phpxGroupNumber(names []string, name string) int {
	for i := len(names) - 1; i > 0; i-- {
		if names[i] == name {
			return i
		}
	}
	return 0
}

// namedRow turns a match row into the array php returns for a pattern with
// named groups: the name precedes the number of its group.
func namedRow(row any, names []string) any {
	out := model.NewArray()
	model.RangeValues(row, func(k, v any) bool {
		g := int(k.(int64))
		if g < len(names) && names[g] != "" {
			out.Set(names[g], v)
		}
		out.Set(int64(g), v)
		return true
	})
	return out
}

// namedColumns is namedRow for PREG_PATTERN_ORDER: one column per group.
func namedColumns(cols []any, names []string) any {
	out := model.NewArray()
	for g, col := range cols {
		if g < len(names) && names[g] != "" {
			out.Set(names[g], col)
		}
		out.Set(int64(g), col)
	}
	return out
}
