//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxGuard is one (object, name) a magic accessor is running for. php lets a
// __get that reads its own property see the plain property, not itself again.
type phpxGuard struct {
	obj  *model.Object
	name string
	kind string
}

// phpxMagicGet answers `$obj->name` through __get when the property is not
// there to read, or is not reachable from where the script stands. In the
// quiet mode isset() and empty() set, __isset decides first and a class
// without one answers nothing.
func (rt *Runtime) phpxMagicGet(obj *model.Object, name string, hidden bool, scope *Scope) (any, bool, error) {
	if obj.Class == nil {
		return nil, false, nil
	}
	if _, present := obj.Props[name]; present && !hidden {
		return nil, false, nil
	}
	get, hasGet := lookupPHPMethod(obj.Class, "__get")
	if rt.phpxGuarded(obj, name, "get") {
		return nil, false, nil
	}
	if rt.phpxQuiet > 0 {
		final := rt.phpxQuietFinal
		isset, has := lookupPHPMethod(obj.Class, "__isset")
		if !has {
			return nil, hasGet, nil
		}
		defer rt.phpxGuard(obj, name, "get")()
		set, err := rt.invokeMethod(obj, isset, []any{name}, scope)
		if err != nil || !phpTruthy(set) {
			return nil, true, err
		}
		if !hasGet || final {
			rt.phpxQuietFinal = false
			return true, true, nil
		}
	}
	if !hasGet {
		return nil, false, nil
	}
	defer rt.phpxGuard(obj, name, "get")()
	v, err := rt.invokeMethod(obj, get, []any{name}, scope)
	return v, true, err
}

// phpxMagicSet routes a write of a property that is not there through __set.
func (rt *Runtime) phpxMagicSet(obj *model.Object, name string, val any, scope *Scope) (bool, error) {
	if obj.Class == nil {
		return false, nil
	}
	if _, present := obj.Props[name]; present {
		return false, nil
	}
	set, ok := lookupPHPMethod(obj.Class, "__set")
	if !ok || rt.phpxGuarded(obj, name, "set") {
		return false, nil
	}
	defer rt.phpxGuard(obj, name, "set")()
	_, err := rt.invokeMethod(obj, set, []any{name, val}, scope)
	return true, err
}

// phpxMagicUnset routes unset($obj->name) of a missing property through __unset.
func (rt *Runtime) phpxMagicUnset(obj *model.Object, name string, scope *Scope) (bool, error) {
	if obj.Class == nil {
		return false, nil
	}
	if _, present := obj.Props[name]; present {
		return false, nil
	}
	unset, ok := lookupPHPMethod(obj.Class, "__unset")
	if !ok || rt.phpxGuarded(obj, name, "unset") {
		return false, nil
	}
	defer rt.phpxGuard(obj, name, "unset")()
	_, err := rt.invokeMethod(obj, unset, []any{name}, scope)
	return true, err
}

func (rt *Runtime) phpxGuarded(obj *model.Object, name, kind string) bool {
	for _, g := range rt.phpxGuards {
		if g.obj == obj && g.name == name && g.kind == kind {
			return true
		}
	}
	return false
}

// phpxGuard marks (obj, name, kind) as running and answers the func that clears it.
func (rt *Runtime) phpxGuard(obj *model.Object, name, kind string) func() {
	rt.phpxGuards = append(rt.phpxGuards, phpxGuard{obj, name, kind})
	n := len(rt.phpxGuards)
	return func() { rt.phpxGuards = rt.phpxGuards[:n-1] }
}

// phpxQuietEnter and phpxQuietLeave bracket the evaluation of an isset() or
// empty() argument.
func (rt *Runtime) phpxQuietEnter(final bool) {
	rt.phpxQuiet++
	rt.phpxQuietFinal = final
}
func (rt *Runtime) phpxQuietLeave() {
	if rt.phpxQuiet > 0 {
		rt.phpxQuiet--
	}
	rt.phpxQuietFinal = false
}

// phpxInQuiet reports whether the expression being evaluated is the argument of
// isset() or empty().
func (rt *Runtime) phpxInQuiet() bool { return rt.phpxQuiet > 0 }
