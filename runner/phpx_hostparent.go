//go:build phpx

package runner

import (
	"reflect"
	"strings"
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

// phpxHostInherit builds the Go value behind an object whose class extends a
// host class and declares no constructor of its own: php runs the parent's
// constructor with the arguments `new` was given.
func (rt *Runtime) phpxHostInherit(obj *model.Object, c *model.Class, args []any, scope *Scope) error {
	for _, name := range c.Ancestors {
		ctor, found := rt.lookupConstructor(name)
		if !found {
			continue
		}
		host, err := rt.invokeEntry(ctor, args, scope)
		if err != nil {
			return err
		}
		phpxHosts.Store(obj, host)
		return nil
	}
	return nil
}

// phpxWrapHostStatic gives a value a host static answered the class that was
// called. `static::createFromFormat()` from a subclass of DateTime answers an
// instance of that subclass, as php has it, and the subclass's object carries
// the host value beside it.
func (rt *Runtime) phpxWrapHostStatic(written, resolved string, out any, scope *Scope) any {
	if out == nil {
		return out
	}
	if _, isObject := out.(*model.Object); isObject {
		return out
	}
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return out
	}
	called := resolved
	switch strings.ToLower(written) {
	case "parent", "self", "static":
		if sc, err := resolveCalledClass("static", scope); err == nil {
			called = sc
		}
	}
	c, ok := rt.lookupClass(called)
	if !ok || strings.EqualFold(c.Name, resolved) || !phpxExtendsHost(c, resolved) {
		return out
	}
	obj := model.NewObject(c)
	defaults := rt.classDefaultScope(c, scope)
	for _, f := range c.Fields {
		var def any
		if f.Default != nil {
			if v, err := rt.Eval(f.Default, defaults); err == nil {
				def = v
			}
		}
		obj.SetProp(f.Name, def)
	}
	phpxHosts.Store(obj, out)
	return obj
}

// phpxExtendsHost reports whether the script class has the host class among
// its ancestors.
func phpxExtendsHost(c *model.Class, host string) bool {
	for _, a := range c.Ancestors {
		if strings.EqualFold(a, host) {
			return true
		}
	}
	return false
}

// phpxHostStaticOf finds a host static among the ancestors of a script class:
// `Carbon::createFromFormat()` runs DateTime's and answers a Carbon.
func (rt *Runtime) phpxHostStaticOf(c *model.Class, name, method string, args []any, scope *Scope) (any, bool, error) {
	for _, ancestor := range c.Ancestors {
		if entry, ok := rt.funcs[ancestor+"::"+method]; ok {
			out, err := rt.invokeEntry(entry, args, scope)
			return rt.phpxWrapHostStatic(name, ancestor, out, scope), true, err
		}
	}
	return nil, false, nil
}

// phpxHostCall runs a method on the Go value behind a script object. A method
// that answers the host value itself (a DateTime setter) answers the object,
// and one that answers another value of the host's type (a DateTimeImmutable
// modifier) answers a copy of the object around it, as php's `new static` has it.
func (rt *Runtime) phpxHostCall(obj *model.Object, host any, method string, args []any, scope *Scope) (any, error) {
	out, err := rt.callGoMethod(host, method, args, func() *Scope { return scope })
	if err != nil || out == nil {
		return out, err
	}
	if out == host {
		return obj, nil
	}
	hv, ov := reflect.ValueOf(host), reflect.ValueOf(out)
	if hv.Kind() == reflect.Pointer && ov.Kind() == reflect.Pointer && hv.Type() == ov.Type() && !ov.IsNil() {
		dup := model.NewObject(obj.Class)
		obj.Range(func(name string, val any) bool {
			dup.SetProp(name, model.CopyValue(val))
			return true
		})
		phpxHosts.Store(dup, out)
		return dup, nil
	}
	return out, nil
}

// phpxHostParentCall runs `parent::method()` of a host class on the host value
// of $this, the way php runs the internal parent method on the object.
func (rt *Runtime) phpxHostParentCall(class, method string, args []any, scope *Scope) (any, bool, error) {
	if _, isHost := rt.lookupConstructor(class); !isHost {
		return nil, false, nil
	}
	current, ok := scope.Get("this")
	obj, isObj := current.(*model.Object)
	if !ok || !isObj {
		return nil, false, nil
	}
	host, has := phpxHostValue(obj)
	if !has {
		return nil, false, nil
	}
	out, err := rt.phpxHostCall(obj, host, method, args, scope)
	return out, true, err
}

// phpxHostHasMethod reports whether the Go value has the method, without regard
// to case; a script class that extends a host class answers any other name
// through its own __call.
func phpxHostHasMethod(host any, name string) bool {
	rv := reflect.ValueOf(host)
	return rv.MethodByName(name).IsValid() || methodByNameFold(rv, name).IsValid()
}
