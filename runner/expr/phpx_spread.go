package expr

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// spreadOp is the operator the parser gives `...$args` in an argument list:
// a Unary whose operand is the array to unpack.
const spreadOp = "..."

// spreadValues is what a spread argument evaluates to: the elements to splice
// into the argument list in its place.
type spreadValues []any

func isSpread(e model.Expr) bool {
	u, ok := e.(*model.Unary)
	return ok && u.Op == spreadOp && !u.Postfix
}

func hasSpread(args []model.Expr) bool {
	for _, a := range args {
		if isSpread(a) {
			return true
		}
	}
	return false
}

// phpxSpread compiles a call that unpacks an argument. It reports false for
// every other expression, and the dispatcher carries on as upstream does.
func (dc *directCompiler) phpxSpread(e model.Expr) (closure, bool, error) {
	var (
		name   string
		prefix []closure
		args   []model.Expr
		err    error
	)
	switch n := e.(type) {
	case *model.Call:
		if n.FirstClass || !hasSpread(n.Args) {
			return nil, false, nil
		}
		args = n.Args
		if n.Fallback != "" || containsBackslash(n.Name) {
			name, prefix = "__func", []closure{constClosure(n.Name), constClosure(n.Fallback)}
		} else {
			name = n.Name
			dc.addCall(n.Name)
		}
	case *model.StaticCall:
		if n.FirstClass || !hasSpread(n.Args) {
			return nil, false, nil
		}
		var method closure
		if method, err = dc.nameOrExpr(n.Method, n.MethodExpr); err == nil {
			name, prefix, args = "__static", []closure{constClosure(n.Class), method}, n.Args
		}
	case *model.MethodCall:
		if n.FirstClass || !hasSpread(n.Args) {
			return nil, false, nil
		}
		var base, method closure
		if base, err = dc.compile(n.Base); err == nil {
			if method, err = dc.nameOrExpr(n.Method, n.MethodExpr); err == nil {
				name, prefix, args = "__call", []closure{base, method}, n.Args
			}
		}
	case *model.New:
		if !hasSpread(n.Args) {
			return nil, false, nil
		}
		var class closure
		if class, err = dc.nameOrExpr(n.Class, n.ClassExpr); err == nil {
			name, prefix, args = "__new", []closure{class}, n.Args
		}
	case *model.Invoke:
		if n.FirstClass || !hasSpread(n.Args) {
			return nil, false, nil
		}
		var callee closure
		if callee, err = dc.compile(n.Callee); err == nil {
			name, prefix, args = "__invoke", []closure{callee}, n.Args
		}
	default:
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	all := append([]closure(nil), prefix...)
	for _, a := range args {
		c, err := dc.compileSpreadArg(a)
		if err != nil {
			return nil, true, err
		}
		all = append(all, c)
	}
	return spreadCall(name, all), true, nil
}

// compileSpreadArg compiles one argument: a spread yields its elements, any
// other expression compiles as it does everywhere else.
func (dc *directCompiler) compileSpreadArg(a model.Expr) (closure, error) {
	if !isSpread(a) {
		return dc.compile(a)
	}
	x, err := dc.compile(a.(*model.Unary).X)
	if err != nil {
		return nil, err
	}
	return func(env *Env) (any, error) {
		v, err := x(env)
		if err != nil {
			return nil, err
		}
		switch t := v.(type) {
		case *model.Array:
			out := make(spreadValues, 0, t.Len())
			t.Range(func(_, val any) bool {
				out = append(out, val)
				return true
			})
			return out, nil
		case []any:
			return spreadValues(t), nil
		}
		return nil, fmt.Errorf("Only arrays and Traversables can be unpacked")
	}, nil
}

// spreadCall is baseCall for an argument list that holds a spread: it splices
// each spreadValues into the slice the callee receives.
func spreadCall(name string, args []closure) closure {
	return func(env *Env) (any, error) {
		fn, ok := env.Base[name].(func(...any) (any, error))
		if !ok {
			return nil, fmt.Errorf("cannot call %s: not a function", name)
		}
		in := make([]any, 0, len(args))
		for _, c := range args {
			v, err := c(env)
			if err != nil {
				return nil, err
			}
			if s, ok := v.(spreadValues); ok {
				in = append(in, s...)
				continue
			}
			in = append(in, v)
		}
		return fn(in...)
	}
}
