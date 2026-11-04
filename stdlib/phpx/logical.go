//go:build phpx

package phpx

import (
	"fmt"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxXor)
	runner.RegisterBinding(registerPhpxPrint)
}

// registerPhpxXor installs the function the parser lowers `a xor b` to.
func registerPhpxXor(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_xor", func(a, b any) bool {
		return phpval.Truthy(a) != phpval.Truthy(b)
	})
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
