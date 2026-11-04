//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxStringOffset writes no string offset untagged.
func (rt *Runtime) phpxStringOffset(_ *model.Index, _ any, _ string, _ any, _ *Scope) (bool, error) {
	return false, nil
}
