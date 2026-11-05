//go:build phpx

package runner

import (
	"reflect"

	"github.com/titpetric/phpx/model"
)

// phpxInstallRefExpr adds __refexpr(target), the setter of an output parameter
// that is a property: curl_multi_exec($mh, $this->active).
func (rt *Runtime) phpxInstallRefExpr(env map[string]any, ref *scopeRef) {
	env["__refexpr"] = adapt(func(target model.Expr) func(any) {
		scope := ref.scope
		return func(v any) { _ = rt.assignTo(target, v, scope) }
	})
}

// phpxHandleID answers `(int) $handle` for a Go value a binding made, a curl
// handle among them: an identity that holds while the value lives. php answers
// the object id.
func phpxHandleID(v any) (int64, bool) {
	if PhpxIsResource(v) {
		return PhpxResourceID(v), true
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return 0, false
	}
	if _, isObject := v.(*model.Object); isObject {
		return 0, false
	}
	return int64(rv.Pointer()), true
}
