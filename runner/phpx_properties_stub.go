//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxPropVerdict answers the untagged build, where this tree is phpscript:
// every property is reachable from anywhere.
func (rt *Runtime) phpxPropVerdict(_ *model.Object, _ string, _ *Scope) (bool, error) {
	return false, nil
}

// phpxBindProps records nothing untagged.
func phpxBindProps(_ *model.Class, _ []model.Field, _ *model.Class) {}

// phpxPropAccess allows every access untagged.
func (rt *Runtime) phpxPropAccess(_ *model.Object, _ string, _ *Scope) error { return nil }

// phpxEachProp visits every property untagged.
func (rt *Runtime) phpxEachProp(obj *model.Object, _ *Scope, yield func(name string, v any) bool) {
	obj.Range(yield)
}

// phpxUninitialized is false untagged: every declared property starts set.
func phpxUninitialized(_ model.Field) bool { return false }

// phpxPropUnset allows every unset untagged.
func (rt *Runtime) phpxPropUnset(_ *model.Object, _ string, _ *Scope) error { return nil }

// phpxIndirectWrite allows every indexed write untagged.
func (rt *Runtime) phpxIndirectWrite(_ model.Expr, _ *Scope) error { return nil }
