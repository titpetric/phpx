package parser

import (
	"github.com/titpetric/phpx/model"
)

// phpxBinary builds the binary node for op. `a <=> b` becomes the call
// `__phpx_spaceship(a, b)`, so both operands evaluate once and neither engine
// needs an operator of its own; every other operator is upstream's node.
func (p *parser) phpxBinary(op string, left, right model.Expr) model.Expr {
	if op == "<=>" {
		return &model.Call{Name: "__phpx_spaceship", Args: []model.Expr{left, right}}
	}
	return p.newBinary(op, left, right)
}
