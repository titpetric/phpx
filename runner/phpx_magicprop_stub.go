//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxGuard is absent in the untagged build; the Runtime field needs a type.
type phpxGuard struct{}

// phpxMagicGet answers the untagged build, which has no __get.
func (rt *Runtime) phpxMagicGet(_ *model.Object, _ string, _ bool, _ *Scope) (any, bool, error) {
	return nil, false, nil
}

// phpxMagicSet answers the untagged build, which has no __set.
func (rt *Runtime) phpxMagicSet(_ *model.Object, _ string, _ any, _ *Scope) (bool, error) {
	return false, nil
}

// phpxMagicUnset answers the untagged build, which has no __unset.
func (rt *Runtime) phpxMagicUnset(_ *model.Object, _ string, _ *Scope) (bool, error) {
	return false, nil
}

func (rt *Runtime) phpxQuietEnter(_ bool) {}
func (rt *Runtime) phpxQuietLeave()       {}

// phpxInQuiet is never true untagged.
func (rt *Runtime) phpxInQuiet() bool { return false }

// phpxHostField finds no host property untagged.
func phpxHostField(_ any, _ string) (any, bool) { return nil, false }

// phpxHostFieldSet sets no host property untagged.
func phpxHostFieldSet(_ any, _ string, _ any) bool { return false }
