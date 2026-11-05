//go:build phpx

package runner

import (
	"reflect"
	"sync"

	"github.com/titpetric/phpx/model"
)

// phpxSnapshotSource reports whether a by-value foreach has to walk a copy of
// its array. php iterates the array as it was when the loop began, so a body
// that unsets from it, or calls something that appends to it, leaves the walk
// alone. A source that is an expression (a call, a literal) is the loop's own,
// and a body with no call, no unset and no include cannot reach the array.
func phpxSnapshotSource(n *model.Foreach) bool {
	if n.ByRef {
		return false
	}
	switch model.UnwrapParenthesized(n.Source).(type) {
	case *model.Var, *model.PropAccess, *model.StaticProp, *model.Index:
	default:
		return false
	}
	if v, ok := phpxSnapshotCache.Load(n); ok {
		return v.(bool)
	}
	reach := phpxMayReach(n.Body)
	phpxSnapshotCache.Store(n, reach)
	return reach
}

var phpxSnapshotCache sync.Map // *model.Foreach -> bool

// phpxMayReach walks a statement list for a node that can write to storage the
// list does not name: a call of any kind, an unset, an include.
func phpxMayReach(body []model.Stmt) bool {
	found := false
	var walk func(v reflect.Value)
	walk = func(v reflect.Value) {
		if found || !v.IsValid() {
			return
		}
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if v.IsNil() {
				return
			}
			if v.Kind() == reflect.Pointer {
				switch v.Interface().(type) {
				case *model.Call, *model.MethodCall, *model.StaticCall, *model.New,
					*model.Invoke, *model.Unset, *model.Include:
					found = true
					return
				}
			}
			walk(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					walk(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				walk(v.Index(i))
			}
		}
	}
	for _, s := range body {
		walk(reflect.ValueOf(s))
	}
	return found
}

// phpxSnapshot returns a shallow copy of the array the loop walks.
func phpxSnapshot(src *model.Array) *model.Array {
	out := model.NewArraySize(src.Len())
	src.Range(func(k, v any) bool {
		out.Set(k, v)
		return true
	})
	return out
}
