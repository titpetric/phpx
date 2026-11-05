//go:build phpx

package parser

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxGotoStmt reads `goto name;` and the label `name:`. A label is an
// identifier followed by a single colon at the start of a statement; the
// keywords whose colon belongs to the alternative syntax or to a switch are
// not labels.
func (p *parser) phpxGotoStmt(t token) (model.Stmt, bool) {
	if strings.EqualFold(t.val, "goto") && p.peek(1).kind == tIdent {
		p.next()
		name := p.next().val
		p.optSemi()
		return &model.Goto{Label: name, Line: t.line}, true
	}
	next := p.peek(1)
	if next.kind == tOp && next.val == ":" {
		switch strings.ToLower(t.val) {
		case "default", "else", "case", "parent", "self", "static":
			return nil, false
		}
		p.next()
		p.next()
		return &model.Label{Name: t.val, Line: t.line}, true
	}
	return nil, false
}
