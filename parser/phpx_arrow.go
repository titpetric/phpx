package parser

import (
	"errors"

	"github.com/titpetric/phpx/model"
)

var errNotArrow = errors.New("not an arrow function")

// phpxParseArrow parses `fn(params) [: T] => expr` after the `fn` keyword. The
// body is `return expr;`. php captures the enclosing scope by value on demand,
// so the closure's capture list is every variable the body names that is not a
// parameter; naming one the enclosing frame lacks captures null.
func (p *parser) phpxParseArrow() (*model.Closure, error) {
	byRef := false
	if p.isOp("&") {
		p.next()
		byRef = true
	}
	params, err := p.parseParams()
	if err != nil {
		// `fn(1)` is a call to a function named fn, not a parameter list.
		return nil, errNotArrow
	}
	returnType := p.parseReturnType()
	if !p.isOp("=>") {
		return nil, errNotArrow
	}
	p.next()
	start := p.i
	body, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	bound := map[string]bool{"this": true}
	for _, pr := range params {
		bound[pr.Name] = true
	}
	var uses []model.ClosureUse
	for _, t := range p.toks[start:p.i] {
		if t.kind == tVar && !bound[t.val] {
			bound[t.val] = true
			uses = append(uses, model.ClosureUse{Name: t.val})
		}
	}
	ret := &model.Return{Value: body}
	return &model.Closure{Params: params, Uses: uses, Body: []model.Stmt{ret}, ReturnType: returnType, ByRef: byRef}, nil
}

// phpxIsArrow reports whether the current `fn` token starts an arrow function.
func (p *parser) phpxIsArrow() bool {
	return p.isOp("(") || p.isOp("&")
}

// phpxClosureOf parses the closure after the keyword kw. `fn` followed by a
// parameter list and `=>` is an arrow function; anything else, `fn()` as a
// plain function name included, falls back to upstream's reading.
func (p *parser) phpxClosureOf(kw string) (*model.Closure, error) {
	if kw == "fn" && p.phpxIsArrow() {
		at := p.i
		cl, err := p.phpxParseArrow()
		if !errors.Is(err, errNotArrow) {
			return cl, err
		}
		p.i = at
	}
	return p.parseClosure()
}
