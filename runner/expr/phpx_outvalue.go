//go:build phpx

package expr

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// phpxOutValueArgs passes an output parameter that the callee also reads as
// two arguments: the setter that writes the variable, and the variable's
// current value. openssl_decrypt reads $tag to verify it, where a binding that
// takes the parameter by reference alone would see only the write.
func (dc *directCompiler) phpxOutValueArgs(n *model.Call, args []closure) ([]closure, error) {
	position := -1
	switch baseCallName(n.Name) {
	case "openssl_encrypt", "openssl_decrypt":
		position = 5
	}
	if position < 0 || len(n.Args) <= position || len(args) <= position {
		return args, nil
	}
	if _, ok := modelutil.UnwrapParenthesized(n.Args[position]).(*model.Var); !ok {
		return args, nil
	}
	cur, err := dc.compile(n.Args[position])
	if err != nil {
		return nil, err
	}
	out := make([]closure, 0, len(args)+1)
	out = append(out, args[:position+1]...)
	out = append(out, cur)
	out = append(out, args[position+1:]...)
	return out, nil
}
