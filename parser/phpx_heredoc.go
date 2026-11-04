package parser

import (
	"strings"
)

// phpxHeredoc reads `<<<ID`, `<<<"ID"` or `<<<'ID'` at l.pos through the line
// that closes it, and emits one string token. A nowdoc is the body verbatim. A
// heredoc is the body read as a double-quoted literal, so the scan that decodes
// escapes and finds `$name` and `{$expr}` is the one a quoted string uses. It
// reports false, leaving l untouched, when the text is not a heredoc opener.
func (l *lexer) phpxHeredoc() bool {
	src := l.src
	i := l.pos + 3
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	var closer byte
	if i < len(src) && (src[i] == '\'' || src[i] == '"') {
		closer = src[i]
		i++
	}
	nowdoc := closer == '\''
	idStart := i
	for i < len(src) && isIdentPart(rune(src[i])) {
		i++
	}
	if i == idStart || !isIdentStart(rune(src[idStart])) {
		return false
	}
	id := src[idStart:i]
	if closer != 0 {
		if i >= len(src) || src[i] != closer {
			return false
		}
		i++
	}
	if i < len(src) && src[i] == '\r' {
		i++
	}
	if i >= len(src) || src[i] != '\n' {
		return false
	}
	bodyStart := i + 1

	// The closing line: optional indentation, the id, then a non-identifier.
	var lines []string
	pos, indent, end := bodyStart, "", -1
	for pos <= len(src) {
		eol := strings.IndexByte(src[pos:], '\n')
		line := src[pos:]
		if eol >= 0 {
			line = src[pos : pos+eol]
		}
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, id) && (len(trimmed) == len(id) || !isIdentPart(rune(trimmed[len(id)]))) {
			indent = line[:len(line)-len(trimmed)]
			end = pos + len(indent) + len(id)
			break
		}
		lines = append(lines, strings.TrimSuffix(line, "\r"))
		if eol < 0 {
			return false
		}
		pos += eol + 1
	}
	if end < 0 {
		return false
	}
	for k, line := range lines {
		lines[k] = strings.TrimPrefix(line, indent)
	}
	body := strings.Join(lines, "\n")

	startLine := l.line
	l.line += strings.Count(src[l.pos:end], "\n")
	l.pos = end
	if nowdoc {
		l.tokens = append(l.tokens, token{kind: tString, val: body, raw: "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(body) + "'", pos: l.pos, line: l.line})
		return true
	}
	raw := `"` + phpxQuoteBody(body) + `"`
	parts, _, interp, err := scanInterp(raw, 1, startLine)
	if err != nil {
		return false
	}
	if interp {
		l.tokens = append(l.tokens, token{kind: tInterp, raw: raw, pos: l.pos, line: l.line})
		return true
	}
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Text)
	}
	l.tokens = append(l.tokens, token{kind: tString, val: b.String(), raw: raw, pos: l.pos, line: l.line})
	return true
}

// phpxQuoteBody makes a heredoc body a double-quoted literal's body: a bare
// quote is escaped, and an escaped quote keeps its backslash as php does.
func phpxQuoteBody(body string) string {
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		switch c := body[i]; {
		case c == '\\' && i+1 < len(body) && body[i+1] == '"':
			b.WriteString(`\\\"`)
			i++
		case c == '\\' && i+1 < len(body):
			b.WriteByte(c)
			b.WriteByte(body[i+1])
			i++
		case c == '"':
			b.WriteString(`\"`)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
