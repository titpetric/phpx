package parser

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// namedArgOp prefixes the operator of the Unary that carries a named argument:
// `f(limit: 3)` holds `Unary{Op: "named:limit", X: 3}`. The runtime binds it to
// the parameter of that name.
const namedArgOp = "named:"

// phpxNamedArg consumes the `name:` that opens a named argument and returns
// the name. An identifier followed by `::` is a static access and is left alone,
// since the lexer reads `::` as one token.
func (p *parser) phpxNamedArg() (string, bool) {
	t := p.cur()
	if t.kind != tIdent || !p.isOpAt(1, ":") {
		return "", false
	}
	p.i += 2
	return t.val, true
}

// phpxNamedValue wraps the value of a named argument.
func phpxNamedValue(name string, e model.Expr) model.Expr {
	return &model.Unary{Op: namedArgOp + name, X: e}
}

// phpxSpreadItem consumes the `...` of an array literal element.
func (p *parser) phpxSpreadItem() bool {
	if p.isOp(".") && p.isOpAt(1, ".") && p.isOpAt(2, ".") {
		p.i += 3
		return true
	}
	return false
}

// phpxIsNamedOp reports a Unary operator that spells a named argument.
func phpxIsNamedOp(op string) bool { return strings.HasPrefix(op, namedArgOp) }
