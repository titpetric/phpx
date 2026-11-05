package compat

import (
	"strconv"
	"strings"
)

// maxSubroutineDepth bounds the expansion of a pattern that calls itself.
// `\{ ([^{}] | (?R))* \}` matches braces nested this deep and no deeper, which
// is what the recursion is used for in practice.
const maxSubroutineDepth = 4

// pcreGroup is one capturing group of a pattern body, as indexes into it.
type pcreGroup struct {
	open, inner, end int // '(' , first byte of the contents, the ')' byte
	num              int
	name             string
}

// phpxSubroutines expands what neither engine has: `(?&name)`, `(?P>name)`,
// `(?1)`, `(?-1)`, `(?R)` and `(?(DEFINE)...)`. A call becomes the group's own
// text in a non-capturing group, so the numbering of the pattern's groups is the
// one PCRE gives it. A group defined in DEFINE keeps its place in the numbering.
// A call that reaches itself is expanded to maxSubroutineDepth levels and fails
// to match beyond them.
func phpxSubroutines(body string, extended bool) string {
	if !strings.Contains(body, "(?&") && !strings.Contains(body, "(?P>") && !strings.Contains(body, "(?R)") &&
		!strings.Contains(body, "(?(DEFINE)") && !phpxHasNumberedCall(body) {
		return body
	}
	s := &subExpander{body: body, extended: extended}
	s.scan()
	return s.render(0, len(body), 0, false)
}

func phpxHasNumberedCall(body string) bool {
	for i := 0; i+3 < len(body); i++ {
		if body[i] == '(' && body[i+1] == '?' {
			j := i + 2
			if body[j] == '-' || body[j] == '+' {
				j++
			}
			k := j
			for k < len(body) && body[k] >= '0' && body[k] <= '9' {
				k++
			}
			if k > j && k < len(body) && body[k] == ')' {
				return true
			}
		}
	}
	return false
}

type subExpander struct {
	body     string
	extended bool
	groups   []pcreGroup
}

// skipOpaque answers the index after an escape, a class or an x-mode comment
// that starts at i, or i when none does.
func (s *subExpander) skipOpaque(i int) int {
	b := s.body
	switch {
	case b[i] == '\\' && i+1 < len(b):
		return i + 2
	case b[i] == '[':
		j := i + 1
		if j < len(b) && b[j] == '^' {
			j++
		}
		if j < len(b) && b[j] == ']' {
			j++
		}
		for j < len(b) && b[j] != ']' {
			if b[j] == '\\' {
				j++
			}
			j++
		}
		return j + 1
	case b[i] == '#' && s.extended:
		j := i
		for j < len(b) && b[j] != '\n' {
			j++
		}
		return j
	case b[i] == '(' && strings.HasPrefix(b[i:], "(?#"):
		j := strings.IndexByte(b[i:], ')')
		if j < 0 {
			return len(b)
		}
		return i + j + 1
	}
	return i
}

// opener reads the group syntax at b[i] == '(': its length, whether it
// captures and its name.
func (s *subExpander) opener(i int) (n int, capture bool, name string) {
	rest := s.body[i+1:]
	switch {
	case strings.HasPrefix(rest, "?<") && len(rest) > 2 && rest[2] != '=' && rest[2] != '!':
		if e := strings.IndexByte(rest, '>'); e > 0 {
			return 1 + e + 1, true, rest[2:e]
		}
	case strings.HasPrefix(rest, "?P<"):
		if e := strings.IndexByte(rest, '>'); e > 0 {
			return 1 + e + 1, true, rest[3:e]
		}
	case strings.HasPrefix(rest, "?'"):
		if e := strings.IndexByte(rest[2:], '\''); e > 0 {
			return 2 + e + 2, true, rest[2 : 2+e]
		}
	case strings.HasPrefix(rest, "?"), strings.HasPrefix(rest, "*"):
		return 1, false, ""
	}
	return 1, true, ""
}

// scan finds every capturing group and the index of its closing parenthesis.
func (s *subExpander) scan() {
	var stack []int // indexes into groups, -1 for a group that does not capture
	count := 0
	for i := 0; i < len(s.body); i++ {
		if j := s.skipOpaque(i); j != i {
			i = j - 1
			continue
		}
		switch s.body[i] {
		case '(':
			n, capture, name := s.opener(i)
			if capture {
				count++
				s.groups = append(s.groups, pcreGroup{open: i, inner: i + n, num: count, name: name})
				stack = append(stack, len(s.groups)-1)
			} else {
				stack = append(stack, -1)
			}
			i += n - 1
		case ')':
			if len(stack) > 0 {
				if g := stack[len(stack)-1]; g >= 0 {
					s.groups[g].end = i
				}
				stack = stack[:len(stack)-1]
			}
		}
	}
}

func (s *subExpander) byName(name string) *pcreGroup {
	for i := range s.groups {
		if s.groups[i].name == name {
			return &s.groups[i]
		}
	}
	return nil
}

func (s *subExpander) byNumber(n int) *pcreGroup {
	for i := range s.groups {
		if s.groups[i].num == n {
			return &s.groups[i]
		}
	}
	return nil
}

// render writes body[lo:hi] with its calls expanded. uncapture turns the groups
// it passes into non-capturing ones, for a copy of a group's text.
func (s *subExpander) render(lo, hi, depth int, uncapture bool) string {
	var out strings.Builder
	seen := 0 // capturing groups opened before the position, for relative calls
	for _, g := range s.groups {
		if g.open < lo {
			seen++
		}
	}
	for i := lo; i < hi; i++ {
		if j := s.skipOpaque(i); j != i {
			if j > hi {
				j = hi
			}
			out.WriteString(s.body[i:j])
			i = j - 1
			continue
		}
		if s.body[i] != '(' {
			out.WriteByte(s.body[i])
			continue
		}
		if call, n, ok := s.call(i, seen); ok {
			out.WriteString(s.expandCall(call, depth))
			i += n - 1
			continue
		}
		if strings.HasPrefix(s.body[i:], "(?(DEFINE)") {
			end := s.matching(i)
			out.WriteString("(?:")
			out.WriteString(s.render(i+len("(?(DEFINE)"), end, depth, uncapture))
			out.WriteString("){0}")
			i = end
			continue
		}
		n, capture, _ := s.opener(i)
		if capture {
			seen++
			if uncapture {
				out.WriteString("(?:")
				i += n - 1
				continue
			}
		}
		out.WriteByte('(')
		out.WriteString(s.body[i+1 : i+n])
		i += n - 1
	}
	return out.String()
}

// matching answers the index of the parenthesis that closes the one at open.
func (s *subExpander) matching(open int) int {
	depth := 0
	for i := open; i < len(s.body); i++ {
		if j := s.skipOpaque(i); j != i {
			i = j - 1
			continue
		}
		switch s.body[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(s.body)
}

// call reads a call at body[i] == '(' and answers its target (nil for the whole
// pattern), the length of the call and whether there is one.
func (s *subExpander) call(i, seen int) (*pcreGroup, int, bool) {
	rest := s.body[i:]
	switch {
	case strings.HasPrefix(rest, "(?&"), strings.HasPrefix(rest, "(?P>"):
		start := 3
		if rest[2] == 'P' {
			start = 4
		}
		if e := strings.IndexByte(rest, ')'); e > start {
			if g := s.byName(rest[start:e]); g != nil {
				return g, e + 1, true
			}
		}
	case strings.HasPrefix(rest, "(?R)"):
		return nil, 4, true
	default:
		if len(rest) > 3 && rest[1] == '?' {
			j := 2
			sign := byte(0)
			if rest[j] == '-' || rest[j] == '+' {
				sign = rest[j]
				j++
			}
			k := j
			for k < len(rest) && rest[k] >= '0' && rest[k] <= '9' {
				k++
			}
			if k > j && k < len(rest) && rest[k] == ')' {
				n, _ := strconv.Atoi(rest[j:k])
				switch sign {
				case '-':
					n = seen - n + 1
				case '+':
					n = seen + n
				}
				if n == 0 {
					return nil, k + 1, true
				}
				if g := s.byNumber(n); g != nil {
					return g, k + 1, true
				}
			}
		}
	}
	return nil, 0, false
}

func (s *subExpander) expandCall(target *pcreGroup, depth int) string {
	if depth >= maxSubroutineDepth {
		return "(?!)"
	}
	if target == nil {
		return "(?:" + s.render(0, len(s.body), depth+1, true) + ")"
	}
	return "(?:" + s.render(target.inner, target.end, depth+1, true) + ")"
}
