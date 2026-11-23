package expr

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// phpxSplitArg compiles an argument that a builtin writes in place. The
// environment's __phpx_splitbase resolves the argument's container and splits
// an array shared with a copy along the path, so the binding writes the array
// the script named. A build without the helper compiles the argument as a
// value.
func (dc *directCompiler) phpxSplitArg(a model.Expr) (closure, bool) {
	plain, err := dc.compile(a)
	if err != nil {
		return nil, false
	}
	node := modelutil.UnwrapParenthesized(a)
	return func(env *Env) (any, error) {
		call, ok := env.Base["__phpx_splitbase"].(func(...any) (any, error))
		if !ok {
			return plain(env)
		}
		return call(node)
	}, true
}
