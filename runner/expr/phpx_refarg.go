package expr

import (
	"github.com/titpetric/phpx/modelutil"
	"github.com/titpetric/phpx/model"
)

// phpxRefArg compiles an lvalue argument (a variable, or a property of a
// variable) to a *model.RefArg once the program declares a by-reference
// parameter somewhere. It declines when no such parameter exists or the
// environment carries no __phpx_refarg helper, and the argument then compiles
// as a value.
func (dc *directCompiler) phpxRefArg(a model.Expr, fn ...string) (closure, bool) {
	if !model.PhpxAnyByRef.Load() {
		return nil, false
	}
	switch x := modelutil.UnwrapParenthesized(a).(type) {
	case *model.Var:
		if x.Name == "this" {
			return nil, false
		}
	case *model.PropAccess:
		if _, ok := modelutil.UnwrapParenthesized(x.Base).(*model.Var); !ok || x.NameExpr != nil {
			return nil, false
		}
	default:
		return nil, false
	}
	plain, err := dc.compile(a)
	if err != nil {
		return nil, false
	}
	node := modelutil.UnwrapParenthesized(a)
	return func(env *Env) (any, error) {
		call, ok := env.Base["__phpx_refarg"].(func(...any) (any, error))
		if !ok {
			return plain(env)
		}
		name, fallback := "", ""
		if len(fn) == 2 {
			name, fallback = fn[0], fn[1]
		}
		return call(node, name, fallback)
	}, true
}
