//go:build phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxNewStaticProp reads `Class::$prop` after the class name of a `new`: the
// static property holds the name of the class to build, as in
// `new static::$responseFactoryClass()`.
func (p *parser) phpxNewStaticProp(class string) (model.Expr, bool) {
	if !p.isOp("::") || p.peek(1).kind != tVar {
		return nil, false
	}
	p.next() // ::
	name := p.next().val
	return &model.StaticProp{Class: class, Name: name}, true
}
