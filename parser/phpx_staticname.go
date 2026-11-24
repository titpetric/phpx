//go:build phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxStaticPropName reads what follows the `$` of `Class::${expr}` and
// `Class::$$name`. The `$` is the property sigil; the rest is evaluated for
// the name, so `$$name` names the property `$name`'s value holds and `$$$name`
// one level further. The braces hold the name expression itself.
func (p *parser) phpxStaticPropName() (model.Expr, error) {
	p.next() // the sigil, the `$` the lexer emitted with no name after it
	if p.isOp("{") {
		p.next()
		name, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.eatOp("}"); err != nil {
			return nil, err
		}
		return name, nil
	}
	if p.cur().kind == tVar && p.cur().val != "" {
		return p.newVar(p.next().val), nil
	}
	return p.phpxParseVarVar()
}
