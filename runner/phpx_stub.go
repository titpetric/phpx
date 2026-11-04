//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxGlobal answers the untagged build, where this tree is phpscript: the
// statement parses and binds nothing (docs/design.md). The overlay in phpx.go
// replaces it under `-tags phpx`, and a fork commit stays reviewable against
// upstream because the untagged build still prints what upstream prints.
func (rt *Runtime) phpxGlobal(_ *model.Global, _ *Scope) error {
	return nil
}

// phpxDeclareFunc answers the untagged build: a declaration below the top
// level never runs, as upstream has it.
func (rt *Runtime) phpxDeclareFunc(_ *model.FuncDecl) error {
	return nil
}
