//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/modelutil"
	"github.com/titpetric/phpx/model"
)

// phpxCheckInterfaceContracts answers the untagged build with upstream's
// check: a class has the methods it declares in its own body, and neither
// `extends` nor a trait confers one. The build in phpx_interfaces.go replaces
// it under `-tags phpx`.
func phpxCheckInterfaceContracts(prog *model.Program, _ map[string]*model.Class) error {
	return modelutil.CheckInterfaceContracts(prog)
}

// phpxMergeInterfaces answers the untagged build, where an interface holds its
// own constants and confers none of them on the interface that extends it.
func (rt *Runtime) phpxMergeInterfaces(_ []model.Stmt) {}

// phpxInterfaceConstTree answers the interface's own constants untagged.
func (rt *Runtime) phpxInterfaceConstTree(id *model.InterfaceDecl, _ map[string]bool) []model.Field {
	return id.Consts
}
