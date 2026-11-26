package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// This file implements PHP's static class members: `Class::method()`,
// `Class::$property` and the `self::` / `static::` spellings of both.
//
// A static member belongs to the class, not to an instance, and has no
// receiver to hang off the object model in value.go. A method runs in a
// scope holding its class alone; the Runtime keeps one bag of properties per
// class (see Runtime.classStatics), and composer's `self::$registeredLoaders`
// is observable from any instance and from a later static call.

// resolveClassName maps the contextual class names PHP allows in a static
// reference onto a real one. `self` and `static` both name the class of the
// running method, because there is no inheritance and late static binding
// resolves to the same class. `parent` names the parent of the class whose
// body runs, which the frame records when the fork inherits a member
// (runner/phpx_inheritance.go); an untagged build records none and refuses it.
func resolveClassName(class string, scope *Scope) (string, error) {
	switch class {
	case "self", "static":
		if scope != nil {
			if current, ok := scope.Get("__class__"); ok {
				if name, ok := current.(string); ok && name != "" {
					return name, nil
				}
			}
		}
	case "parent":
		if scope != nil {
			if current, ok := scope.Get("__parent__"); ok {
				if name, ok := current.(string); ok && name != "" {
					return name, nil
				}
			}
		}
		return "", phpxNoParentError()
	}
	return strings.TrimPrefix(class, "\\"), nil
}

// classDefaultScope returns the scope where a class's declared defaults resolve, binding `__class__`.
func (rt *Runtime) classDefaultScope(class *model.Class, caller *Scope) *Scope {
	scope := rt.newScope()
	if caller != nil {
		if file, ok := caller.Get("__FILE__"); ok {
			scope.Set("__FILE__", file)
		}
		if dir, ok := caller.Get("__DIR__"); ok {
			scope.Set("__DIR__", dir)
		}
	}
	scope.Set("__class__", class.Name)
	return scope
}

// staticBag returns a class's live static storage, seeded from its defaults.
func (rt *Runtime) staticBag(class *model.Class, scope *Scope) (map[string]any, error) {
	if bag, ok := rt.classStatics[class.Name]; ok {
		return bag, nil
	}
	bag := make(map[string]any, len(class.Statics))
	defaults := rt.classDefaultScope(class, scope)
	for _, field := range class.Statics {
		var value any
		if field.Default != nil {
			evaluated, err := rt.Eval(field.Default, defaults)
			if err != nil {
				return nil, fmt.Errorf("static %s::$%s: %w", class.Name, field.Name, err)
			}
			value = evaluated
		}
		bag[field.Name] = value
	}
	rt.classStatics[class.Name] = bag
	return bag, nil
}

// staticStorage returns the bag behind a static reference, autoloading the class when undeclared.
func (rt *Runtime) staticStorage(class string, scope *Scope) (map[string]any, error) {
	name, err := resolveClassName(class, scope)
	if err != nil {
		return nil, err
	}
	if !rt.hasClass(name) {
		if err := rt.autoload(name, scope); err != nil {
			return nil, err
		}
	}
	decl, ok := rt.lookupClass(name)
	if !ok {
		return nil, fmt.Errorf("static property %s::$...: unknown class", name)
	}
	return rt.staticBag(decl, scope)
}

// staticsKey is the scope slot holding a closure frame's function-static table.
const staticsKey = "__statics__"

// staticVarBag returns a `static $x` statement's storage and earlier values.
func (rt *Runtime) staticVarBag(n *model.StaticVar, scope *Scope) (map[string]any, bool) {
	if v, ok := scope.Get(staticsKey); ok {
		if table, ok := v.(map[*model.StaticVar]map[string]any); ok {
			bag, seeded := table[n]
			if !seeded {
				bag = map[string]any{}
				table[n] = bag
			}
			return bag, seeded
		}
	}
	bag, seeded := rt.funcStatics[n]
	if !seeded {
		bag = map[string]any{}
		rt.funcStatics[n] = bag
	}
	return bag, seeded
}

// execStaticVar binds a `static $x` name, running its initializer on first execution only.
func (rt *Runtime) execStaticVar(n *model.StaticVar, scope *Scope) error {
	bag, seeded := rt.staticVarBag(n, scope)
	for _, decl := range n.Vars {
		if !seeded {
			var value any
			if decl.Default != nil {
				evaluated, err := rt.Eval(decl.Default, scope)
				if err != nil {
					return fmt.Errorf("static $%s: %w", decl.Name, err)
				}
				value = evaluated
			}
			bag[decl.Name] = value
		}
		scope.bindStatic(decl.Name, bag)
	}
	return nil
}

// helperStaticProp reads `Class::$name`, an undeclared property reading as null.
func (rt *Runtime) helperStaticProp(ref *scopeRef) func(class, name string) (any, error) {
	return func(class, name string) (any, error) {
		bag, err := rt.staticStorageFor(class, name, ref.scope)
		if err != nil {
			return nil, err
		}
		return bag[name], nil
	}
}

// setStaticProp writes `Class::$name`.
func (rt *Runtime) setStaticProp(class, name string, value any, scope *Scope) error {
	bag, err := rt.staticStorageFor(class, name, scope)
	if err != nil {
		return err
	}
	bag[name] = value
	return nil
}

// helperStaticCall dispatches `Class::method` to a host static or a declared PHP method.
func (rt *Runtime) helperStaticCall(ref *scopeRef) func(class string, methodValue any, args ...any) (any, error) {
	return func(class string, methodValue any, args ...any) (any, error) {
		scope := ref.scope
		// `Class::$m(...)` arrives with the method name as a runtime value.
		method := phpString(methodValue)
		name, err := resolveCalledClass(class, scope)
		if err != nil {
			return nil, err
		}
		// This reads the host table by exact name alone, one map lookup on
		// the common path, where lookupFunc's case-insensitive fallback
		// scans the whole function table on every miss, the cost of every PHP
		// static call.
		if entry, ok := rt.funcs[name+"::"+method]; ok {
			out, err := rt.invokeEntry(entry, args, scope)
			return rt.phpxWrapHostStatic(class, name, out, scope), err
		}
		if !rt.hasClass(name) {
			if err := rt.autoload(name, scope); err != nil {
				return nil, err
			}
		}
		if decl, ok := rt.lookupClass(name); ok {
			fn, ok := lookupPHPMethod(decl, method)
			if !ok {
				if out, handled, err := rt.phpxHostStaticOf(decl, name, method, args, scope); handled {
					return out, err
				}
				if out, handled, err := rt.phpxMagicCallStatic(decl, name, method, args, scope); handled {
					return out, err
				}
				return nil, fmt.Errorf("call to undefined method %s::%s()", name, method)
			}
			if err := rt.phpxCheckVisible(scope, decl, fn); err != nil {
				// php sends a call to an inaccessible method to __callStatic.
				if out, handled, merr := rt.phpxMagicCallStatic(decl, name, method, args, scope); handled {
					return out, merr
				}
				return nil, err
			}
			if current, ok := scope.Get("this"); !fn.Static && ok {
				// The fork forwards `$this` to an inherited method, whose
				// declaring class is an ancestor of the instance's own.
				if obj, ok := current.(*model.Object); ok && phpxCarries(obj.Class, decl) {
					return rt.invokeMethod(obj, fn, args, scope)
				}
			}
			return rt.invokeStatic(decl, phpxForwardedClass(class, decl.Name, scope), fn, args, scope)
		}
		// No PHP class of that name: the target can only be a host static and
		// pay for the case-insensitive lookup here and not on every call.
		if entry, ok := rt.lookupEntry(name + "::" + method); ok {
			return rt.invokeEntry(entry, args, scope)
		}
		if handled, herr := rt.phpxHostConstruct(name, method, args, scope); handled {
			return nil, herr
		}
		if out, handled, herr := rt.phpxHostParentCall(name, method, args, scope); handled {
			return out, herr
		}
		return nil, fmt.Errorf("static call %s::%s(): unknown class", name, method)
	}
}

// invokeStatic runs a method with no receiver. `$this` stays unset, as in a
// genuinely static method or a class call with no instance behind it, while
// the scope binds `__class__` so `self::` inside the body resolves back to the
// same class. helperStaticCall forwards instance calls before reaching here.
func (rt *Runtime) invokeStatic(class *model.Class, called string, decl *model.FuncDecl, args []any, caller *Scope) (any, error) {
	scope := rt.newScope()
	scope.fn = decl.Name
	rt.pushFrame(scope)
	defer rt.popFrame()
	if decl.Filename != "" {
		rt.setScopeFile(scope, decl.Filename)
	}
	scope.args = args
	owner := rt.phpxOwnerClass(class, decl)
	if err := rt.phpxCallAbstract(class, decl); err != nil {
		return nil, err
	}
	scope.Set("__class__", owner)
	phpxSetCalled(scope, called)
	if parent := rt.phpxParentScope(owner); parent != "" {
		scope.Set("__parent__", rt.boxed(parent))
	}
	if err := rt.bindParams(decl, args, scope); err != nil {
		return nil, err
	}
	value, _, runErr := rt.exec(decl.Body, scope)
	return value, combineErrors(runErr, rt.runDeferred(scope, 0))
}

// helperInvoke calls any PHP callable held in a value, as in `$fn($x)`.
func (rt *Runtime) helperInvoke(ref *scopeRef) func(callee any, args ...any) (any, error) {
	return func(callee any, args ...any) (any, error) {
		scope := ref.scope
		fn, ok := rt.callableWithScope(scope, callee)
		if !ok {
			return nil, phpNotCallable(callee)
		}
		return fn(rt.phpxInvokeArgs(callee, args)...)
	}
}

// helperCallable builds the first-class callable `name(...)` or `$fn(...)` names.
func (rt *Runtime) helperCallable(ref *scopeRef) func(target any, fallback string) (any, error) {
	return func(target any, fallback string) (any, error) {
		return rt.closureValue(target, fallback, ref.scope)
	}
}

// helperCallableMember builds a first-class callable over `$obj->method(...)` or `Class::method(...)`.
func (rt *Runtime) helperCallableMember(ref *scopeRef) func(target, method any) (any, error) {
	return func(target, method any) (any, error) {
		scope := ref.scope
		if class, ok := target.(string); ok {
			resolved, err := resolveClassName(class, scope)
			if err != nil {
				return nil, err
			}
			target = resolved
		}
		return rt.closureMember(target, phpString(method), scope)
	}
}
