//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxCallAbstract answers the untagged build, where this tree is phpscript: a
// method with no body runs as an empty body and returns null, as upstream
// does. The build in phpx_abstract.go refuses it under `-tags phpx`.
func (rt *Runtime) phpxCallAbstract(_ *model.Class, _ *model.FuncDecl) error {
	return nil
}

// phpxInstantiable answers the untagged build: a class instantiates whatever
// its declaration said, because `abstract` confers nothing there.
func phpxInstantiable(_ *model.Class) error {
	return nil
}

// phpxDeclOwner answers the untagged build, which has no declaring-class table
// to read: the instance's class is the only one it knows.
func (rt *Runtime) phpxDeclOwner(class *model.Class, decl *model.FuncDecl) string {
	return rt.phpxOwnerClass(class, decl)
}
