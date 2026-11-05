//go:build phpx

package parser

// phpxSkipAttribute drops a `#[ ... ]` attribute group, up to its closing
// bracket, so one written inside a parameter list ends where the group ends and
// not at the end of the line. Reflection reads the groups from the source text.
func (l *lexer) phpxSkipAttribute() bool {
	depth := 0
	i := l.pos + 1
	for i < len(l.src) {
		switch c := l.src[i]; c {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				for l.pos <= i {
					l.advanceRune()
				}
				return true
			}
		case '\'', '"':
			i++
			for i < len(l.src) && l.src[i] != c {
				if l.src[i] == '\\' {
					i++
				}
				i++
			}
		}
		i++
	}
	return false
}
