//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxTokenizer)
}

// registerPhpxTokenizer installs token_get_all and token_name with php 8.4's
// token ids, and the T_* constants those ids belong to.
func registerPhpxTokenizer(rt *runner.Runtime) {
	for name, id := range phpxTokenConstants {
		rt.SetConst(name, id)
	}
	rt.RegisterFunc("token_get_all", func(code string, _ ...int64) *model.Array {
		lexer := &phpTokens{src: code, line: 1}
		lexer.run()
		out := model.NewArraySize(len(lexer.out))
		for _, t := range lexer.out {
			if t.id < 256 {
				out.Append(t.text)
				continue
			}
			item := model.NewArraySize(3)
			item.Append(t.id)
			item.Append(t.text)
			item.Append(int64(t.line))
			out.Append(item)
		}
		return out
	})
	rt.RegisterFunc("token_name", func(id int64) string {
		if name, ok := phpxTokenNames[id]; ok {
			return name
		}
		return "UNKNOWN"
	})
}

type phpToken struct {
	id   int64
	text string
	line int
}

// phpTokens tokenizes php source the way token_get_all does.
type phpTokens struct {
	src  string
	pos  int
	line int
	out  []phpToken
	// afterArrow is set after -> or ?->, where a keyword is a property name.
	afterArrow bool
}

var phpKeywords = map[string]string{
	"abstract": "T_ABSTRACT", "and": "T_LOGICAL_AND", "array": "T_ARRAY", "as": "T_AS", "break": "T_BREAK",
	"callable": "T_CALLABLE", "case": "T_CASE", "catch": "T_CATCH", "class": "T_CLASS", "clone": "T_CLONE",
	"const": "T_CONST", "continue": "T_CONTINUE", "declare": "T_DECLARE", "default": "T_DEFAULT", "do": "T_DO",
	"echo": "T_ECHO", "else": "T_ELSE", "elseif": "T_ELSEIF", "empty": "T_EMPTY", "enddeclare": "T_ENDDECLARE",
	"endfor": "T_ENDFOR", "endforeach": "T_ENDFOREACH", "endif": "T_ENDIF", "endswitch": "T_ENDSWITCH",
	"endwhile": "T_ENDWHILE", "eval": "T_EVAL", "exit": "T_EXIT", "die": "T_EXIT", "extends": "T_EXTENDS",
	"final": "T_FINAL", "finally": "T_FINALLY", "fn": "T_FN", "for": "T_FOR", "foreach": "T_FOREACH",
	"function": "T_FUNCTION", "global": "T_GLOBAL", "goto": "T_GOTO", "if": "T_IF", "implements": "T_IMPLEMENTS",
	"include": "T_INCLUDE", "include_once": "T_INCLUDE_ONCE", "instanceof": "T_INSTANCEOF", "insteadof": "T_INSTEADOF",
	"interface": "T_INTERFACE", "isset": "T_ISSET", "list": "T_LIST", "match": "T_MATCH", "namespace": "T_NAMESPACE",
	"new": "T_NEW", "or": "T_LOGICAL_OR", "print": "T_PRINT", "private": "T_PRIVATE", "protected": "T_PROTECTED",
	"public": "T_PUBLIC", "readonly": "T_READONLY", "require": "T_REQUIRE", "require_once": "T_REQUIRE_ONCE",
	"return": "T_RETURN", "static": "T_STATIC", "switch": "T_SWITCH", "throw": "T_THROW", "trait": "T_TRAIT",
	"try": "T_TRY", "unset": "T_UNSET", "use": "T_USE", "var": "T_VAR", "while": "T_WHILE", "xor": "T_LOGICAL_XOR",
	"yield": "T_YIELD", "__class__": "T_CLASS_C", "__dir__": "T_DIR", "__file__": "T_FILE",
	"__function__": "T_FUNC_C", "__line__": "T_LINE", "__method__": "T_METHOD_C", "__namespace__": "T_NS_C",
	"__trait__": "T_TRAIT_C", "__halt_compiler": "T_HALT_COMPILER", "enum": "T_ENUM",
}

// operators, longest first.
var phpOperators = []struct{ text, name string }{
	{"<<<", ""}, {"**=", "T_POW_EQUAL"}, {"...", "T_ELLIPSIS"}, {"<=>", "T_SPACESHIP"}, {"===", "T_IS_IDENTICAL"},
	{"!==", "T_IS_NOT_IDENTICAL"}, {"<<=", "T_SL_EQUAL"}, {">>=", "T_SR_EQUAL"}, {"??=", "T_COALESCE_EQUAL"},
	{"?->", "T_NULLSAFE_OBJECT_OPERATOR"}, {"==", "T_IS_EQUAL"}, {"!=", "T_IS_NOT_EQUAL"}, {"<>", "T_IS_NOT_EQUAL"},
	{"<=", "T_IS_SMALLER_OR_EQUAL"}, {">=", "T_IS_GREATER_OR_EQUAL"}, {"&&", "T_BOOLEAN_AND"}, {"||", "T_BOOLEAN_OR"},
	{"??", "T_COALESCE"}, {"++", "T_INC"}, {"--", "T_DEC"}, {"+=", "T_PLUS_EQUAL"}, {"-=", "T_MINUS_EQUAL"},
	{"*=", "T_MUL_EQUAL"}, {"/=", "T_DIV_EQUAL"}, {".=", "T_CONCAT_EQUAL"}, {"%=", "T_MOD_EQUAL"}, {"**", "T_POW"},
	{"<<", "T_SL"}, {">>", "T_SR"}, {"&=", "T_AND_EQUAL"}, {"|=", "T_OR_EQUAL"}, {"^=", "T_XOR_EQUAL"},
	{"->", "T_OBJECT_OPERATOR"}, {"=>", "T_DOUBLE_ARROW"}, {"::", "T_DOUBLE_COLON"},
}

var phpCasts = map[string]string{
	"int": "T_INT_CAST", "integer": "T_INT_CAST", "bool": "T_BOOL_CAST", "boolean": "T_BOOL_CAST",
	"float": "T_DOUBLE_CAST", "double": "T_DOUBLE_CAST", "real": "T_DOUBLE_CAST", "string": "T_STRING_CAST",
	"binary": "T_STRING_CAST", "array": "T_ARRAY_CAST", "object": "T_OBJECT_CAST", "unset": "T_UNSET_CAST",
}

func (l *phpTokens) emit(name, text string) {
	l.out = append(l.out, phpToken{id: phpxTokenConstants[name], text: text, line: l.line})
	l.line += strings.Count(text, "\n")
}

func (l *phpTokens) emitChar(c byte) {
	l.out = append(l.out, phpToken{id: int64(c), text: string(c), line: l.line})
}

func isIdentStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func isIdentByte(c byte) bool { return isIdentStart(c) || c >= '0' && c <= '9' }

func (l *phpTokens) run() {
	for l.pos < len(l.src) {
		l.html()
		if l.pos >= len(l.src) {
			break
		}
		l.php()
	}
}

// html reads inline text up to an open tag, then the tag.
func (l *phpTokens) html() {
	start := l.pos
	for l.pos < len(l.src) {
		if strings.HasPrefix(l.src[l.pos:], "<?=") {
			break
		}
		if strings.HasPrefix(strings.ToLower(l.src[l.pos:min(l.pos+5, len(l.src))]), "<?php") &&
			(l.pos+5 == len(l.src) || strings.IndexByte(" \t\r\n", l.src[l.pos+5]) >= 0) {
			break
		}
		l.pos++
	}
	if l.pos > start {
		l.emit("T_INLINE_HTML", l.src[start:l.pos])
	}
	if l.pos >= len(l.src) {
		return
	}
	if strings.HasPrefix(l.src[l.pos:], "<?=") {
		l.emit("T_OPEN_TAG_WITH_ECHO", "<?=")
		l.pos += 3
		return
	}
	end := l.pos + 5
	if end < len(l.src) {
		if l.src[end] == '\r' && end+1 < len(l.src) && l.src[end+1] == '\n' {
			end += 2
		} else {
			end++
		}
	}
	l.emit("T_OPEN_TAG", l.src[l.pos:end])
	l.pos = end
}

// php reads tokens until a close tag or the end of the source.
func (l *phpTokens) php() {
	for l.pos < len(l.src) {
		if l.step() {
			return
		}
	}
}

// step reads one token. It reports true after a close tag, which ends php mode.
func (l *phpTokens) step() bool {
	{
		c := l.src[l.pos]
		rest := l.src[l.pos:]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			j := l.pos
			for j < len(l.src) && strings.IndexByte(" \t\n\r", l.src[j]) >= 0 {
				j++
			}
			l.emit("T_WHITESPACE", l.src[l.pos:j])
			l.pos = j
		case strings.HasPrefix(rest, "?>"):
			end := l.pos + 2
			if end < len(l.src) && l.src[end] == '\n' {
				end++
			} else if strings.HasPrefix(l.src[end:], "\r\n") {
				end += 2
			}
			l.emit("T_CLOSE_TAG", l.src[l.pos:end])
			l.pos = end
			return true
		case strings.HasPrefix(rest, "#["):
			l.emit("T_ATTRIBUTE", "#[")
			l.pos += 2
		case c == '#' || strings.HasPrefix(rest, "//"):
			j := l.pos
			for j < len(l.src) && l.src[j] != '\n' && l.src[j] != '\r' && !strings.HasPrefix(l.src[j:], "?>") {
				j++
			}
			l.emit("T_COMMENT", l.src[l.pos:j])
			l.pos = j
		case strings.HasPrefix(rest, "/*"):
			j := strings.Index(l.src[l.pos+2:], "*/")
			end := len(l.src)
			if j >= 0 {
				end = l.pos + 2 + j + 2
			}
			text := l.src[l.pos:end]
			name := "T_COMMENT"
			if strings.HasPrefix(text, "/**") && len(text) > 3 && strings.IndexByte(" \t\r\n", text[3]) >= 0 {
				name = "T_DOC_COMMENT"
			}
			l.emit(name, text)
			l.pos = end
		case c == '$' && l.pos+1 < len(l.src) && isIdentStart(l.src[l.pos+1]):
			j := l.pos + 1
			for j < len(l.src) && isIdentByte(l.src[j]) {
				j++
			}
			l.emit("T_VARIABLE", l.src[l.pos:j])
			l.pos = j
			l.afterArrow = false
		case c >= '0' && c <= '9' || c == '.' && l.pos+1 < len(l.src) && l.src[l.pos+1] >= '0' && l.src[l.pos+1] <= '9':
			l.number()
		case c == '\'':
			l.singleQuoted()
		case c == '"':
			l.doubleQuoted()
		case strings.HasPrefix(rest, "<<<"):
			if !l.heredoc() {
				l.emit("T_SL", "<<")
				l.pos += 2
			}
		case isIdentStart(c) || c == '\\' && l.pos+1 < len(l.src) && isIdentStart(l.src[l.pos+1]):
			l.word()
		case c == '(':
			if !l.cast() {
				l.emitChar(c)
				l.pos++
			}
		default:
			l.operator()
		}
	}
	return false
}

func (l *phpTokens) operator() {
	rest := l.src[l.pos:]
	for _, op := range phpOperators {
		if op.name != "" && strings.HasPrefix(rest, op.text) {
			l.emit(op.name, op.text)
			l.pos += len(op.text)
			l.afterArrow = op.name == "T_OBJECT_OPERATOR" || op.name == "T_NULLSAFE_OBJECT_OPERATOR"
			return
		}
	}
	c := rest[0]
	if c == '&' {
		// php 8.1 splits & by what follows it.
		j := l.pos + 1
		for j < len(l.src) && strings.IndexByte(" \t\r\n", l.src[j]) >= 0 {
			j++
		}
		follows := j < len(l.src) && (l.src[j] == '$' || strings.HasPrefix(l.src[j:], "..."))
		name := "T_AMPERSAND_NOT_FOLLOWED_BY_VAR_OR_VARARG"
		if follows {
			name = "T_AMPERSAND_FOLLOWED_BY_VAR_OR_VARARG"
		}
		l.emit(name, "&")
		l.pos++
		return
	}
	l.emitChar(c)
	l.pos++
	l.afterArrow = false
}

func (l *phpTokens) number() {
	j := l.pos
	src := l.src
	isFloat := false
	switch {
	case strings.HasPrefix(src[j:], "0x") || strings.HasPrefix(src[j:], "0X"):
		j += 2
		for j < len(src) && (src[j] == '_' || src[j] >= '0' && src[j] <= '9' || src[j] >= 'a' && src[j] <= 'f' || src[j] >= 'A' && src[j] <= 'F') {
			j++
		}
	case strings.HasPrefix(src[j:], "0o") || strings.HasPrefix(src[j:], "0O"):
		j += 2
		for j < len(src) && (src[j] == '_' || src[j] >= '0' && src[j] <= '7') {
			j++
		}
	case strings.HasPrefix(src[j:], "0b") || strings.HasPrefix(src[j:], "0B"):
		j += 2
		for j < len(src) && (src[j] == '_' || src[j] == '0' || src[j] == '1') {
			j++
		}
	default:
		for j < len(src) && (src[j] == '_' || src[j] >= '0' && src[j] <= '9') {
			j++
		}
		if j < len(src) && src[j] == '.' && (j+1 >= len(src) || src[j+1] != '.') {
			isFloat = true
			j++
			for j < len(src) && (src[j] == '_' || src[j] >= '0' && src[j] <= '9') {
				j++
			}
		}
		if j < len(src) && (src[j] == 'e' || src[j] == 'E') {
			k := j + 1
			if k < len(src) && (src[k] == '+' || src[k] == '-') {
				k++
			}
			if k < len(src) && src[k] >= '0' && src[k] <= '9' {
				isFloat = true
				for k < len(src) && src[k] >= '0' && src[k] <= '9' {
					k++
				}
				j = k
			}
		}
	}
	text := src[l.pos:j]
	name := "T_LNUMBER"
	if isFloat || len(strings.TrimLeft(strings.ReplaceAll(text, "_", ""), "0x")) > 19 && !strings.HasPrefix(text, "0") {
		name = "T_DNUMBER"
	}
	l.emit(name, text)
	l.pos = j
	l.afterArrow = false
}

func (l *phpTokens) singleQuoted() {
	j := l.pos + 1
	for j < len(l.src) && l.src[j] != '\'' {
		if l.src[j] == '\\' {
			j++
		}
		j++
	}
	j = min(j+1, len(l.src))
	l.emit("T_CONSTANT_ENCAPSED_STRING", l.src[l.pos:j])
	l.pos = j
}

// word reads an identifier, a keyword or a qualified name.
func (l *phpTokens) word() {
	start := l.pos
	j := l.pos
	qualified, absolute := false, false
	if l.src[j] == '\\' {
		absolute = true
		j++
	}
	for {
		for j < len(l.src) && isIdentByte(l.src[j]) {
			j++
		}
		if j+1 < len(l.src) && l.src[j] == '\\' && isIdentStart(l.src[j+1]) {
			qualified = true
			j++
			continue
		}
		break
	}
	text := l.src[start:j]
	lower := strings.ToLower(text)
	l.pos = j
	switch {
	case l.afterArrow:
		l.afterArrow = false
		l.emit("T_STRING", text)
	case absolute:
		l.emit("T_NAME_FULLY_QUALIFIED", text)
	case qualified && strings.HasPrefix(lower, "namespace\\"):
		l.emit("T_NAME_RELATIVE", text)
	case qualified:
		l.emit("T_NAME_QUALIFIED", text)
	default:
		name, keyword := phpKeywords[lower]
		if lower == "enum" && !l.enumFollows() {
			keyword = false
		}
		if lower == "yield" {
			if from := l.yieldFrom(); from != "" {
				l.emit("T_YIELD_FROM", text+from)
				l.pos += len(from)
				return
			}
		}
		if keyword {
			l.emit(name, text)
			return
		}
		l.emit("T_STRING", text)
	}
}

func (l *phpTokens) enumFollows() bool {
	j := l.pos
	for j < len(l.src) && strings.IndexByte(" \t\r\n", l.src[j]) >= 0 {
		j++
	}
	if j == l.pos || j >= len(l.src) || !isIdentStart(l.src[j]) {
		return false
	}
	k := j
	for k < len(l.src) && isIdentByte(l.src[k]) {
		k++
	}
	w := strings.ToLower(l.src[j:k])
	return w != "extends" && w != "implements"
}

func (l *phpTokens) yieldFrom() string {
	j := l.pos
	for j < len(l.src) && strings.IndexByte(" \t\r\n", l.src[j]) >= 0 {
		j++
	}
	if j > l.pos && strings.EqualFold(l.src[j:min(j+4, len(l.src))], "from") && (j+4 >= len(l.src) || !isIdentByte(l.src[j+4])) {
		return l.src[l.pos : j+4]
	}
	return ""
}

func (l *phpTokens) cast() bool {
	j := l.pos + 1
	for j < len(l.src) && (l.src[j] == ' ' || l.src[j] == '\t') {
		j++
	}
	k := j
	for k < len(l.src) && isIdentByte(l.src[k]) {
		k++
	}
	name, ok := phpCasts[strings.ToLower(l.src[j:k])]
	if !ok {
		return false
	}
	m := k
	for m < len(l.src) && (l.src[m] == ' ' || l.src[m] == '\t') {
		m++
	}
	if m >= len(l.src) || l.src[m] != ')' {
		return false
	}
	l.emit(name, l.src[l.pos:m+1])
	l.pos = m + 1
	return true
}
