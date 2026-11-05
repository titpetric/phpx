//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxInstallSPL adds nothing untagged.
func (rt *Runtime) phpxInstallSPL(_ map[string]any) {}

// phpxArrayAccessSet writes through no ArrayAccess untagged.
func (rt *Runtime) phpxArrayAccessSet(_ *model.Index, _ any, _ string, _ any, _ *Scope) (bool, error) {
	return false, nil
}

// phpxArrayAccessUnset calls no offsetUnset untagged.
func (rt *Runtime) phpxArrayAccessUnset(_, _ any) bool { return false }

// phpxForeachObject iterates no Traversable untagged.
func (rt *Runtime) phpxForeachObject(_ *model.Object, _ func(k, v any) bool) (bool, error) {
	return false, nil
}

func (rt *Runtime) phpxListPart(_ *model.Object, _ int64) any { return nil }
