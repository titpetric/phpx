package parser

import (
	"fmt"

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

// phpxCoalesce lowers `a ?? b` to `isset(a) ? a : b`. The left side is read
// twice, so it must be free of side effects; a call on the left is refused.
func phpxCoalesce(left, right model.Expr, line int) (model.Expr, error) {
	if !repeatableExpr(left) {
		return nil, fmt.Errorf("line %d: ?? needs a variable, index or property on its left", line)
	}
	return &model.Ternary{
		Cond: &model.Call{Name: "isset", Args: []model.Expr{left}},
		Then: left,
		Else: right,
	}, nil
}
