package parser

import (
	"fmt"
	"sync/atomic"

	"github.com/titpetric/phpx/model"
)

// phpxParseCoalesce parses `a ?? b`, right associative and looser than `||`.
func (p *parser) phpxParseCoalesce() (model.Expr, error) {
	left, err := p.parseBinary(1)
	if err != nil {
		return nil, err
	}
	if !p.isOp("??") {
		return left, nil
	}
	line := p.next().line
	right, err := p.phpxParseCoalesce()
	if err != nil {
		return nil, err
	}
	return phpxCoalesce(left, right, line)
}

// phpxCoalesce lowers `a ?? b` to `isset(a) ? a : b`. A side-effect free left
// side is read twice. A call on the left, alone or under an index or property
// chain, is stored once in a temporary: `($t = f()['k']) !== null ? $t : b`.
// The runtime answers null for a missing key or property, so the read needs no
// isset(), which does not see a variable assigned inside its own expression.
func phpxCoalesce(left, right model.Expr, line int) (model.Expr, error) {
	if repeatableExpr(left) {
		return &model.Ternary{
			Cond: &model.Call{Name: "isset", Args: []model.Expr{left}},
			Then: left,
			Else: right,
		}, nil
	}
	tmp := &model.Var{Name: fmt.Sprintf("__phpx_c%d", phpxTemps.Add(1))}
	store := &model.Parenthesized{X: &model.AssignExpr{Target: tmp, Op: "=", Value: left, Line: line}}
	return &model.Ternary{
		Cond: &model.Binary{Op: "!==", Left: store, Right: &model.Lit{Value: nil, Raw: "null"}},
		Then: tmp,
		Else: right,
	}, nil
}

var phpxTemps atomic.Int64
