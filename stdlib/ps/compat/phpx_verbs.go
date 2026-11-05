package compat

import (
	"strconv"
	"strings"
)

// phpxSkipPrefix names the group phpxSkipVerbs wraps around the text before a
// (*SKIP)(*F) pair.
const phpxSkipPrefix = "__phpx_skip"

// phpxSkipVerbs rewrites `A(*SKIP)(*F)|B`, php's way to say "match B anywhere
// but inside an A", to `(?<__phpx_skip0>A)|B`. The pattern matcher then drops a
// match in which that group took part and keeps the others: the alternative
// that matched A consumed its text, and the search went on after it, as the
// verb sends a PCRE match to do.
func phpxSkipVerbs(body string, extended bool) string {
	if !strings.Contains(body, "(*SKIP)") {
		return body
	}
	var out strings.Builder
	var altStart []int // where the alternative being read begins in out, per depth
	altStart = append(altStart, 0)
	skips := 0
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '\\' && i+1 < len(body):
			out.WriteByte(c)
			i++
			out.WriteByte(body[i])
		case c == '[':
			end := i + 1
			if end < len(body) && body[end] == '^' {
				end++
			}
			if end < len(body) && body[end] == ']' {
				end++
			}
			for end < len(body) && body[end] != ']' {
				if body[end] == '\\' {
					end++
				}
				end++
			}
			if end >= len(body) {
				end = len(body) - 1
			}
			out.WriteString(body[i : end+1])
			i = end
		case c == '#' && extended:
			for i < len(body) && body[i] != '\n' {
				out.WriteByte(body[i])
				i++
			}
			if i < len(body) {
				out.WriteByte('\n')
			}
		case c == '(' && strings.HasPrefix(body[i:], "(*SKIP)"):
			rest := body[i+len("(*SKIP)"):]
			verb := 0
			switch {
			case strings.HasPrefix(rest, "(*FAIL)"):
				verb = len("(*FAIL)")
			case strings.HasPrefix(rest, "(*F)"):
				verb = len("(*F)")
			}
			if verb == 0 {
				out.WriteByte(c)
				continue
			}
			text := out.String()
			start := altStart[len(altStart)-1]
			out.Reset()
			out.WriteString(text[:start])
			out.WriteString("(?<" + phpxSkipPrefix + string(rune('0'+skips%10)) + ">")
			out.WriteString(text[start:])
			out.WriteString(")")
			skips++
			i += len("(*SKIP)") + verb - 1
		case c == '(':
			altStart = append(altStart, out.Len()+1)
			out.WriteByte(c)
		case c == ')':
			if len(altStart) > 1 {
				altStart = altStart[:len(altStart)-1]
			}
			out.WriteByte(c)
		case c == '|':
			out.WriteByte(c)
			altStart[len(altStart)-1] = out.Len()
		default:
			out.WriteByte(c)
		}
	}
	return out.String()
}

// replaceMatches is replaceAll over the matches findAllIndex keeps, for a
// pattern whose group numbers the engines do not share. repl holds ${n}
// references.
func (p *pattern) replaceMatches(subject, repl string) string {
	var b strings.Builder
	last := 0
	for _, idx := range p.findAllIndex(subject, 0) {
		b.WriteString(subject[last:idx[0]])
		for i := 0; i < len(repl); i++ {
			if repl[i] == '$' && i+1 < len(repl) && repl[i+1] == '{' {
				if end := strings.IndexByte(repl[i:], '}'); end > 0 {
					if n, err := strconv.Atoi(repl[i+2 : i+end]); err == nil {
						if 2*n+1 < len(idx) && idx[2*n] >= 0 {
							b.WriteString(subject[idx[2*n]:idx[2*n+1]])
						}
						i += end
						continue
					}
				}
			}
			b.WriteByte(repl[i])
		}
		last = idx[1]
	}
	b.WriteString(subject[last:])
	return b.String()
}
