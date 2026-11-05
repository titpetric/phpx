//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/internal/phpval"
)

// phpxInstallVarVar adds __phpx_varvar(name) and __phpx_varvar_set(name, value),
// the scope reads and writes behind `$$name`.
func (rt *Runtime) phpxInstallVarVar(env map[string]any, ref *scopeRef) {
	env["__phpx_varvar"] = adapt(func(name any) (any, error) {
		v, _ := ref.scope.Get(phpval.String(name))
		return v, nil
	})
	env["__phpx_varvar_set"] = adapt(func(name, value any) (any, error) {
		rt.setVar(ref.scope, phpval.String(name), value)
		return value, nil
	})
}
