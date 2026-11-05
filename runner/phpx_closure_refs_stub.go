//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxCaptureRef records nothing untagged.
func phpxCaptureRef(refs map[string]*model.Cell, _ model.ClosureUse, _ *Scope) map[string]*model.Cell {
	return refs
}

// phpxSyncRefs writes nothing back untagged.
func (rt *Runtime) phpxSyncRefs(_ closureEnv, _ *Scope) func() { return func() {} }
