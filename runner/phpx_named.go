//go:build phpx

package runner

import (
	"fmt"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner/expr"
)

// phpxError is a refused call, which a script catches as Error.
type phpxError string

func (e phpxError) Error() string { return string(e) }

// ThrowableClass makes the refusal an Error.
func (e phpxError) ThrowableClass() string { return "Error" }

// phpxNamedArgs orders the named arguments of a call into the positions of the
// parameters they name, and fills a position nobody gave with the parameter's
// default, which php does too. A call with no named argument returns args as
// it came. A name no parameter has is php's Error.
func (rt *Runtime) phpxNamedArgs(decl *model.FuncDecl, args []any, scope *Scope) ([]any, error) {
	first := -1
	for i, a := range args {
		if _, ok := a.(expr.NamedArg); ok {
			first = i
			break
		}
	}
	if first < 0 {
		return args, nil
	}
	// The positional arguments are kept whole: the ones past the declared
	// parameters belong to a variadic.
	ordered := make([]any, max(len(decl.Params), first))
	given := make([]bool, len(ordered))
	var extras []expr.NamedArg
	n := copy(ordered, args[:first])
	for i := 0; i < n; i++ {
		given[i] = true
	}
	for _, a := range args[first:] {
		na, ok := a.(expr.NamedArg)
		if !ok {
			return nil, phpxError("Cannot use positional argument after named argument")
		}
		at := -1
		for i, p := range decl.Params {
			if p.Name == na.Name && !p.Variadic {
				at = i
				break
			}
		}
		if at < 0 {
			if variadic := phpxVariadicIndex(decl); variadic >= 0 {
				extras = append(extras, na)
				continue
			}
			return nil, phpxError("Unknown named parameter $" + na.Name)
		}
		if given[at] {
			return nil, phpxError("Named parameter $" + na.Name + " overwrites previous argument")
		}
		ordered[at], given[at] = na.Val, true
	}
	last := first - 1
	for i, ok := range given {
		if ok {
			last = i
		}
	}
	for i := 0; i < last; i++ {
		if given[i] {
			continue
		}
		if i >= len(decl.Params) {
			continue
		}
		p := decl.Params[i]
		if p.Default == nil {
			return nil, phpxError(fmt.Sprintf("Argument #%d ($%s) not passed", i+1, p.Name))
		}
		v, err := rt.Eval(p.Default, scope)
		if err != nil {
			return nil, err
		}
		ordered[i] = v
	}
	out := ordered[:last+1]
	if len(extras) > 0 {
		// Named arguments no parameter names are the variadic's, keyed by name.
		v := phpxVariadicIndex(decl)
		for len(out) < v {
			out = append(out, nil)
		}
		for _, na := range extras {
			out = append(out, na)
		}
	}
	return out, nil
}

// phpxVariadicIndex is the position of the variadic parameter, or -1.
func phpxVariadicIndex(decl *model.FuncDecl) int {
	for i, p := range decl.Params {
		if p.Variadic {
			return i
		}
	}
	return -1
}

// phpxVariadicAppend adds one collected argument to a variadic's array: a named
// one keeps its name.
func phpxVariadicAppend(rest *model.Array, arg any) {
	if na, ok := arg.(expr.NamedArg); ok {
		rest.Set(na.Name, na.Val)
		return
	}
	rest.Append(arg)
}

// phpxBareArgs strips the names off a call into a Go binding, which takes its
// arguments by position alone.
func phpxBareArgs(args []any) []any {
	for i, a := range args {
		if _, ok := a.(expr.NamedArg); !ok {
			continue
		}
		out := append([]any(nil), args...)
		for j := i; j < len(out); j++ {
			if na, ok := out[j].(expr.NamedArg); ok {
				out[j] = na.Val
			}
		}
		return out
	}
	return args
}

// phpxEntryArgs strips the names off the arguments of a call into a Go
// binding. A PHP function, registered with a plain variadic closure, keeps
// them for bindParams.
func phpxEntryArgs(args []any, fn any) []any {
	if !expr.NamedSeen.Load() {
		return args
	}
	if _, user := fn.(func(...any) (any, error)); user {
		return args
	}
	return phpxBareArgs(args)
}

// phpxArgsValue is what func_get_args() answers: an array a script can write
// to, where upstream answers the frame's own slice.
func phpxArgsValue(args []any) any {
	out := model.NewArraySize(len(args))
	for _, a := range args {
		out.Append(a)
	}
	return out
}

// phpxOwned is phpxOwnedArray for a program on this runtime. A program that
// imported a namespace (`use Runtime\Go`) holds slices it made on purpose, with
// a capacity and a length that survive a clear, and they stay slices.
func (rt *Runtime) phpxOwned(v any) any {
	if len(rt.imported) > 0 {
		return v
	}
	return phpxOwnedArray(v)
}

// phpxOwnedArray turns a Go slice a binding answered into the array a script
// can write to, where a variable takes it. Upstream keeps the slice it was
// given, which reads and iterates; array_pop($parts) and unset($parts[1]) after
// `$parts = array_reverse($list)` have nowhere to write, and php writes there.
func phpxOwnedArray(v any) any {
	switch t := v.(type) {
	case []any:
		out := model.NewArraySize(len(t))
		for _, e := range t {
			out.Append(e)
		}
		return out
	case []string:
		out := model.NewArraySize(len(t))
		for _, e := range t {
			out.Append(e)
		}
		return out
	}
	return v
}

// phpxInstallArgs adds func_num_args() and func_get_arg() beside func_get_args().
func (rt *Runtime) phpxInstallArgs(env map[string]any, ref *scopeRef) {
	env["func_num_args"] = adapt(func() int64 { return int64(len(funcGetArgs(ref.scope))) })
	env["func_get_arg"] = adapt(func(i int64) any {
		args := funcGetArgs(ref.scope)
		if i < 0 || int(i) >= len(args) {
			return nil
		}
		return args[i]
	})
}
