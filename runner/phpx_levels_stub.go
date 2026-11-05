//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxOuterLevels answers the untagged build, where break and continue leave
// the innermost loop alone.
func (rt *Runtime) phpxOuterLevels(_ int) {}

// phpxLeave answers the untagged build: no level is ever left to pass on.
func (rt *Runtime) phpxLeave(_ flow) bool { return false }

// phpxLabelIndex answers the untagged build, which has no goto.
func (rt *Runtime) phpxLabelIndex(_ []model.Stmt) int { return -1 }
