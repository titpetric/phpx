//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxHookGet answers the untagged build, which has no property hooks.
func (rt *Runtime) phpxHookGet(_ *model.Object, _ string, _ *Scope) (any, bool, error) {
	return nil, false, nil
}

// phpxHookSet answers the untagged build, which has no property hooks.
func (rt *Runtime) phpxHookSet(_ *model.Object, _ string, _ any, _ *Scope) (bool, error) {
	return false, nil
}
