//go:build !phpx

package runner

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxClass answers the untagged build, where this tree is phpscript: a class
// arrives with the members it declared and neither `extends` nor a trait
// confers anything, which is the rule docs/design.md holds. The build in
// phpx_inheritance.go replaces it under `-tags phpx`.
func (rt *Runtime) phpxClass(cd *model.ClassDecl, _ map[string]*model.ClassDecl, _ map[string]*model.TraitDecl, stmts []model.Stmt) (*model.Class, error) {
	c := &model.Class{Name: cd.Name, Implements: model.InterfaceNames(cd, stmts), Fields: cd.Fields, Statics: cd.Statics, Consts: cd.Consts, Methods: map[string]*model.FuncDecl{}}
	for _, m := range cd.Methods {
		c.Methods[m.Name] = m
	}
	return c, nil
}

// phpxOwnerClass is the class of the instance, which is the only class an
// untagged build has.
func (rt *Runtime) phpxOwnerClass(class *model.Class, _ *model.FuncDecl) string {
	if class == nil {
		return ""
	}
	return class.Name
}

// phpxParentScope records no parent: extends records a name and builds no
// chain to walk, and `parent::` reaches resolveClassName with no name to read.
func (rt *Runtime) phpxParentScope(_ string) string {
	return ""
}

// phpxNoParentError answers upstream's wording for the refusal, which an
// untagged build keeps so the fixture pinning it still passes.
func phpxNoParentError() error {
	return fmt.Errorf(`cannot use "parent": no inheritance at runtime`)
}

// phpxCarries forwards `$this` only to the class the lookup found.
func phpxCarries(class, target *model.Class) bool {
	return class == target
}

// phpxCatchExtends carries the upstream answer: a catch clause matches the
// class the throwable records and walks no chain.
func phpxCatchExtends(_ string, _ error) bool {
	return false
}
