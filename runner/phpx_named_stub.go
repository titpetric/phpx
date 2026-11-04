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
func phpxEntryArgs(_ any, args []any) []any { return args }

// phpxArgsValue answers the frame's own slice untagged.
func phpxArgsValue(args []any) any { return args }

// phpxOwnedArray keeps a Go slice as upstream does.
func phpxOwnedArray(v any) any { return v }
