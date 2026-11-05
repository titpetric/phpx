package parser

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxQualifiedLit answers the literal for `\true`, `\false` and `\null`.
// php reads the fully qualified spelling as the same constant, and generated
// code uses it (google/apiclient's aliases.php guards a block with `\false`).
// A name followed by `::` or `(` is a class or a function and is left alone.
func (p *parser) phpxQualifiedLit(name string) (model.Expr, bool) {
	if p.isOp("::") || p.isOp("(") {
		return nil, false
	}
	switch strings.ToLower(name) {
	case "true":
		return p.newLit(true), true
	case "false":
		return p.newLit(false), true
	case "null":
		return p.newLit(nil), true
	}
	return nil, false
}
