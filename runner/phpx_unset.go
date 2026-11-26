package runner

import (
	"fmt"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// This file holds unset(), which the fork owns: a target whose value is the
// last reference to an instance with a `__destruct` notes that instance before
// the storage goes (phpx_destruct.go). Unsetting a name nothing set is not an
// error, and that is what lets `unset($map[$key])` run unconditionally. The
// function sits here rather than in runner_assign.go because that file runs to
// the 400 lines a file doing one thing may carry.

// execUnset removes each target from the scope, array, property bag or static bag.
func (rt *Runtime) execUnset(n *model.Unset, scope *Scope) error {
	for _, target := range n.Targets {
		switch tgt := modelutil.UnwrapParenthesized(target).(type) {
		case *model.Var:
			if old, ok := scope.Get(tgt.Name); ok {
				rt.phpxNoteDropped(old)
			}
			scope.Unset(tgt.Name)
		case *model.PropAccess:
			base, err := rt.Eval(tgt.Base, scope)
			if err != nil {
				return err
			}
			name, err := rt.propName(tgt, scope)
			if err != nil {
				return err
			}
			if obj, ok := base.(*model.Object); ok {
				if handled, err := rt.phpxMagicUnset(obj, name, scope); handled {
					if err != nil {
						return err
					}
					continue
				}
				if err := rt.phpxPropUnset(obj, name, scope); err != nil {
					return err
				}
				if old, ok := obj.Prop(rt.phpxSlotKey(obj, name, scope)); ok {
					rt.phpxNoteDropped(old)
				}
				obj.DeleteProp(rt.phpxSlotKey(obj, name, scope))
			}
		case *model.StaticProp:
			name, err := rt.phpxStaticName(tgt, scope)
			if err != nil {
				return err
			}
			bag, err := rt.staticStorageFor(tgt.Class, name, scope)
			if err != nil {
				return err
			}
			delete(bag, name)
		case *model.Index:
			if tgt.Index == nil {
				return fmt.Errorf("unset: [] is not an unset target")
			}
			// The unset writes into the base, so a base array that another
			// handle reads is split along the path first.
			base, err := rt.phpxSplitBase(tgt.Base, scope)
			if err != nil {
				return err
			}
			key, err := rt.Eval(tgt.Index, scope)
			if err != nil {
				return err
			}
			if arr, ok := base.(*model.Array); ok {
				if old, ok := arr.Get(normalizeKey(key)); ok {
					rt.phpxNoteDropped(old)
				}
				arr.Delete(normalizeKey(key))
				continue
			}
			if keyed, ok := base.(model.Keyed); ok {
				keyed.Delete(keyString(key))
				continue
			}
			if rt.phpxArrayAccessUnset(base, key) {
				continue
			}
			// A reindexed slice is a new value and it goes back to where the
			// base reads from. A map deleted in place answers nothing.
			if replacement, changed := unsetGoIndex(base, key); changed {
				if err := rt.assignTo(tgt.Base, replacement, scope); err != nil {
					return err
				}
			}
		case *model.Call:
			// `unset($$name)` is the variable-variable read the parser lowers
			// to a call; the variable it names is the one to unset.
			handled, err := rt.phpxVarVarUnset(tgt, scope)
			if err != nil {
				return err
			}
			if !handled {
				return fmt.Errorf("unset: unsupported target %T", target)
			}
		default:
			return fmt.Errorf("unset: unsupported target %T", target)
		}
	}
	return nil
}
