//go:build phpx

package phpx

import "strings"

// scanDouble finds the closing quote of a double-quoted string that starts at
// start, skipping escapes and {$ ... } regions, and reports whether the string
// interpolates anything.
func (l *phpTokens) scanDouble(start int) (end int, interp bool) {
	j := start + 1
	for j < len(l.src) {
		c := l.src[j]
		switch {
		case c == '\\':
			j += 2
			continue
		case c == '"':
			return j, interp
		case c == '$' && j+1 < len(l.src) && (isIdentStart(l.src[j+1]) || l.src[j+1] == '{'):
			interp = true
		case c == '{' && j+1 < len(l.src) && l.src[j+1] == '$':
			interp = true
			j = l.skipBraces(j)
			continue
		}
		j++
	}
	return len(l.src), interp
}

// skipBraces returns the index just past the } that closes the { at j, reading
// quoted strings inside as units.
func (l *phpTokens) skipBraces(j int) int {
	depth := 0
	for j < len(l.src) {
		switch l.src[j] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return j + 1
			}
		case '\'', '"':
			q := l.src[j]
			j++
			for j < len(l.src) && l.src[j] != q {
				if l.src[j] == '\\' {
					j++
				}
				j++
			}
		}
		j++
	}
	return j
}

func (l *phpTokens) doubleQuoted() {
	end, interp := l.scanDouble(l.pos)
	if !interp {
		stop := min(end+1, len(l.src))
		l.emit("T_CONSTANT_ENCAPSED_STRING", l.src[l.pos:stop])
		l.pos = stop
		return
	}
	l.emitChar('"')
	l.pos++
	l.parts(end)
	if l.pos < len(l.src) {
		l.emitChar('"')
		l.pos++
	}
}

// parts tokenizes a string body up to end: literal runs and the embedded
// variables and expressions.
func (l *phpTokens) parts(end int) {
	litStart := l.pos
	flush := func(upTo int) {
		if upTo > litStart {
			l.emit("T_ENCAPSED_AND_WHITESPACE", l.src[litStart:upTo])
		}
	}
	for l.pos < end {
		c := l.src[l.pos]
		switch {
		case c == '\\':
			l.pos += 2
		case c == '{' && l.pos+1 < end && l.src[l.pos+1] == '$':
			flush(l.pos)
			l.emit("T_CURLY_OPEN", "{")
			l.pos++
			l.braced()
			litStart = l.pos
		case c == '$' && l.pos+1 < end && l.src[l.pos+1] == '{':
			flush(l.pos)
			l.emit("T_DOLLAR_OPEN_CURLY_BRACES", "${")
			l.pos += 2
			j := l.pos
			for j < len(l.src) && isIdentByte(l.src[j]) {
				j++
			}
			if j > l.pos && j < len(l.src) && (l.src[j] == '}' || l.src[j] == '[') {
				l.emit("T_STRING_VARNAME", l.src[l.pos:j])
				l.pos = j
			}
			l.braced()
			litStart = l.pos
		case c == '$' && l.pos+1 < end && isIdentStart(l.src[l.pos+1]):
			flush(l.pos)
			l.simpleVariable(end)
			litStart = l.pos
		default:
			l.pos++
		}
	}
	l.pos = min(l.pos, end)
	flush(l.pos)
}

// braced tokenizes php code until the } that closes the group opened before it,
// and emits that }. The depth starts at one.
func (l *phpTokens) braced() {
	depth := 1
	for l.pos < len(l.src) && depth > 0 {
		before := len(l.out)
		c := l.src[l.pos]
		switch c {
		case '{':
			depth++
			l.emitChar(c)
			l.pos++
		case '}':
			depth--
			l.emitChar(c)
			l.pos++
		default:
			l.step()
		}
		_ = before
	}
}

// simpleVariable reads "$name", "$name[key]" and "$name->prop" inside a string.
func (l *phpTokens) simpleVariable(end int) {
	j := l.pos + 1
	for j < end && isIdentByte(l.src[j]) {
		j++
	}
	l.emit("T_VARIABLE", l.src[l.pos:j])
	l.pos = j
	switch {
	case l.pos < end && l.src[l.pos] == '[':
		k := l.pos + 1
		var name string
		switch {
		case k < end && l.src[k] == '$':
			m := k + 1
			for m < end && isIdentByte(l.src[m]) {
				m++
			}
			name = "T_VARIABLE"
			l.emitChar('[')
			l.emit(name, l.src[k:m])
			k = m
		case k < end && (l.src[k] >= '0' && l.src[k] <= '9' || l.src[k] == '-'):
			m := k + 1
			for m < end && l.src[m] >= '0' && l.src[m] <= '9' {
				m++
			}
			l.emitChar('[')
			l.emit("T_NUM_STRING", l.src[k:m])
			k = m
		case k < end && isIdentStart(l.src[k]):
			m := k
			for m < end && isIdentByte(l.src[m]) {
				m++
			}
			l.emitChar('[')
			l.emit("T_STRING", l.src[k:m])
			k = m
		default:
			l.pos = k - 1
			return
		}
		if k < end && l.src[k] == ']' {
			l.emitChar(']')
			k++
		}
		l.pos = k
	case strings.HasPrefix(l.src[l.pos:end], "->") && l.pos+2 < end && isIdentStart(l.src[l.pos+2]):
		l.emit("T_OBJECT_OPERATOR", "->")
		m := l.pos + 2
		for m < end && isIdentByte(l.src[m]) {
			m++
		}
		l.emit("T_STRING", l.src[l.pos+2:m])
		l.pos = m
	case strings.HasPrefix(l.src[l.pos:end], "?->") && l.pos+3 < end && isIdentStart(l.src[l.pos+3]):
		l.emit("T_NULLSAFE_OBJECT_OPERATOR", "?->")
		m := l.pos + 3
		for m < end && isIdentByte(l.src[m]) {
			m++
		}
		l.emit("T_STRING", l.src[l.pos+3:m])
		l.pos = m
	}
}

// heredoc reads <<<ID, <<<"ID" and <<<'ID' through the closing line. It reports
// false when the text is not a heredoc opener.
func (l *phpTokens) heredoc() bool {
	src := l.src
	i := l.pos + 3
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	var quote byte
	if i < len(src) && (src[i] == '\'' || src[i] == '"') {
		quote = src[i]
		i++
	}
	idStart := i
	for i < len(src) && isIdentByte(src[i]) {
		i++
	}
	if i == idStart || !isIdentStart(src[idStart]) {
		return false
	}
	id := src[idStart:i]
	if quote != 0 {
		if i >= len(src) || src[i] != quote {
			return false
		}
		i++
	}
	if strings.HasPrefix(src[i:], "\r\n") {
		i += 2
	} else if i < len(src) && (src[i] == '\n' || src[i] == '\r') {
		i++
	} else {
		return false
	}
	// find the closing line
	bodyStart, pos := i, i
	closeAt, indentEnd := -1, -1
	for pos <= len(src) {
		j := pos
		for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
			j++
		}
		if strings.HasPrefix(src[j:], id) && (j+len(id) >= len(src) || !isIdentByte(src[j+len(id)])) {
			closeAt, indentEnd = pos, j+len(id)
			break
		}
		nl := strings.IndexByte(src[pos:], '\n')
		if nl < 0 {
			break
		}
		pos += nl + 1
	}
	if closeAt < 0 {
		return false
	}
	l.emit("T_START_HEREDOC", src[l.pos:bodyStart])
	l.pos = bodyStart
	if closeAt > bodyStart {
		if quote == '\'' {
			l.emit("T_ENCAPSED_AND_WHITESPACE", src[bodyStart:closeAt])
			l.pos = closeAt
		} else {
			l.parts(closeAt)
		}
	}
	l.emit("T_END_HEREDOC", src[closeAt:indentEnd])
	l.pos = indentEnd
	return true
}
