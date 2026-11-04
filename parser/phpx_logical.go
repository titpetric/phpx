package parser

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxParseLogical parses the word operators `or`, `xor` and `and`, which bind
// looser than assignment: `$x = f() or die()` assigns first. `and` and `or`
// are `&&` and `||`; `xor` is `__phpx_xor(a, b)`, a call so both operands run.
func (p *parser) phpxParseLogical(level int) (model.Expr, error) {
	if level == 3 {
		return p.parseAssign()
	}
	left, err := p.phpxParseLogical(level + 1)
	if err != nil {
		return nil, err
	}
	for {
		t := p.cur()
		if t.kind != tIdent || phpxLogicalLevel(strings.ToLower(t.val)) != level {
			return left, nil
		}
		op := strings.ToLower(p.next().val)
		right, err := p.phpxParseLogical(level + 1)
		if err != nil {
			return nil, err
		}
		switch op {
		case "xor":
			left = &model.Call{Name: "__phpx_xor", Args: []model.Expr{left, right}}
		case "and":
			left = p.newBinary("&&", left, right)
		default:
			left = p.newBinary("||", left, right)
		}
	}
}

// phpxLogicalLevel orders the word operators: 0 is `or`, 1 `xor`, 2 `and`.
func phpxLogicalLevel(op string) int {
	switch op {
	case "or":
		return 0
	case "xor":
		return 1
	case "and":
		return 2
	}
	return -1
}

// phpxForList reads the comma separated expressions of a for clause, which php
// evaluates in order. One expression is the statement upstream parses; more are
// the sequence `[(a), (b)][0]`.
func (p *parser) phpxForList() (model.Stmt, error) {
	first, err := p.parseSimpleStmt()
	if err != nil || !p.isOp(",") {
		return first, err
	}
	steps := []model.ArrayItem{{Val: phpxStmtExpr(first)}}
	for p.isOp(",") {
		p.next()
		next, err := p.parseSimpleStmt()
		if err != nil {
			return nil, err
		}
		steps = append(steps, model.ArrayItem{Val: phpxStmtExpr(next)})
	}
	seq := &model.Index{Base: &model.ArrayLit{Items: steps}, Index: &model.Lit{Value: int64(0), Raw: "0"}}
	return p.newExprStmt(seq), nil
}

// phpxStmtExpr turns a for clause statement back into an expression.
func phpxStmtExpr(s model.Stmt) model.Expr {
	switch n := s.(type) {
	case *model.Assign:
		// An assignment expression takes `=` alone: `$i += 4` is `$i = $i + 4`.
		if n.Op != "=" {
			return &model.Parenthesized{X: &model.AssignExpr{Target: n.Target, Op: "=", Value: &model.Binary{Op: strings.TrimSuffix(n.Op, "="), Left: n.Target, Right: n.Value}}}
		}
		return &model.Parenthesized{X: &model.AssignExpr{Target: n.Target, Op: n.Op, Value: n.Value}}
	case *model.ExprStmt:
		return &model.Parenthesized{X: n.X}
	}
	return &model.Lit{Value: nil, Raw: "null"}
}

// phpxForCond reads the condition of a for, which php lets stay empty.
func (p *parser) phpxForCond() (model.Expr, error) {
	if p.isOp(";") {
		return &model.Lit{Value: true, Raw: "true"}, nil
	}
	return p.parseExpr()
}

// phpxPrint parses `print expr`, an expression that answers 1: it is the call
// `__phpx_print(expr)`, which writes the string form of its argument.
func (p *parser) phpxPrint() (model.Expr, error) {
	x, err := p.parseAssign()
	if err != nil {
		return nil, err
	}
	return &model.Call{Name: "__phpx_print", Args: []model.Expr{x}}, nil
}
