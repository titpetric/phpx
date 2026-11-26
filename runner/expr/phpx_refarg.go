package expr

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// refCellArg compiles a `&$x` array item to the lvalue's cell.
func (dc *directCompiler) refCellArg(a model.Expr) (closure, bool) {
	switch modelutil.UnwrapParenthesized(a).(type) {
	case *model.Var, *model.PropAccess, *model.Index:
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
		v, err := call(node, "", "")
		if err != nil {
			return nil, err
		}
		if ra, ok := v.(*model.RefArg); ok && ra.Make != nil {
			return ra.Make(), nil
		}
		return v, nil
	}, true
}

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
	case *model.Index:
		// An element of an array names a slot a reference parameter can bind,
		// as `$this->arrays[] = &$array` relies on. phpxCellOf makes the cell
		// only when the base answers an array; anything else is a value.
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
