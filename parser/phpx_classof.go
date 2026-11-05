package parser

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxClassOf consumes `::class` after an expression and reports whether it
// did. `$obj::class` lowers to `get_class($obj)`, which is what php answers
// for an object; a literal `Name::class` never reaches it because the name
// parser reads that form first.
func (p *parser) phpxClassOf() bool {
	if p.isOp("::") && p.peek(1).val == "class" {
		p.next()
		p.next()
		return true
	}
	return false
}

// phpxStaticOnValue parses what follows `::` on a value that holds a class
// name or an object: `$cls::NAME`, `$cls::$prop`, `$cls::method(...)` and
// `$cls::$method(...)`. Each lowers to a call that takes the class from
// __phpx_classref, since the static nodes carry a class name and no expression.
func (p *parser) phpxStaticOnValue(base model.Expr) (model.Expr, error) {
	p.next() // ::
	class := &model.Call{Name: "__phpx_classref", Args: []model.Expr{base}}
	t := p.cur()
	switch {
	case p.isOp("{"):
		// `$cls::{expr}(...)` names the method by an expression.
		p.next()
		name, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.eatOp("}"); err != nil {
			return nil, err
		}
		return p.phpxCallStatic(base, name)
	case t.kind == tVar && p.isOpAt(1, "("):
		p.next()
		return p.phpxCallStatic(base, &model.Var{Name: t.val})
	case t.kind == tVar:
		p.next()
		return &model.Call{Name: "__phpx_static_prop", Args: []model.Expr{class, p.newLit(t.val)}}, nil
	case t.kind == tIdent && p.isOpAt(1, "("):
		p.next()
		return p.phpxCallStatic(base, p.newLit(t.val))
	case t.kind == tIdent:
		p.next()
		name := &model.Binary{Op: ".", Left: class, Right: p.newLit("::" + t.val)}
		return &model.Call{Name: "constant", Args: []model.Expr{name}}, nil
	}
	return nil, fmt.Errorf("line %d: unexpected %s after ::", t.line, t)
}

// phpxCallStatic builds call_user_func([base, method], args...).
func (p *parser) phpxCallStatic(base, method model.Expr) (model.Expr, error) {
	args, _, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}
	target := &model.ArrayLit{Items: []model.ArrayItem{{Val: base}, {Val: method}}}
	return &model.Call{Name: "call_user_func", Args: append([]model.Expr{target}, args...)}, nil
}

// phpxAssignStaticProp turns `$cls::$prop = v` and `$cls::$prop .= v` into
// the setter form of __phpx_static_prop, which reads when it gets no value.
func (p *parser) phpxAssignStaticProp(get *model.Call) (model.Expr, error) {
	op := p.next().val
	value, err := p.parseAssign()
	if err != nil {
		return nil, err
	}
	if op != "=" {
		value = &model.Binary{Op: strings.TrimSuffix(op, "="), Left: get, Right: value}
	}
	args := append(append([]model.Expr{}, get.Args...), value, p.newLit(true))
	return &model.Call{Name: get.Name, Args: args}, nil
}
