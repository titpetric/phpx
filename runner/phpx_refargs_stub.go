//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxBindRefArgs answers the untagged build, where no argument is a reference.
func phpxBindRefArgs(_ *model.FuncDecl, args []any, _ *Scope) []any { return args }

// phpxInvokeArgs answers the untagged build, where no argument is wrapped.
func (rt *Runtime) phpxInvokeArgs(_ any, args []any) []any { return args }
