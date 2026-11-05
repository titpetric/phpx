//go:build phpx

package runner

import (
	"fmt"
	"github.com/titpetric/phpx/parser"
	"reflect"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxMethod finds a method of obj's class by its php name, which is case
// insensitive.
func phpxMethod(obj *model.Object, name string) (*model.FuncDecl, bool) {
	if obj == nil || obj.Class == nil {
		return nil, false
	}
	if m, ok := obj.Class.Methods[name]; ok {
		return m, true
	}
	for key, m := range obj.Class.Methods {
		if strings.EqualFold(key, name) {
			return m, true
		}
	}
	return nil, false
}

// PhpxCallMethod calls a method of a script object whatever its visibility. It
// reports false when the class declares none by that name.
func (rt *Runtime) PhpxCallMethod(obj *model.Object, name string, args ...any) (any, bool, error) {
	decl, ok := phpxMethod(obj, name)
	if !ok {
		return nil, false, nil
	}
	v, err := rt.invokeMethod(obj, decl, args, rt.newScope())
	return v, true, err
}

// phpxInstallSPL adds the helper compiled expressions read an ArrayAccess
// object through: __phpx_offset(op, object, key[, value]).
func (rt *Runtime) phpxInstallSPL(env map[string]any) {
	env["__phpx_offset"] = adapt(func(op string, obj *model.Object, rest ...any) (any, error) {
		return rt.phpxOffset(op, obj, rest...)
	})
	env["__phpx_quiet"] = adapt(func(on, final bool) any {
		if on {
			rt.phpxQuietEnter(final)
		} else {
			rt.phpxQuietLeave()
		}
		return nil
	})
}

// phpxOffset runs one ArrayAccess operation: get, exists, set, unset.
func (rt *Runtime) phpxOffset(op string, obj *model.Object, rest ...any) (any, error) {
	method := map[string]string{"get": "offsetGet", "exists": "offsetExists", "set": "offsetSet", "unset": "offsetUnset"}[op]
	var key any
	if len(rest) > 0 {
		key = rest[0]
	}
	args := []any{key}
	if op == "set" && len(rest) > 1 {
		args = append(args, rest[1])
	}
	v, ok, err := rt.PhpxCallMethod(obj, method, args...)
	if !ok {
		return nil, fmt.Errorf("Cannot use object of type %s as array", obj.Class.Name)
	}
	return v, err
}

// phpxArrayAccessSet writes `$object[$k] = v`, `$object[] = v` and the compound
// forms. It reports false when base is not a script object with offsetSet.
func (rt *Runtime) phpxArrayAccessSet(target *model.Index, base any, op string, rhs any, scope *Scope) (bool, error) {
	obj, ok := base.(*model.Object)
	if !ok {
		return false, nil
	}
	if _, has := phpxMethod(obj, "offsetSet"); !has {
		return false, nil
	}
	var key any
	if target.Index != nil && op != "[]=" {
		k, err := rt.Eval(target.Index, scope)
		if err != nil {
			return true, err
		}
		key = k
	}
	if op != "=" && op != "" && op != "[]=" {
		cur, err := rt.phpxOffset("get", obj, key)
		if err != nil {
			return true, err
		}
		if rhs, err = applyAssignOp(op, cur, rhs); err != nil {
			return true, err
		}
	}
	_, err := rt.phpxOffset("set", obj, key, rhs)
	return true, err
}

// phpxArrayAccessUnset runs offsetUnset for `unset($object[$k])`.
func (rt *Runtime) phpxArrayAccessUnset(base, key any) bool {
	obj, ok := base.(*model.Object)
	if !ok {
		return false
	}
	if _, has := phpxMethod(obj, "offsetUnset"); !has {
		return false
	}
	_, _ = rt.phpxOffset("unset", obj, key)
	return true
}

// PhpxIterate walks a Traversable script object: an IteratorAggregate is
// resolved through getIterator() until an Iterator is reached, and an Iterator
// is driven through rewind, valid, current, key and next. It reports false for
// an object that is neither, which a foreach then reads as its properties.
func (rt *Runtime) PhpxIterate(obj *model.Object, yield func(key, value any) bool) (bool, error) {
	for depth := 0; depth < 16; depth++ {
		if _, ok := phpxMethod(obj, "getIterator"); ok {
			next, _, err := rt.PhpxCallMethod(obj, "getIterator")
			if err != nil {
				return true, err
			}
			switch it := next.(type) {
			case *model.Object:
				obj = it
				continue
			case *model.Array:
				it.Range(yield)
				return true, nil
			case PhpxHostIterator:
				return true, phpxForeachHost(it, yield)
			}
			return true, fmt.Errorf("getIterator() must return a Traversable")
		}
		break
	}
	for _, m := range []string{"rewind", "valid", "current", "key", "next"} {
		if _, ok := phpxMethod(obj, m); !ok {
			return false, nil
		}
	}
	if _, _, err := rt.PhpxCallMethod(obj, "rewind"); err != nil {
		return true, err
	}
	for {
		valid, _, err := rt.PhpxCallMethod(obj, "valid")
		if err != nil {
			return true, err
		}
		if !phpTruthy(valid) {
			return true, nil
		}
		value, _, err := rt.PhpxCallMethod(obj, "current")
		if err != nil {
			return true, err
		}
		key, _, err := rt.PhpxCallMethod(obj, "key")
		if err != nil {
			return true, err
		}
		if !yield(key, value) {
			return true, nil
		}
		if _, _, err := rt.PhpxCallMethod(obj, "next"); err != nil {
			return true, err
		}
	}
}

// phpxForeachObject drives a foreach over a Traversable object.
func (rt *Runtime) phpxForeachObject(obj *model.Object, iter func(k, v any) bool) (bool, error) {
	return rt.PhpxIterate(obj, iter)
}

// PhpxThrowValue is the error `throw $value` raises: a built-in throwable as
// itself, an instance of a declared class wrapped with its class.
func PhpxThrowValue(v any) error {
	if err, ok := v.(error); ok {
		return err
	}
	if obj, ok := v.(*model.Object); ok {
		return newObjectError(obj)
	}
	return fmt.Errorf("uncaught exception: %s", phpString(v))
}

// PhpxClone is php's `clone`: a new object of the same class holding a copy of
// every property, arrays by value and objects by handle, and then __clone() is
// called on the copy.
func (rt *Runtime) PhpxClone(v any) (any, error) {
	obj, ok := v.(*model.Object)
	if !ok {
		// A value a binding returned (DateTime, DateInterval) is a pointer to a
		// struct; the copy is a new struct holding the same fields.
		if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Struct {
			dup := reflect.New(rv.Elem().Type())
			dup.Elem().Set(rv.Elem())
			return dup.Interface(), nil
		}
		return nil, fmt.Errorf("__clone method called on non-object")
	}
	dup := model.NewObject(obj.Class)
	obj.Range(func(name string, val any) bool {
		dup.SetProp(name, model.CopyValue(val))
		return true
	})
	if _, _, err := rt.PhpxCallMethod(dup, "__clone"); err != nil {
		return nil, err
	}
	return dup, nil
}

// PhpxStaticCall calls a static method by class and name whatever its
// visibility, with the called class set to class.
func (rt *Runtime) PhpxStaticCall(class, method string, args ...any) (any, bool, error) {
	c, ok := rt.lookupClass(strings.TrimPrefix(class, "\\"))
	if !ok {
		return nil, false, nil
	}
	for name, decl := range c.Methods {
		if strings.EqualFold(name, method) {
			obj := model.NewObject(c)
			v, err := rt.invokeMethod(obj, decl, args, rt.newScope())
			return v, true, err
		}
	}
	return nil, false, nil
}

// PhpxPublicProps visits the public properties of obj in their order, which is
// what json_encode, get_object_vars from outside and a foreach from outside see.
func (rt *Runtime) PhpxPublicProps(obj *model.Object, yield func(name string, v any) bool) {
	info, _ := phpxHiddenProps.Load(obj.Class)
	table, _ := info.(map[string]phpxPropInfo)
	obj.Range(func(name string, v any) bool {
		if p, hidden := table[name]; hidden && p.visibility != "" && p.visibility != "public" {
			return true
		}
		return yield(name, v)
	})
}

// PhpxFrameMark is the depth of the frame stack.
func (rt *Runtime) PhpxFrameMark() int { return len(rt.frames) }

// PhpxFrameDetach takes the frames above mark off the stack and returns them,
// which is what a generator does when it yields.
func (rt *Runtime) PhpxFrameDetach(mark int) []*Scope {
	if mark > len(rt.frames) {
		return nil
	}
	saved := append([]*Scope(nil), rt.frames[mark:]...)
	rt.frames = rt.frames[:mark]
	return saved
}

// PhpxFrameAttach puts detached frames back on top, as a generator does on resume.
func (rt *Runtime) PhpxFrameAttach(saved []*Scope) { rt.frames = append(rt.frames, saved...) }

// PhpxEval runs php source as eval() does: declarations are hoisted, the
// statements run in a scope of their own, and a return gives the value. A parse
// error is a ParseError the script can catch.
func (rt *Runtime) PhpxEval(code string) (any, error) {
	prog, err := parser.ParseFile("eval()'d code", "<?php "+code)
	if err != nil {
		return nil, &phpxParseError{msg: err.Error()}
	}
	if err := rt.hoist(prog, "eval()'d code"); err != nil {
		return nil, err
	}
	value, _, runErr := rt.exec(prog.Stmts, rt.newScope())
	return value, runErr
}

// phpxParseError is the ParseError an eval of bad source raises.
type phpxParseError struct{ msg string }

func (e *phpxParseError) Error() string          { return e.msg }
func (e *phpxParseError) ThrowableClass() string { return "ParseError" }

// PhpxBacktrace is debug_backtrace(): one entry per running frame, the nearest
// first, each with the object and class of the method running in it. A frame
// holds no function name or call site here, so those entries are absent.
func (rt *Runtime) PhpxBacktrace(withObjects bool, limit int) *model.Array {
	out := model.NewArray()
	for i := len(rt.frames) - 1; i >= 0; i-- {
		if limit > 0 && out.Len() >= limit {
			break
		}
		scope := rt.frames[i]
		entry := model.NewArray()
		if class, ok := scope.Get("__class__"); ok {
			entry.Set("class", class)
		}
		if this, ok := scope.Get("this"); ok && this != nil {
			if withObjects {
				entry.Set("object", this)
			}
			entry.Set("type", "->")
		}
		entry.Set("function", "")
		out.Append(entry)
	}
	return out
}
