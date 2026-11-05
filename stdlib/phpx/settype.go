//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerSettype)
}

// registerSettype binds settype(). The compiler passes it the variable's setter,
// the variable's current value and the type name, in that order, because a
// binding cannot read through an output parameter.
func registerSettype(rt *runner.Runtime) {
	rt.RegisterFunc("settype", settype)
}

// settype converts cur to the named type and hands the result to set.
func settype(set func(any), cur any, typ string) bool {
	switch strings.ToLower(typ) {
	case "int", "integer", "float", "double", "string", "bool", "boolean", "array":
		set(runner.PhpxCast(strings.ToLower(typ), cur))
	case "null":
		set(nil)
	default:
		return false
	}
	return true
}
