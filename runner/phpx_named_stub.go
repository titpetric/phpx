//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxNamedArgs binds nothing by name untagged.
func (rt *Runtime) phpxNamedArgs(_ *model.FuncDecl, args []any, _ *Scope) ([]any, error) {
	return args, nil
}

// phpxBareArgs leaves the arguments alone untagged.
func phpxBareArgs(args []any) []any { return args }

// phpxEntryArgs leaves the arguments alone untagged.
func phpxEntryArgs(args []any, _ any) []any { return args }

// phpxArgsValue answers the frame's own slice untagged.
func phpxArgsValue(args []any) any { return args }

// phpxOwnedArray keeps a Go slice as upstream does.
func phpxOwnedArray(v any) any { return v }

// phpxOwned keeps a Go slice as upstream does.
func (rt *Runtime) phpxOwned(v any) any { return v }

// phpxVariadicAppend appends the argument untagged.
func phpxVariadicAppend(rest *model.Array, arg any) { rest.Append(arg) }

// phpxCopyRead shares the array untagged.
func phpxCopyRead(_ model.Expr, v any) any { return v }

// phpxBindArray binds the argument as it came untagged.
func phpxBindArray(_ *model.FuncDecl, _ string, arg any) any { return arg }

// phpxInstallArgs adds nothing untagged.
func (rt *Runtime) phpxInstallArgs(_ map[string]any, _ *scopeRef) {}
