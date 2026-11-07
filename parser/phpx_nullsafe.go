package parser

import (
	"fmt"
	"sync/atomic"

	"github.com/titpetric/phpx/model"
)

// phpxGuard is one `?->` of a postfix chain: the temporary that holds the
// value before it and the expression that produced that value.
type phpxGuard struct {
	tmp  *model.Var
	orig model.Expr
}

// phpxTemps numbers the temporaries the lowerings bind.
var phpxTemps atomic.Int64

// phpxNullsafe records a `?->` guard on the chain and answers the temporary
// the rest of the chain reads from. The caller's `->` branch consumes the
// operator, so a `?->` member parses as an ordinary member of the temporary.
func (p *parser) phpxNullsafe(guards *[]phpxGuard, e model.Expr) model.Expr {
	tmp := &model.Var{Name: fmt.Sprintf("__phpx_n%d", phpxTemps.Add(1))}
	*guards = append(*guards, phpxGuard{tmp: tmp, orig: e})
	return tmp
}

// phpxWrapGuards lowers `a?->b->c` to `($t = a) === null ? null : $t->b->c`.
// The whole rest of the chain sits in the else arm, so a null short-circuits
// it as php does; a second `?->` nests inside the first's else arm.
func phpxWrapGuards(guards []phpxGuard, e model.Expr) model.Expr {
	for i := len(guards) - 1; i >= 0; i-- {
		g := guards[i]
		store := &model.Parenthesized{X: &model.AssignExpr{Target: g.tmp, Op: "=", Value: g.orig}}
		e = &model.Ternary{
			Cond: &model.Binary{Op: "===", Left: store, Right: &model.Lit{Value: nil, Raw: "null"}},
			Then: &model.Lit{Value: nil, Raw: "null"},
			Else: e,
		}
	}
	return e
}
