//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// PhpxRebind answers a copy of a closure bound to newThis and, when scope is
// not empty, to the class named scope, which is how composer's loader reads a
// private property from `Closure::bind(fn, null, ClassLoader::class)`. A value
// that is not a closure literal answers false.
func (rt *Runtime) PhpxRebind(closure any, newThis any, scope string) (any, bool) {
	c, ok := closure.(*Callable)
	if !ok || c.closure == nil {
		return nil, false
	}
	env := c.env
	env.this = newThis
	if scope != "" {
		env.class = scope
		env.called = scope
	}
	return rt.newClosure(c.closure, env), true
}

// phpxClosureMethod answers `$closure->bindTo($this, $scope)` and
// `$closure->call($newThis, ...$args)` on a closure value. handled is false
// for any other method or receiver, which go on to the reflection path.
func (rt *Runtime) phpxClosureMethod(base any, method string, args []any) (result any, handled bool, err error) {
	c, ok := base.(*Callable)
	if !ok || c.closure == nil {
		return nil, false, nil
	}
	switch strings.ToLower(method) {
	case "bindto":
		var newThis any
		scope := ""
		if len(args) > 0 {
			newThis = args[0]
		}
		if len(args) > 1 {
			scope = phpxScopeName(args[1])
		}
		out, _ := rt.PhpxRebind(c, newThis, scope)
		return out, true, nil
	case "call":
		if len(args) == 0 {
			return nil, true, &ArgumentCountError{Want: 1, Got: 0}
		}
		obj, isObj := args[0].(*model.Object)
		if !isObj || obj.Class == nil {
			return nil, true, fmt.Errorf("Closure::call(): Argument #1 ($newThis) must be of type object")
		}
		bound, _ := rt.PhpxRebind(c, obj, obj.Class.Name)
		out, err := bound.(*Callable).Call(args[1:]...)
		return out, true, err
	}
	return nil, false, nil
}

// phpxScopeName reads the scope argument of bind and bindTo: a class name, an
// object whose class is the scope, or "static" for the closure's own.
func phpxScopeName(v any) string {
	switch s := v.(type) {
	case string:
		if s != "static" {
			return strings.TrimPrefix(s, "\\")
		}
	case *model.Object:
		if s.Class != nil {
			return s.Class.Name
		}
	}
	return ""
}
