package parser

import (
	"github.com/titpetric/phpx/model"
)

// A function whose body holds `yield` is a generator. The parser lowers it:
//
//	function f($a) { yield $a; }
//
// becomes `function f($a) { return __phpx_generator(function () use ($a) {
// yield... }); }`, and each yield is the call __phpx_yield(key, value) or
// __phpx_yield_from(source) inside that inner closure. The binding runs the
// closure on its own goroutine and hands control back and forth, so the body
// resumes where it stopped.

// phpxYield parses what follows the `yield` keyword.
func (p *parser) phpxYield() (model.Expr, error) {
	p.sawYield = true
	if p.isKw("from") {
		p.next()
		src, err := p.parseAssign()
		if err != nil {
			return nil, err
		}
		return &model.Call{Name: "__phpx_yield_from", Args: []model.Expr{src}}, nil
	}
	none := p.newLit(nil)
	if p.isOp(";") || p.isOp(")") || p.isOp(",") || p.isOp("]") {
		return &model.Call{Name: "__phpx_yield", Args: []model.Expr{none, none, p.newLit(false)}}, nil
	}
	first, err := p.parseTernary()
	if err != nil {
		return nil, err
	}
	if p.isOp("=>") {
		p.next()
		val, err := p.parseTernary()
		if err != nil {
			return nil, err
		}
		return &model.Call{Name: "__phpx_yield", Args: []model.Expr{first, val, p.newLit(true)}}, nil
	}
	return &model.Call{Name: "__phpx_yield", Args: []model.Expr{none, first, p.newLit(false)}}, nil
}

// phpxGeneratorBody wraps a body that yielded. outer is the sawYield flag of
// the function being parsed around this one, restored either way.
func (p *parser) phpxGeneratorBody(params []model.Param, captured []model.ClosureUse, body []model.Stmt, outer bool) []model.Stmt {
	yielded := p.sawYield
	p.sawYield = outer
	if !yielded {
		return body
	}
	// The inner closure sees the parameters and what the outer closure captured.
	uses := append(make([]model.ClosureUse, 0, len(params)+len(captured)), captured...)
	for _, pr := range params {
		uses = append(uses, model.ClosureUse{Name: pr.Name})
	}
	inner := &model.Closure{Uses: uses, Body: body}
	call := &model.Call{Name: "__phpx_generator", Args: []model.Expr{inner}}
	return []model.Stmt{&model.Return{Value: call}}
}

// phpxAssignTail finishes an assignment whose target a prefix operator has
// already parsed: `!$x = f()`, `!$x ??= f()`.
func (p *parser) phpxAssignTail(target model.Expr) (model.Expr, error) {
	t := p.next()
	right, err := p.parseAssign()
	if err != nil {
		return nil, err
	}
	if t.val == "??=" {
		co, err := phpxCoalesce(target, right, t.line)
		if err != nil {
			return nil, err
		}
		return p.newAssignExpr(target, "=", co, t.line), nil
	}
	return p.newAssignExpr(target, t.val, right, t.line), nil
}
