//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// Hook methods are named by the parser: `__phpx_get_NAME`, `__phpx_set_NAME`.
const (
	phpxGetHook = "__phpx_get_"
	phpxSetHook = "__phpx_set_"
)

// phpxHookGet answers `$obj->name` through the property's get hook. Inside the
// hook the property reads its backing value, as php does.
func (rt *Runtime) phpxHookGet(obj *model.Object, name string, scope *Scope) (any, bool, error) {
	if obj.Class == nil {
		return nil, false, nil
	}
	get, ok := lookupPHPMethod(obj.Class, phpxGetHook+name)
	if !ok || rt.phpxGuarded(obj, name, "hookget") {
		return nil, false, nil
	}
	defer rt.phpxGuard(obj, name, "hookget")()
	v, err := rt.invokeMethod(obj, get, nil, scope)
	return v, true, err
}

// phpxHookSet routes `$obj->name = value` through the property's set hook. The
// hook's own write to the property stores the value.
func (rt *Runtime) phpxHookSet(obj *model.Object, name string, val any, scope *Scope) (bool, error) {
	if obj.Class == nil {
		return false, nil
	}
	set, ok := lookupPHPMethod(obj.Class, phpxSetHook+name)
	if !ok || rt.phpxGuarded(obj, name, "hookset") {
		return false, nil
	}
	defer rt.phpxGuard(obj, name, "hookset")()
	_, err := rt.invokeMethod(obj, set, []any{val}, scope)
	return true, err
}
