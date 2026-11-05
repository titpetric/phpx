//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxMagicCall calls no __call untagged.
func (rt *Runtime) phpxMagicCall(_ *model.Object, _ string, _ []any, _ *Scope) (any, bool, error) {
	return nil, false, nil
}

// phpxMagicCallStatic answers the untagged build, which has no __callStatic.
func (rt *Runtime) phpxMagicCallStatic(_ *model.Class, _, _ string, _ []any, _ *Scope) (any, bool, error) {
	return nil, false, nil
}
