//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxInstallEnv adds the helpers the fork's compiled expressions call.
func (rt *Runtime) phpxInstallEnv(env map[string]any, ref *scopeRef) {
	env["__phpx_assign"] = adapt(rt.phpxHelperAssign(ref))
	rt.phpxInstallRefAssign(env, ref)
	rt.phpxInstallRefArg(env, ref)
	rt.phpxInstallVarVar(env, ref)
	rt.phpxInstallTraverse(env)
	rt.phpxInstallSPL(env)
}

// phpxHelperAssign writes an assignment expression's target: the value is
// already evaluated, a compound operator reads the target's current value
// first, and the result is the stored value.
func (rt *Runtime) phpxHelperAssign(ref *scopeRef) func(node any, val any) (any, error) {
	return func(node any, val any) (any, error) {
		n := node.(*model.AssignExpr)
		if n.Op != "=" && n.Op != "" {
			cur, err := rt.readLValue(n.Target, ref.scope)
			if err != nil {
				return nil, err
			}
			if val, err = applyAssignOp(n.Op, cur, val); err != nil {
				return nil, err
			}
		}
		if err := rt.assignTo(n.Target, val, ref.scope); err != nil {
			return nil, err
		}
		return val, nil
	}
}
