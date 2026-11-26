//go:build phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxNewStaticProp reads `static::$class` and the index accessors after it, so
// `new static::$classes[static::class]()` names the class in the property
// rather than the property itself. It reports false when the tokens do not
// spell that form, leaving them to the caller.
func (p *parser) phpxNewStaticProp(class string) (model.Expr, bool, error) {
	if !p.isOp("::") || p.peek(1).kind != tVar {
		return nil, false, nil
	}
	p.next() // ::
	name := p.next().val
	var expr model.Expr = &model.StaticProp{Class: class, Name: name}
	for p.isOp("[") {
		p.next()
		if p.isOp("]") {
			// `$x[] =` names no class; the caller reports the syntax.
			return nil, false, nil
		}
		index, err := p.parseExpr()
		if err != nil {
			return nil, false, err
		}
		if err := p.eatOp("]"); err != nil {
			return nil, false, err
		}
		expr = &model.Index{Base: expr, Index: index}
	}
	return expr, true, nil
}
