//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
)

// phpxVarVarName is the call name the parser gives a variable variable.
const phpxVarVarName = "__phpx_varvar"

// phpxVarVarUnset unsets the variable a lowered `$$name` names.
func (rt *Runtime) phpxVarVarUnset(call *model.Call, scope *Scope) (bool, error) {
	if call.Name != phpxVarVarName || len(call.Args) != 1 {
		return false, nil
	}
	key, err := rt.Eval(call.Args[0], scope)
	if err != nil {
		return true, err
	}
	name := phpval.String(key)
	if old, ok := scope.Get(name); ok {
		rt.phpxNoteDropped(old)
	}
	scope.Unset(name)
	return true, nil
}

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
