package expr

import (
	"github.com/titpetric/phpx/model"
)

// phpxAssignExpr compiles an assignment used as an expression whose target is
// not a plain variable, or whose operator is compound: `self::$i = new self`,
// `$a['k'] = v`, `$o->p .= 'x'`. The write goes through the runtime's own
// assignment, `__phpx_assign`, installed beside `__set` by the environment.
// It reports false for a `$var =` assignment, which upstream compiles.
func (dc *directCompiler) phpxAssignExpr(e model.Expr) (closure, bool, error) {
	n, ok := e.(*model.AssignExpr)
	if !ok {
		return nil, false, nil
	}
	if _, plain := model.UnwrapParenthesized(n.Target).(*model.Var); plain && (n.Op == "=" || n.Op == "") {
		return nil, false, nil
	}
	val, err := dc.compile(n.Value)
	if err != nil {
		return nil, true, err
	}
	return baseCall("__phpx_assign", []closure{constClosure(n), val}), true, nil
}
