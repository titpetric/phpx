//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxCheckVisible answers the untagged build, where this tree is phpscript:
// a call runs whatever visibility the method declared.
func (rt *Runtime) phpxCheckVisible(_ *Scope, _ *model.Class, _ *model.FuncDecl) error {
	return nil
}
