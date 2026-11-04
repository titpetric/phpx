//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxXor)
}

// registerPhpxXor installs the function the parser lowers `a xor b` to.
func registerPhpxXor(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_xor", func(a, b any) bool {
		return phpval.Truthy(a) != phpval.Truthy(b)
	})
}
