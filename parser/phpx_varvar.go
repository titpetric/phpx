package parser

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// The names of the two calls a variable variable becomes. The runtime compiles
// them against the scope, which a function cannot reach.
const (
	phpxVarVarGet = "__phpx_varvar"
	phpxVarVarSet = "__phpx_varvar_set"
)

// phpxParseVarVar reads what follows the `$` of `$$name`, `$$$name` or
// `${expr}` and answers the read of the variable that name holds.
func (p *parser) phpxParseVarVar() (model.Expr, error) {
	var name model.Expr
	switch {
	case p.cur().kind == tVar:
		t := p.next()
		if t.val == "" {
			inner, err := p.phpxParseVarVar()
			if err != nil {
				return nil, err
			}
			name = inner
		} else {
			name = p.newVar(t.val)
		}
	case p.isOp("{"):
		p.next()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if err := p.eatOp("}"); err != nil {
			return nil, err
		}
		name = e
	default:
		return nil, fmt.Errorf("line %d: expected a variable name after $", p.cur().line)
	}
	return &model.Call{Name: phpxVarVarGet, Args: []model.Expr{name}}, nil
}

// phpxAssignVarVar turns `$$name = v` and `$$name .= v` into the setter call.
func (p *parser) phpxAssignVarVar(get *model.Call) (model.Expr, error) {
	op := p.next().val
	value, err := p.parseAssign()
	if err != nil {
		return nil, err
	}
	if op != "=" {
		value = &model.Binary{Op: strings.TrimSuffix(op, "="), Left: get, Right: value}
	}
	return &model.Call{Name: phpxVarVarSet, Args: []model.Expr{get.Args[0], value}}, nil
}
