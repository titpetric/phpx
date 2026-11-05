package compat

import (
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
)

// groupMap relates the groups the engines number to the groups php numbers.
//
// PCRE numbers every group left to right; the backtracking engine numbers the
// unnamed ones first and the named after, and neither knows `(?|` where each
// alternative starts counting again. The pattern is rewritten so that the
// engines see plain groups in source order (the physical numbers) and a match is
// read back through logical, one number per group as php counts them.
type groupMap struct {
	names   []string // by logical number, "" for an unnamed group
	logical []int    // physical number -> logical number, -1 for a dropped group
	skips   []int    // physical numbers of the (*SKIP)(*F) groups
	total   int      // logical groups, not counting group 0
}

// apply reads a match through the map: the row php sees, and whether a skip
// verb discards the match.
func (m *groupMap) apply(idx []int) ([]int, bool) {
	out := make([]int, 2*(m.total+1))
	for i := range out {
		out[i] = -1
	}
	for phys := 0; 2*phys+1 < len(idx); phys++ {
		l := -1
		if phys < len(m.logical) {
			l = m.logical[phys]
		}
		if idx[2*phys] < 0 {
			continue
		}
		if l < 0 {
			return nil, true
		}
		if out[2*l] < 0 {
			out[2*l], out[2*l+1] = idx[2*phys], idx[2*phys+1]
		}
	}
	return out, false
}

// phpxGroupMap rewrites the named groups, the branch resets and the skip groups
// of a PCRE body to plain groups and answers how their numbers map back. A body
// with none of them comes back as it was with a nil map. A back reference by
// name is rewritten to its number.
func phpxGroupMap(body string, extended bool) (string, *groupMap) {
	if !strings.Contains(body, "(?<") && !strings.Contains(body, "(?P") && !strings.Contains(body, "(?'") && !strings.Contains(body, "(?n)") &&
		!strings.Contains(body, `\k`) && !strings.Contains(body, `\g{`) && !strings.Contains(body, "(?|") {
		return body, nil
	}
	// (?J) lets two groups share a name; the groups are numbered apart here
	// already, so the flag has nothing left to do and neither engine has it.
	body = strings.ReplaceAll(body, "(?J)", "")
	// (?n) turns off the capture of unnamed groups from there on.
	autoOff := strings.Index(body, "(?n)")
	body = strings.ReplaceAll(body, "(?n)", "")
	type frame struct {
		reset             bool
		start, maxLogical int
	}
	var out strings.Builder
	out.Grow(len(body))
	m := &groupMap{names: []string{""}, logical: []int{0}}
	physNames := []string{""}
	changed := false
	logical := 0
	var frames []frame
	inClass := false
	open := func(name string) {
		physNames = append(physNames, name)
		if strings.HasPrefix(name, phpxSkipPrefix) {
			m.logical = append(m.logical, -1)
			m.skips = append(m.skips, len(physNames)-1)
			changed = true
			return
		}
		logical++
		m.logical = append(m.logical, logical)
		for len(m.names) <= logical {
			m.names = append(m.names, "")
		}
		if name != "" {
			m.names[logical] = name
			changed = true
		}
	}
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			if ref, width, ok := phpxNameRef(body[i:], physNames, m.logical); ok {
				out.WriteString(ref)
				i += width - 1
				changed = true
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
		case c == '|':
			if n := len(frames); n > 0 && frames[n-1].reset {
				if logical > frames[n-1].maxLogical {
					frames[n-1].maxLogical = logical
				}
				logical = frames[n-1].start
			}
			out.WriteByte(c)
		case c == ')':
			if n := len(frames); n > 0 {
				if frames[n-1].reset && frames[n-1].maxLogical > logical {
					logical = frames[n-1].maxLogical
				}
				frames = frames[:n-1]
			}
			out.WriteByte(c)
		case c == '(':
			rest := body[i+1:]
			switch {
			case strings.HasPrefix(rest, "?|"):
				frames = append(frames, frame{reset: true, start: logical, maxLogical: logical})
				out.WriteString("(?:")
				i += 2
				changed = true
				continue
			case strings.HasPrefix(rest, "?<") && len(rest) > 2 && rest[2] != '=' && rest[2] != '!':
				if end := strings.IndexByte(rest, '>'); end > 0 {
					open(rest[2:end])
					frames = append(frames, frame{})
					out.WriteByte('(')
					i += 1 + end
					continue
				}
			case strings.HasPrefix(rest, "?P<"):
				if end := strings.IndexByte(rest, '>'); end > 0 {
					open(rest[3:end])
					frames = append(frames, frame{})
					out.WriteByte('(')
					i += 1 + end
					continue
				}
			case strings.HasPrefix(rest, "?'"):
				if end := strings.IndexByte(rest[2:], '\''); end > 0 {
					open(rest[2 : 2+end])
					frames = append(frames, frame{})
					out.WriteByte('(')
					i += 2 + end + 1
					continue
				}
			case strings.HasPrefix(rest, "?P="):
				if end := strings.IndexByte(rest, ')'); end > 0 {
					if n := phpxGroupNumber(physNames, rest[3:end]); n > 0 {
						out.WriteString(`(?:\` + strconv.Itoa(n) + `)`)
						i += 1 + end
						changed = true
						continue
					}
				}
			case strings.HasPrefix(rest, "?"), strings.HasPrefix(rest, "*"):
				// A non-capturing group, a lookaround, an inline flag or a verb.
			default:
				if autoOff >= 0 && i >= autoOff {
					frames = append(frames, frame{})
					out.WriteString("(?:")
					changed = true
					continue
				}
				open("")
				frames = append(frames, frame{})
				out.WriteByte('(')
				continue
			}
			frames = append(frames, frame{})
			out.WriteByte(c)
		default:
			out.WriteByte(c)
		}
	}
	if !changed {
		return out.String(), nil
	}
	m.total = logical
	for len(m.names) <= m.total {
		m.names = append(m.names, "")
	}
	named := false
	for _, n := range m.names {
		if n != "" {
			named = true
		}
	}
	if !named {
		m.names = nil
	}
	return out.String(), m
}

// phpxNameRef reads `\k<name>`, `\k{name}`, `\k'name'` and `\g{name}` at the
// start of s and answers the numbered back reference.
func phpxNameRef(s string, names []string, _ []int) (string, int, bool) {
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
//
// Two groups may share a name under (?J): the one that matched keeps the name,
// and an unmatched one only claims it when no other did, as php's add_named does.
func namedRow(row any, names []string, idx []int) any {
	out := model.NewArray()
	model.RangeValues(row, func(k, v any) bool {
		g := int(k.(int64))
		if g < len(names) && names[g] != "" {
			matched := 2*g+1 < len(idx) && idx[2*g] >= 0
			if _, taken := out.Get(names[g]); matched || !taken {
				out.Set(names[g], v)
			}
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
			if _, taken := out.Get(names[g]); !taken {
				out.Set(names[g], col)
			}
		}
		out.Set(int64(g), col)
	}
	return out
}
