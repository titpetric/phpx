//go:build phpx

package core

import (
	"io"
	iofs "io/fs"
	"os"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

// highlight_file and highlight_string answer PHP's syntax-highlighted HTML:
// the source, escaped, in <pre><code> and coloured spans. PHP's error
// renderer calls highlight_file() to show the source of a failure, so a Laravel
// application that reaches the error page needs it.
func init() {
	runner.RegisterArea(`Runtime\PX\Core\Highlight`, registerPhpxHighlight)
}

func registerPhpxHighlight(rt *runner.Runtime) {
	highlight := func(source string, ret ...any) (any, error) {
		out := phpxHighlightSource(source)
		if len(ret) > 0 && phpval.Truthy(ret[0]) {
			return out, nil
		}
		if _, err := io.WriteString(rt.Output(), out); err != nil {
			return nil, err
		}
		return true, nil
	}
	// highlight_string highlights the source given as a string; $return
	// answers it as a string instead of writing it, as PHP's does.
	rt.RegisterFunc("highlight_string", highlight)
	// highlight_file highlights the source of $filename; $return answers it as
	// a string, and false names a file that cannot be read.
	file := func(filename string, ret ...any) (any, error) {
		source, ok := phpxHighlightRead(rt, filename)
		if !ok {
			return false, nil
		}
		return highlight(source, ret...)
	}
	rt.RegisterFunc("highlight_file", file)
	// show_source is PHP's alias of highlight_file.
	rt.RegisterFunc("show_source", file)
}

// phpxHighlightRead reads the file to highlight from the source filesystem
// first, then from the host, the order file_get_contents uses.
func phpxHighlightRead(rt *runner.Runtime, filename string) (string, bool) {
	if source := rt.FS(); source != nil {
		if body, err := iofs.ReadFile(source, rt.ResolvePath(filename)); err == nil {
			return string(body), true
		}
	}
	body, err := os.ReadFile(filename)
	if err != nil && strings.HasPrefix(filename, "/") {
		// A host run with no source filesystem reads an absolute path
		// directly. Under a virtual root the same path is relative to the
		// process's working directory, so a second read tries that spelling.
		body, err = os.ReadFile(strings.TrimPrefix(filename, "/"))
	}
	if err != nil {
		return "", false
	}
	return string(body), true
}

// PHP's highlighter colours: the default is an identifier, variable, number,
// tag or magic constant; a keyword or punctuation mark; a comment; a string; and
// inline HTML, which prints in the wrapper's own colour and outside a span.
const (
	phpxHLDefault = "#0000BB"
	phpxHLKeyword = "#007700"
	phpxHLComment = "#FF8000"
	phpxHLString  = "#DD0000"
)

// phpxHLKeywords is PHP's case-insensitive keyword list. A word absent here is
// an identifier and prints in the default colour, as php prints it.
var phpxHLKeywords = map[string]bool{
	"abstract": true, "and": true, "array": true, "as": true, "break": true,
	"callable": true, "case": true, "catch": true, "class": true, "clone": true,
	"const": true, "continue": true, "declare": true, "default": true, "die": true,
	"do": true, "echo": true, "else": true, "elseif": true, "empty": true,
	"enddeclare": true, "endfor": true, "endforeach": true, "endif": true,
	"endswitch": true, "endwhile": true, "eval": true, "exit": true, "extends": true,
	"final": true, "finally": true, "fn": true, "for": true, "foreach": true,
	"function": true, "global": true, "goto": true, "if": true, "implements": true,
	"include": true, "include_once": true, "instanceof": true, "insteadof": true,
	"interface": true, "isset": true, "list": true, "match": true, "namespace": true,
	"new": true, "or": true, "print": true, "private": true,
	"protected": true, "public": true, "readonly": true, "require": true,
	"require_once": true, "return": true, "static": true, "switch": true,
	"throw": true, "trait": true, "try": true, "unset": true, "use": true,
	"var": true, "while": true, "xor": true, "yield": true,
}

// phpxHighlighter writes one highlighted source, keeping the open span colour
// between tokens so a run of one colour prints as one span, as php groups them.
type phpxHighlighter struct {
	b     strings.Builder
	color string
	// lastWord is the last identifier or keyword written, which decides the
	// contextual words `from` and `enum`.
	lastWord string
}

// put writes text in color, opening or closing a span only when the colour
// changes. An empty colour means inline HTML, which the wrapper paints.
func (h *phpxHighlighter) put(color, text string) {
	if text == "" {
		return
	}
	if color != h.color {
		if h.color != "" {
			h.b.WriteString("</span>")
		}
		if color != "" {
			h.b.WriteString(`<span style="color: `)
			h.b.WriteString(color)
			h.b.WriteString(`">`)
		}
		h.color = color
	}
	phpxHLEscape(&h.b, text)
}

// phpxHighlightSource renders source the way highlight_string() and
// highlight_file() answer it.
func phpxHighlightSource(src string) string {
	h := &phpxHighlighter{}
	h.b.WriteString(`<pre><code style="color: #000000">`)
	i := 0
	inPHP := false
	for i < len(src) {
		if !inPHP {
			j := phpxHLOpenTag(src, i)
			if j < 0 {
				h.put("", src[i:])
				break
			}
			h.put("", src[i:j])
			i = h.openTag(src, j)
			inPHP = true
			continue
		}
		i, inPHP = h.php(src, i)
	}
	if h.color != "" {
		h.b.WriteString("</span>")
	}
	h.b.WriteString("</code></pre>")
	return h.b.String()
}

// phpxHLOpenTag returns the offset of the next open tag at or after i, or -1.
func phpxHLOpenTag(src string, i int) int {
	for j := i; j+1 < len(src); j++ {
		if src[j] == '<' && src[j+1] == '?' {
			if strings.HasPrefix(src[j:], "<?php") || strings.HasPrefix(src[j:], "<?=") {
				return j
			}
		}
	}
	return -1
}

// openTag writes the open tag, which takes one following whitespace character as
// PHP's token does, and returns the offset after it.
func (h *phpxHighlighter) openTag(src string, i int) int {
	j := i + 3
	if strings.HasPrefix(src[i:], "<?=") {
		j = i + 3
	} else if j < len(src) {
		if src[j] == '\r' && j+1 < len(src) && src[j+1] == '\n' {
			j += 2
		} else if src[j] == ' ' || src[j] == '\t' || src[j] == '\n' || src[j] == '\r' {
			j++
		}
	}
	h.put(phpxHLDefault, src[i:j])
	h.lastWord = ""
	return j
}

// php scans one stretch of PHP source, returning the offset and whether the
// scan is still inside PHP.
func (h *phpxHighlighter) php(src string, i int) (int, bool) {
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f':
			j := i
			for j < len(src) && phpxHLSpace(src[j]) {
				j++
			}
			h.put(h.color, src[i:j])
			i = j

		case strings.HasPrefix(src[i:], "?>"):
			// PHP's close tag takes one following newline, and the highlighter
			// prints it inside the tag's span.
			j := i + 2
			if j < len(src) && src[j] == '\r' && j+1 < len(src) && src[j+1] == '\n' {
				j += 2
			} else if j < len(src) && src[j] == '\n' {
				j++
			}
			h.put(phpxHLDefault, src[i:j])
			return j, false

		case strings.HasPrefix(src[i:], "//"), c == '#' && !strings.HasPrefix(src[i:], "#["):
			j := phpxHLLineEnd(src, i)
			h.put(phpxHLComment, src[i:j])
			i = j

		case strings.HasPrefix(src[i:], "/*"):
			j := phpxHLBlockEnd(src, i)
			h.put(phpxHLComment, src[i:j])
			i = j

		case c == '#' && strings.HasPrefix(src[i:], "#["):
			h.put(phpxHLKeyword, "#[")
			i += 2

		case c == '\'':
			j := phpxHLSingleEnd(src, i)
			h.put(phpxHLString, src[i:j])
			i = j

		case c == '"', c == '`':
			i = h.quoted(src, i, c)

		case strings.HasPrefix(src[i:], "<<<") && phpxHLHeredocAt(src, i):
			i = h.heredoc(src, i)

		case c == '$' && i+1 < len(src) && phpxHLIdentStart(src[i+1]):
			j := phpxHLIdentEnd(src, i+1)
			h.put(phpxHLDefault, src[i:j])
			h.lastWord = ""
			i = j

		case (c == '\\' || phpxHLIdentStart(c)) && phpxHLNameAt(src, i):
			j := phpxHLNameEnd(src, i)
			word := src[i:j]
			switch {
			case phpxHLAfterMember(src, i):
				// A name after -> or ?-> is a member, which php reads as a
				// T_STRING even when it spells a keyword.
				h.put(phpxHLDefault, word)
				h.lastWord = ""
			case h.keyword(src, i, j, word):
				h.put(phpxHLKeyword, word)
				h.lastWord = strings.ToLower(word)
			default:
				h.put(phpxHLDefault, word)
				h.lastWord = ""
			}
			i = j

		case strings.HasPrefix(src[i:], "?->"):
			h.put(phpxHLKeyword, "?->")
			i += 3

		case strings.HasPrefix(src[i:], "->"):
			h.put(phpxHLKeyword, "->")
			i += 2

		case strings.HasPrefix(src[i:], "::"):
			h.put(phpxHLKeyword, "::")
			i += 2

		case c >= '0' && c <= '9':
			j := phpxHLNumberEnd(src, i)
			h.put(phpxHLDefault, src[i:j])
			i = j

		case c == '.' && i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9':
			j := phpxHLNumberEnd(src, i)
			h.put(phpxHLDefault, src[i:j])
			i = j

		case c == '(':
			// A cast is one keyword token, parentheses included.
			if j, ok := phpxHLCastEnd(src, i); ok {
				h.put(phpxHLKeyword, src[i:j])
				i = j
			} else {
				h.put(phpxHLKeyword, "(")
				i++
			}

		default:
			h.put(phpxHLKeyword, src[i:i+1])
			i++
		}
	}
	return i, true
}

// keyword reports whether word prints as a keyword. `enum` is contextual: php
// reads it as one only before a name, and `from` only after `yield`.
func (h *phpxHighlighter) keyword(src string, start, end int, word string) bool {
	low := strings.ToLower(word)
	if low == "enum" {
		j := end
		for j < len(src) && phpxHLSpace(src[j]) {
			j++
		}
		return j < len(src) && (phpxHLIdentStart(src[j]) || src[j] == '\\')
	}
	if low == "from" {
		return h.lastWord == "yield"
	}
	return phpxHLKeywords[low]
}

// quoted writes a double-quoted or backtick literal, whose text is red and
// whose embedded expressions print as code: a variable and its property or
// subscript in the default colour, the brackets and braces green, and a string
// in the braces red.
func (h *phpxHighlighter) quoted(src string, i int, quote byte) int {
	h.put(phpxHLString, src[i:i+1])
	i++
	text := i
	for i < len(src) {
		switch {
		case src[i] == '\\' && i+1 < len(src):
			i += 2
		case src[i] == quote:
			h.put(phpxHLString, src[text:i+1])
			return i + 1
		case src[i] == '$' && i+1 < len(src) && phpxHLIdentStart(src[i+1]):
			h.put(phpxHLString, src[text:i])
			i = h.simpleInterp(src, i)
			text = i
		case src[i] == '{' && i+1 < len(src) && src[i+1] == '$':
			h.put(phpxHLString, src[text:i])
			i = h.curlyInterp(src, i)
			text = i
		default:
			i++
		}
	}
	h.put(phpxHLString, src[text:])
	return i
}

// simpleInterp writes `$name`, `$name[key]` and `$name->prop`, the shapes a
// simple interpolation reaches, and returns the offset after it.
func (h *phpxHighlighter) simpleInterp(src string, i int) int {
	j := phpxHLIdentEnd(src, i+1)
	h.put(phpxHLDefault, src[i:j])
	i = j
	if i < len(src) && src[i] == '[' {
		end := strings.IndexByte(src[i:], ']')
		if end < 0 {
			return i
		}
		key := src[i+1 : i+end]
		h.put(phpxHLKeyword, "[")
		h.interpKey(key)
		h.put(phpxHLKeyword, "]")
		return i + end + 1
	}
	if strings.HasPrefix(src[i:], "->") && i+2 < len(src) && phpxHLIdentStart(src[i+2]) {
		j = phpxHLIdentEnd(src, i+2)
		h.put(phpxHLKeyword, "->")
		h.put(phpxHLDefault, src[i+2:j])
		return j
	}
	return i
}

// interpKey writes a simple subscript: a number or name in the default colour
// and a quoted key as a string.
func (h *phpxHighlighter) interpKey(key string) {
	if key == "" {
		return
	}
	switch {
	case key[0] == '\'' || key[0] == '"':
		h.put(phpxHLString, key)
	case key[0] == '$':
		h.put(phpxHLDefault, key)
	case key[0] >= '0' && key[0] <= '9':
		h.put(phpxHLDefault, key)
	default:
		h.put(phpxHLDefault, key)
	}
}

// curlyInterp writes `{$expr}`: the braces green and the expression as PHP
// code, to the brace that closes the interpolation.
func (h *phpxHighlighter) curlyInterp(src string, i int) int {
	h.put(phpxHLKeyword, "{")
	i++
	depth := 0
	text := i
	for i < len(src) {
		c := src[i]
		switch {
		case c == '}' && depth == 0:
			h.put(phpxHLKeyword, "}")
			return i + 1
		case c == '{':
			depth++
			i++
		case c == '}':
			depth--
			i++
		case c == '\'':
			h.put(phpxHLString, src[text:i])
			j := phpxHLSingleEnd(src, i)
			h.put(phpxHLString, src[i:j])
			i, text = j, j
		case c == '"':
			h.put(phpxHLString, src[text:i])
			i = h.quoted(src, i, '"')
			text = i
		case c == '$' && i+1 < len(src) && phpxHLIdentStart(src[i+1]):
			h.put(phpxHLString, src[text:i])
			i = h.simpleInterp(src, i)
			text = i
		case (c == '\\' || phpxHLIdentStart(c)) && phpxHLNameAt(src, i):
			h.put(phpxHLString, src[text:i])
			j := phpxHLNameEnd(src, i)
			word := src[i:j]
			if phpxHLKeywords[strings.ToLower(word)] {
				h.put(phpxHLKeyword, word)
			} else {
				h.put(phpxHLDefault, word)
			}
			i, text = j, j
		case c >= '0' && c <= '9':
			h.put(phpxHLString, src[text:i])
			j := phpxHLNumberEnd(src, i)
			h.put(phpxHLDefault, src[i:j])
			i, text = j, j
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			// Whitespace inherits the colour of the token before it.
			j := i
			for j < len(src) && phpxHLSpace(src[j]) {
				j++
			}
			h.put(h.color, src[i:j])
			i, text = j, j
		default:
			h.put(phpxHLString, src[text:i])
			h.put(phpxHLKeyword, src[i:i+1])
			i, text = i+1, i+1
		}
	}
	h.put(phpxHLString, src[text:])
	return i
}

// heredoc writes `<<<ID`, a body read as a quoted literal, and the closing ID.
func (h *phpxHighlighter) heredoc(src string, i int) int {
	open, _, body, closeAt, end := phpxHLHeredocRange(src, i)
	if end <= i {
		h.put(phpxHLKeyword, src[i:i+3])
		return i + 3
	}
	h.put(phpxHLKeyword, src[i:open])
	if strings.Contains(src[i:open], "<<<'") {
		// A nowdoc body is text: it interpolates nothing, so it is all red.
		h.put(phpxHLString, src[body[0]:body[1]])
	} else {
		h.interpBody(src, body[0], body[1])
	}
	h.put(phpxHLKeyword, src[closeAt:end])
	return end
}

// interpBody writes a heredoc body: text red, with the expressions a quoted
// literal embeds printed as code.
func (h *phpxHighlighter) interpBody(src string, start, end int) {
	i := start
	text := i
	for i < end {
		switch {
		case src[i] == '\\' && i+1 < end:
			i += 2
		case src[i] == '$' && i+1 < end && phpxHLIdentStart(src[i+1]):
			h.put(phpxHLString, src[text:i])
			i = h.simpleInterp(src, i)
			text = i
		case src[i] == '{' && i+1 < end && src[i+1] == '$':
			h.put(phpxHLString, src[text:i])
			i = h.curlyInterp(src, i)
			text = i
		default:
			i++
		}
	}
	h.put(phpxHLString, src[text:end])
}

// phpxHLHeredocAt reports whether a heredoc opener starts at i.
func phpxHLHeredocAt(src string, i int) bool {
	_, _, _, _, end := phpxHLHeredocRange(src, i)
	return end > i
}

// phpxHLHeredocRange reads a heredoc at i and returns the offset of the opener's
// newline, the identifier, the body's start and end, the closing ID's offset and
// the offset after it. end == i when the text is not a heredoc.
func phpxHLHeredocRange(src string, i int) (open int, id string, body [2]int, closeAt, end int) {
	j := i + 3
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	var quote byte
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		quote = src[j]
		j++
	}
	start := j
	if j >= len(src) || !phpxHLIdentStart(src[j]) {
		return 0, "", body, 0, i
	}
	for j < len(src) && phpxHLIdentPart(src[j]) {
		j++
	}
	id = src[start:j]
	if quote != 0 {
		if j >= len(src) || src[j] != quote {
			return 0, "", body, 0, i
		}
		j++
	}
	if j < len(src) && src[j] == '\r' {
		j++
	}
	if j >= len(src) || src[j] != '\n' {
		return 0, "", body, 0, i
	}
	open = j + 1
	body[0] = open
	// The closing line: optional indentation, the identifier, and a byte that
	// does not continue it.
	for pos := open; pos <= len(src); {
		eol := strings.IndexByte(src[pos:], '\n')
		line := src[pos:]
		if eol >= 0 {
			line = src[pos : pos+eol]
		}
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, id) && (len(trimmed) == len(id) || !phpxHLIdentPart(trimmed[len(id)])) {
			// The closing line's indentation is part of the end tag token, so
			// it prints green with the identifier, not red with the body.
			closeAt = pos
			end = pos + len(line) - len(trimmed) + len(id)
			body[1] = pos
			return open, id, body, closeAt, end
		}
		if eol < 0 {
			break
		}
		pos += eol + 1
	}
	return 0, "", body, 0, i
}

// phpxHLEscape writes text with the three entities PHP's highlighter escapes.
func phpxHLEscape(b *strings.Builder, text string) {
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '\t':
			// PHP's highlighter writes a tab as four spaces wherever it
			// appears, in code and in inline HTML alike.
			b.WriteString("    ")
		default:
			b.WriteByte(text[i])
		}
	}
}

// phpxHLLineEnd returns the offset of the newline ending the comment at i, or
// the end of the source.
func phpxHLLineEnd(src string, i int) int {
	if j := strings.IndexByte(src[i:], '\n'); j >= 0 {
		return i + j
	}
	return len(src)
}

// phpxHLBlockEnd returns the offset after the comment closing at i.
func phpxHLBlockEnd(src string, i int) int {
	if j := strings.Index(src[i+2:], "*/"); j >= 0 {
		return i + 2 + j + 2
	}
	return len(src)
}

// phpxHLSingleEnd returns the offset after the single-quoted literal at i.
func phpxHLSingleEnd(src string, i int) int {
	for j := i + 1; j < len(src); j++ {
		if src[j] == '\\' && j+1 < len(src) {
			j++
			continue
		}
		if src[j] == '\'' {
			return j + 1
		}
	}
	return len(src)
}

// phpxHLNameAt reports whether a name (optionally namespace-qualified) starts at i.
func phpxHLNameAt(src string, i int) bool {
	if i >= len(src) {
		return false
	}
	if phpxHLIdentStart(src[i]) {
		return true
	}
	return src[i] == '\\' && i+1 < len(src) && phpxHLIdentStart(src[i+1])
}

// phpxHLNameEnd returns the offset after the name starting at i, backslashes
// included, since PHP's tokenizer keeps a qualified name as one token.
func phpxHLNameEnd(src string, i int) int {
	j := i
	if j < len(src) && src[j] == '\\' {
		j++
	}
	j = phpxHLIdentEnd(src, j)
	for j < len(src) && src[j] == '\\' && j+1 < len(src) && phpxHLIdentStart(src[j+1]) {
		j = phpxHLIdentEnd(src, j+1)
	}
	return j
}

// phpxHLNumberEnd returns the offset after the number starting at i.
func phpxHLNumberEnd(src string, i int) int {
	j := i
	for j < len(src) {
		c := src[j]
		if c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '.' {
			j++
			continue
		}
		if (c == '+' || c == '-') && j > i && (src[j-1] == 'e' || src[j-1] == 'E') {
			j++
			continue
		}
		break
	}
	return j
}

// phpxHLCasts names the words a cast spells; php prints the whole `(type)` as
// one keyword token.
var phpxHLCasts = map[string]bool{
	"int": true, "integer": true, "bool": true, "boolean": true,
	"float": true, "double": true, "real": true, "string": true,
	"binary": true, "array": true, "object": true, "unset": true,
}

// phpxHLCastEnd returns the offset after a `(type)` cast at i.
func phpxHLCastEnd(src string, i int) (int, bool) {
	j := i + 1
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	k := j
	for k < len(src) && phpxHLIdentPart(src[k]) {
		k++
	}
	if k == j || !phpxHLCasts[strings.ToLower(src[j:k])] {
		return 0, false
	}
	for k < len(src) && (src[k] == ' ' || src[k] == '\t') {
		k++
	}
	if k < len(src) && src[k] == ')' {
		return k + 1, true
	}
	return 0, false
}

// phpxHLAfterMember reports whether the name at i follows -> or ?->, where php
// reads it as a member and not as a keyword.
func phpxHLAfterMember(src string, i int) bool {
	j := i - 1
	for j >= 0 && phpxHLSpace(src[j]) {
		j--
	}
	return j >= 1 && src[j] == '>' && src[j-1] == '-'
}

// phpxHLIdentStart reports whether c starts a PHP identifier.
func phpxHLIdentStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

// phpxHLIdentPart reports whether c continues a PHP identifier.
func phpxHLIdentPart(c byte) bool {
	return phpxHLIdentStart(c) || c >= '0' && c <= '9'
}

// phpxHLIdentEnd returns the offset after the identifier starting at i.
func phpxHLIdentEnd(src string, i int) int {
	for i < len(src) && phpxHLIdentPart(src[i]) {
		i++
	}
	return i
}

// phpxHLSpace reports whether c is whitespace.
func phpxHLSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}
