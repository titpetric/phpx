package parser

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxMatch parses `match (subject) { a, b => x, default => y }`. The arms
// lower to a chain of conditionals over a temporary that holds the subject,
// compared with ===, and an arm-less end raises UnhandledMatchError:
//
//	($t = subject) === a || $t === b ? x : (... ? ... : __phpx_unhandled_match($t))
func (p *parser) phpxMatch() (model.Expr, error) {
	line := p.cur().line // the keyword is already consumed
	if err := p.eatOp("("); err != nil {
		return nil, err
	}
	subject, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if err := p.eatOp(")"); err != nil {
		return nil, err
	}
	if err := p.eatOp("{"); err != nil {
		return nil, err
	}
	type arm struct {
		conds  []model.Expr
		result model.Expr
	}
	var arms []arm
	var fallback model.Expr
	for !p.isOp("}") {
		var a arm
		if strings.EqualFold(p.cur().val, "default") && p.cur().kind == tIdent && (p.isOpAt(1, "=>") || p.isOpAt(1, ",")) {
			p.next()
			if p.isOp(",") {
				p.next()
			}
		} else {
			for {
				c, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				a.conds = append(a.conds, c)
				if !p.isOp(",") {
					break
				}
				p.next()
				if p.isOp("=>") { // a trailing comma in the condition list
					break
				}
			}
		}
		if err := p.eatOp("=>"); err != nil {
			return nil, err
		}
		a.result, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		if a.conds == nil {
			fallback = a.result
		} else {
			arms = append(arms, a)
		}
		if p.isOp(",") {
			p.next()
		}
	}
	if err := p.eatOp("}"); err != nil {
		return nil, err
	}
	tmp := phpxTemp("m")
	store := &model.Parenthesized{X: &model.AssignExpr{Target: tmp, Op: "=", Value: subject, Line: line}}
	var tail model.Expr = fallback
	if tail == nil {
		tail = &model.Call{Name: "__phpx_unhandled_match", Args: []model.Expr{tmp}}
	}
	stored := false
	chain := tail
	for i := len(arms) - 1; i >= 0; i-- {
		var cond model.Expr
		for _, c := range arms[i].conds {
			var left model.Expr = tmp
			if i == 0 && cond == nil {
				left, stored = store, true
			}
			eq := p.newBinary("===", left, c)
			if cond == nil {
				cond = eq
			} else {
				cond = p.newBinary("||", cond, eq)
			}
		}
		chain = &model.Ternary{Cond: cond, Then: arms[i].result, Else: chain}
	}
	if !stored {
		// No arm to carry the store: evaluate the subject for its effects.
		chain = &model.Index{Base: &model.ArrayLit{Items: []model.ArrayItem{{Val: store}, {Val: chain}}}, Index: &model.Lit{Value: int64(1), Raw: "1"}}
	}
	return chain, nil
}

// phpxThrowExpr parses `throw expr` in expression position, as in
// `$x ?? throw new E` and `fn() => throw ...`: the call __phpx_throw(expr).
func (p *parser) phpxThrowExpr() (model.Expr, error) {
	x, err := p.parseAssign()
	if err != nil {
		return nil, fmt.Errorf("throw: %w", err)
	}
	return &model.Call{Name: "__phpx_throw", Args: []model.Expr{x}}, nil
}

// phpxClone parses `clone expr`, which binds tighter than any binary operator
// and looser than `->`: the call __phpx_clone(expr).
func (p *parser) phpxClone() (model.Expr, error) {
	x, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	return &model.Call{Name: "__phpx_clone", Args: []model.Expr{x}}, nil
}
