package parser

import (
	"strconv"
	"strings"
)

// phpxHalt reads the `();` or `?>` after __halt_compiler and stops the lexer
// there: what follows is data the script reads with __COMPILER_HALT_OFFSET__.
// It rewrites __COMPILER_HALT_OFFSET__ to the offset of the data and emits
// nothing for the statement, and reports false when the identifier is no call.
func (l *lexer) phpxHalt() bool {
	i := l.pos
	skip := func() {
		for i < len(l.src) && strings.IndexByte(" \t\r\n", l.src[i]) >= 0 {
			i++
		}
	}
	skip()
	if i >= len(l.src) || l.src[i] != '(' {
		return false
	}
	i++
	skip()
	if i >= len(l.src) || l.src[i] != ')' {
		return false
	}
	i++
	skip()
	switch {
	case i < len(l.src) && l.src[i] == ';':
		i++
	case strings.HasPrefix(l.src[i:], "?>"):
		i += 2
		if i < len(l.src) && l.src[i] == '\n' {
			i++
		}
	default:
		return false
	}
	// Every use of the constant, which the file reads before the data, becomes
	// the offset itself.
	offset := strconv.Itoa(i)
	for n := range l.tokens {
		if l.tokens[n].kind == tIdent && l.tokens[n].val == "__COMPILER_HALT_OFFSET__" {
			l.tokens[n].kind, l.tokens[n].val = tInt, offset
		}
	}
	l.pos = len(l.src)
	return true
}
