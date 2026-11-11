package expr

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// phpxSettypeArgs rewrites the arguments of `settype($var, "type")` into
// (setter, current value, type): the binding writes the variable through the
// setter and reads its value from the middle one.
func (dc *directCompiler) phpxSettypeArgs(n *model.Call, args []closure) ([]closure, error) {
	if n.Name != "settype" && n.Fallback != "settype" || len(args) != 2 {
		return args, nil
	}
	if _, ok := modelutil.UnwrapParenthesized(n.Args[0]).(*model.Var); !ok {
		return args, nil
	}
	cur, err := dc.compile(n.Args[0])
	if err != nil {
		return nil, err
	}
	return []closure{args[0], cur, args[1]}, nil
}
