//go:build phpx

package phpx

import (
	"fmt"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Logical\Print`, registerPhpxPrint)
	runner.RegisterArea(`Runtime\PX\Logical\Match`, registerPhpxMatch)
}

// registerPhpxPrint installs the function the parser lowers `print x` to.
func registerPhpxPrint(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_print", func(v any) (int64, error) {
		s, err := rt.InvokeNamed("strval", v)
		if err != nil {
			return 0, err
		}
		_, err = rt.Output().Write([]byte(fmt.Sprint(s)))
		return 1, err
	})
}

// unhandledMatch is the Error a match with no arm for its subject raises.
type unhandledMatch struct{ message string }

func (e *unhandledMatch) Error() string          { return e.message }
func (e *unhandledMatch) ThrowableClass() string { return "UnhandledMatchError" }

// registerPhpxMatch installs what the parser lowers `throw` as an expression and
// the end of a match to.
func registerPhpxMatch(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_clone", func(v any) (any, error) { return rt.PhpxClone(v) })
	rt.RegisterFunc("__phpx_throw", func(v any) error { return runner.PhpxThrowValue(v) })
	rt.RegisterFunc("__phpx_unhandled_match", func(v any) error {
		// php 8.4 names the type of the subject and not its value.
		shown := "of type " + map[string]string{"string": "string", "int64": "int", "int": "int", "float64": "float",
			"bool": "bool", "<nil>": "null", "*model.Array": "array"}[fmt.Sprintf("%T", v)]
		if obj, ok := v.(*model.Object); ok && obj.Class != nil {
			shown = "of type " + obj.Class.Name
		}
		return &unhandledMatch{message: "Unhandled match case " + shown}
	})
}
