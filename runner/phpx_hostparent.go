//go:build phpx

package runner

import (
	"sync"

	"github.com/titpetric/phpx/model"
)

// A script class may extend a class the host builds in Go (DateTimeImmutable).
// `parent::__construct(...)` builds the Go value and keeps it beside the object,
// and a method the script class does not declare is called on that value.
var phpxHosts sync.Map // *model.Object -> any

// phpxHostConstruct runs a host class's constructor for the object whose
// `parent::__construct` this is. It reports false when name is no host class or
// no object is under construction.
func (rt *Runtime) phpxHostConstruct(name, method string, args []any, scope *Scope) (bool, error) {
	if method != "__construct" {
		return false, nil
	}
	current, ok := scope.Get("this")
	obj, isObj := current.(*model.Object)
	if !ok || !isObj {
		return false, nil
	}
	ctor, found := rt.lookupConstructor(name)
	if !found {
		return false, nil
	}
	host, err := rt.invokeEntry(ctor, args, scope)
	if err != nil {
		return true, err
	}
	phpxHosts.Store(obj, host)
	return true, nil
}

// phpxHostValue is the Go value behind a script object that extends a host class.
func phpxHostValue(obj *model.Object) (any, bool) { return phpxHosts.Load(obj) }
