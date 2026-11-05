//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxPendingClass is absent in the untagged build; the field needs a type.
type phpxPendingClass struct{}

// phpxDeferClass answers the untagged build, which defers nothing.
func (rt *Runtime) phpxDeferClass(_ *model.ClassDecl, _ error, _ map[string]*model.ClassDecl, _ map[string]*model.TraitDecl, _ []model.Stmt, _ string) bool {
	return false
}

// phpxDeclareDeferred answers the untagged build, which has none to declare.
func (rt *Runtime) phpxDeclareDeferred(_ *model.ClassDecl) error { return nil }

// phpxDeclareNested answers the untagged build, which declares only what hoist found.
func (rt *Runtime) phpxDeclareNested(_ *model.ClassDecl) error { return nil }
