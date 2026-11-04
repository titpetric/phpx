package runner

import (
	"errors"

	"github.com/titpetric/phpx/model"
)

// This file runs a PHP callee in a frame of its own: a named function, a
// method with $this bound, a closure over its captured environment, with the
// parameter binding they share and the deferred and shutdown callbacks that
// unwind a frame and a program.

// invokeFunc runs a user-defined function in a fresh scope.
func (rt *Runtime) invokeFunc(decl *model.FuncDecl, args []any) (any, error) {
	scope := rt.acquireScope()
	defer rt.releaseScope(scope)
	rt.pushFrame(scope)
	defer rt.popFrame()
	if decl.Filename != "" {
		rt.setScopeFile(scope, decl.Filename)
	}
	scope.args = args
	if err := rt.bindParams(decl, args, scope); err != nil {
		return nil, err
	}
	val, _, runErr := rt.exec(decl.Body, scope)
	return val, combineErrors(runErr, rt.runDeferred(scope, 0))
}

// invokeMethod runs a method with $this bound to obj. The invocation span uses
// the caller scope; the fresh scope below records where the method body lives,
// and spans starting inside that body read it from there.
func (rt *Runtime) invokeMethod(obj *model.Object, decl *model.FuncDecl, args []any, caller *Scope) (any, error) {
	scope := rt.acquireScope()
	defer rt.releaseScope(scope)
	rt.pushFrame(scope)
	defer rt.popFrame()
	if decl.Filename != "" {
		rt.setScopeFile(scope, decl.Filename)
	}
	scope.Set("this", obj)
	scope.args = args
	// A method the class left to a child has no body to run. The fork refuses
	// the call where an untagged build runs an empty body.
	if err := rt.phpxCallAbstract(obj.Class, decl); err != nil {
		return nil, err
	}
	if obj.Class != nil {
		// The class of the running body, not of the instance: `self::` in a
		// method inherited from a parent resolves to the parent. Its parent is
		// what `parent::` reads, and the fork's frame records both.
		owner := rt.phpxOwnerClass(obj.Class, decl)
		scope.Set("__class__", rt.boxed(owner))
		phpxSetCalled(scope, obj.Class.Name)
		if parent := rt.phpxParentScope(owner); parent != "" {
			scope.Set("__parent__", rt.boxed(parent))
		}
	}
	// The span name costs two string builds and they happen only when an
	// observer is listening.
	if len(rt.observers) > 0 {
		name := obj.ID
		if name == "" && obj.Class != nil {
			name = obj.Class.Name
		}
		if name != "" {
			name += "."
		}
		traceScope := caller
		if traceScope == nil {
			traceScope = scope
		}
		defer rt.trace(traceScope, name+decl.Name)()
	}
	if err := rt.bindParams(decl, args, scope); err != nil {
		return nil, err
	}
	val, _, runErr := rt.exec(decl.Body, scope)
	return val, combineErrors(runErr, rt.runDeferred(scope, 0))
}

// closureEnv is everything a closure carries away from its defining scope:
// the lexical magic constants, the `use (...)` captures, and the bound
// `$this` (nil for a `static function`, or for one written outside a method).
// captureClosureEnv fills it once where newClosure builds the closure value,
// not on each call, which makes the capture a snapshot the way PHP's by-value
// `use` is.
type closureEnv struct {
	filename  any
	directory any
	captured  map[string]any
	this      any
	class     any
	// parent is the class `parent::` resolves to, carried so a closure
	// declared inside a method answers it as the method body does.
	parent any
	// called is the class `static::` resolves to, carried the same way.
	called string

	// statics holds this closure value's function-static bags: PHP keeps a
	// `static $x` storage per closure instance and two counters built by
	// the same factory count independently. Keyed like Runtime.funcStatics.
	statics map[*model.StaticVar]map[string]any
}

// captureClosureEnv snapshots scope for a closure declared in it.
func captureClosureEnv(cl *model.Closure, scope *Scope) closureEnv {
	env := closureEnv{statics: map[*model.StaticVar]map[string]any{}}
	env.filename, _ = scope.Get("__FILE__")
	env.directory, _ = scope.Get("__DIR__")
	if !cl.Static {
		env.this, _ = scope.Get("this")
		env.class, _ = scope.Get("__class__")
		env.parent, _ = scope.Get("__parent__")
		env.called, _ = phpxCalled(scope)
	}
	if len(cl.Uses) > 0 {
		env.captured = make(map[string]any, len(cl.Uses))
		for _, use := range cl.Uses {
			// A by-reference capture binds the same value a by-value one does:
			// the runtime has no reference cells and the distinction only shows
			// up in writes the closure makes back to the enclosing frame, which
			// it cannot do either way.
			value, _ := scope.Get(use.Name)
			env.captured[use.Name] = value
		}
	}
	return env
}

// invokeClosure runs an anonymous function in a fresh scope seeded with its
// captured environment alone: the enclosing frame's other variables
// are not reachable from the body. bindParams runs after the captures and a
// parameter of the same name wins; php rejects that overlap outright, as
// "Cannot use lexical variable $x as a parameter name".
func (rt *Runtime) invokeClosure(cl *model.Closure, args []any, env closureEnv) (any, error) {
	scope := rt.newScope()
	rt.pushFrame(scope)
	defer rt.popFrame()
	scope.args = args
	if env.filename != nil {
		scope.Set("__FILE__", env.filename)
	}
	if env.directory != nil {
		scope.Set("__DIR__", env.directory)
	}
	if env.this != nil {
		scope.Set("this", env.this)
	}
	if env.class != nil {
		scope.Set("__class__", env.class)
	}
	if env.parent != nil {
		scope.Set("__parent__", env.parent)
	}
	phpxSetCalled(scope, env.called)
	for name, value := range env.captured {
		scope.Set(name, value)
	}
	scope.Set(staticsKey, env.statics)
	decl := &model.FuncDecl{Params: cl.Params, Body: cl.Body}
	if err := rt.bindParams(decl, args, scope); err != nil {
		return nil, err
	}
	val, _, runErr := rt.exec(cl.Body, scope)
	return val, combineErrors(runErr, rt.runDeferred(scope, 0))
}

// runDeferred invokes callbacks registered since mark in last-in, first-out
// order. It removes each one before the call and a callback runs at most once,
// including when it fails or registers another during the unwind.
func (rt *Runtime) runDeferred(scope *Scope, mark int) error {
	var errs []error
	for len(scope.deferred) > mark {
		i := len(scope.deferred) - 1
		callback := scope.deferred[i]
		scope.deferred[i] = nil
		scope.deferred = scope.deferred[:i]
		if _, err := rt.invokeWithScopeContext(callback, nil, scope); err != nil {
			errs = append(errs, err)
		}
	}
	return combineErrors(errs...)
}

func (rt *Runtime) runShutdown() error {
	if len(rt.shutdown) == 0 {
		return nil
	}
	defer rt.suspendDeadline()()

	var errs []error
	scope := rt.newScope()
	rt.pushFrame(scope)
	defer rt.popFrame()
	for len(rt.shutdown) > 0 {
		callback := rt.shutdown[0]
		rt.shutdown[0] = nil
		rt.shutdown = rt.shutdown[1:]
		if _, err := rt.invokeWithScopeContext(callback, nil, scope); err != nil {
			errs = append(errs, err)
		}
	}
	return combineErrors(errs...)
}

// combineErrors preserves an existing error unchanged when there is only one,
// while retaining every error when execution and one or more defers fail.
func combineErrors(errs ...error) error {
	nonNil := errs[:0]
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return errors.Join(nonNil...)
	}
}

// bindParams binds positional args to parameter names, applying defaults. A
// variadic parameter collects every remaining argument into one array, an
// empty one when the caller stopped short of it, and is the last binding
// either way, as the parser allows `...` only in final position.
func (rt *Runtime) bindParams(decl *model.FuncDecl, args []any, scope *Scope) error {
	args, err := rt.phpxNamedArgs(decl, args, scope)
	if err != nil {
		return err
	}
	for i, p := range decl.Params {
		if p.Variadic {
			// min, because a caller that stopped short of the variadic
			// parameter leaves i past the end: `f($a, $b, ...$rest)` called as
			// `f("x")` reaches here with i at 2 and one argument, and args[i:]
			// panics on the slice bound.
			tail := args[min(i, len(args)):]
			rest := model.NewArraySize(len(tail))
			for _, arg := range tail {
				rest.Append(arg)
			}
			scope.Set(p.Name, rest)
			return nil
		}
		if i < len(args) {
			scope.Set(p.Name, phpxOwnedArray(args[i]))
			continue
		}
		if p.Default != nil {
			v, err := rt.Eval(p.Default, scope)
			if err != nil {
				return err
			}
			scope.Set(p.Name, v)
			continue
		}
		scope.Set(p.Name, nil)
	}
	return nil
}
