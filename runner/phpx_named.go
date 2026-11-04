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
	ordered := make([]any, len(decl.Params))
	given := make([]bool, len(decl.Params))
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
			return nil, phpxError("Unknown named parameter $" + na.Name)
		}
		if given[at] {
			return nil, phpxError("Named parameter $" + na.Name + " overwrites previous argument")
		}
		ordered[at], given[at] = na.Val, true
	}
	last := -1
	for i, ok := range given {
		if ok {
			last = i
		}
	}
	for i := 0; i < last; i++ {
		if given[i] {
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
	return ordered[:last+1], nil
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
func phpxEntryArgs(fn any, args []any) []any {
	if !expr.NamedSeen.Load() {
		return args
	}
	if _, user := fn.(func(...any) (any, error)); user {
		return args
	}
	return phpxBareArgs(args)
}
