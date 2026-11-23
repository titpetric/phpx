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

// phpxUninitializedRead allows every read untagged.
func (rt *Runtime) phpxUninitializedRead(_ *model.Object, _ string, _ *Scope) error { return nil }

// phpxSlotKey answers the property's own name untagged: one slot per name.
func (rt *Runtime) phpxSlotKey(_ *model.Object, name string, _ *Scope) string { return name }

// phpxSeedShadow writes nothing untagged: there are no hidden slots.
func (rt *Runtime) phpxSeedShadow(_ *model.Object, _ *model.Class, _ *Scope) error { return nil }

// phpxCoerceProperty stores the value as given untagged.
func (rt *Runtime) phpxCoerceProperty(_ *model.Object, _ string, v any) any { return v }
