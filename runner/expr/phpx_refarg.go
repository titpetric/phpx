package expr

import "github.com/titpetric/phpx/model"

// phpxRefArg compiles an lvalue argument (a variable, or a property of a
// variable) to a *model.RefArg once the program declares a by-reference
// parameter somewhere. It declines when no such parameter exists or the
// environment carries no __phpx_refarg helper, and the argument then compiles
// as a value.
func (dc *directCompiler) phpxRefArg(a model.Expr) (closure, bool) {
	if !model.PhpxAnyByRef.Load() {
		return nil, false
	}
	switch x := model.UnwrapParenthesized(a).(type) {
	case *model.Var:
		if x.Name == "this" {
			return nil, false
		}
	case *model.PropAccess:
		if _, ok := model.UnwrapParenthesized(x.Base).(*model.Var); !ok || x.NameExpr != nil {
			return nil, false
		}
	default:
		return nil, false
	}
	plain, err := dc.compile(a)
	if err != nil {
		return nil, false
	}
	node := model.UnwrapParenthesized(a)
	return func(env *Env) (any, error) {
		fn, ok := env.Base["__phpx_refarg"].(func(...any) (any, error))
		if !ok {
			return plain(env)
		}
		return fn(node)
	}, true
}
