//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerArraysExtra)
}

// entry is one key and value of an array a binding walks.
type entry struct{ key, val any }

func entriesOf(arr any) []entry {
	var out []entry
	model.RangeValues(arr, func(k, v any) bool {
		out = append(out, entry{k, v})
		return true
	})
	return out
}

// setOp describes one of the array_diff and array_intersect family: whether
// it keeps what the other arrays lack or share, what it matches on, and which
// of its trailing arguments are comparison callbacks.
type setOp struct {
	intersect bool
	match     string // "value", "key" or "assoc"
	valueCB   bool
	keyCB     bool
}

var setOps = map[string]setOp{
	"array_udiff":             {false, "value", true, false},
	"array_diff_ukey":         {false, "key", false, true},
	"array_diff_uassoc":       {false, "assoc", false, true},
	"array_udiff_assoc":       {false, "assoc", true, false},
	"array_udiff_uassoc":      {false, "assoc", true, true},
	"array_uintersect":        {true, "value", true, false},
	"array_intersect_assoc":   {true, "assoc", false, false},
	"array_uintersect_assoc":  {true, "assoc", true, false},
	"array_intersect_ukey":    {true, "key", false, true},
	"array_intersect_uassoc":  {true, "assoc", false, true},
	"array_uintersect_uassoc": {true, "assoc", true, true},
}

func registerArraysExtra(rt *runner.Runtime) {
	for name, op := range setOps {
		rt.RegisterFunc(name, func(args ...any) (any, error) { return runSetOp(rt, name, op, args) })
	}
	rt.RegisterFunc("array_merge_recursive", func(arrays ...any) any { return mergeRecursive(arrays) })
	rt.RegisterFunc("array_pad", func(arr any, size int64, value any) any {
		items := entriesOf(arr)
		out := model.NewArray()
		missing := int(abs64(size)) - len(items)
		if size < 0 {
			for ; missing > 0; missing-- {
				out.Append(value)
			}
		}
		for _, e := range items {
			if _, isInt := e.key.(int64); isInt {
				out.Append(e.val)
			} else {
				out.Set(e.key, e.val)
			}
		}
		if size > 0 {
			for ; missing > 0; missing-- {
				out.Append(value)
			}
		}
		return out
	})
	rt.RegisterFunc("array_change_key_case", func(arr any, which ...int64) any {
		upper := len(which) > 0 && which[0] == 1
		out := model.NewArray()
		for _, e := range entriesOf(arr) {
			k := e.key
			if s, ok := k.(string); ok {
				if upper {
					k = strings.ToUpper(s)
				} else {
					k = strings.ToLower(s)
				}
			}
			out.Set(k, e.val)
		}
		return out
	})
	rt.RegisterFunc("array_count_values", func(arr any) any {
		out := model.NewArray()
		for _, e := range entriesOf(arr) {
			switch e.val.(type) {
			case int64, string:
			default:
				continue
			}
			key := phpval.Key(e.val)
			n, _ := out.Get(key)
			count, _ := n.(int64)
			out.Set(key, count+1)
		}
		return out
	})
	registerArrayPredicates(rt)
	registerArrayWalk(rt)
	registerNatSort(rt)
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// runSetOp answers array_diff_* and array_intersect_*: the arrays, then the
// callbacks the function takes, value comparison before key comparison.
func runSetOp(rt *runner.Runtime, name string, op setOp, args []any) (any, error) {
	callbacks := 0
	if op.valueCB {
		callbacks++
	}
	if op.keyCB {
		callbacks++
	}
	if len(args) < 1+callbacks {
		return nil, &valueError{message: name + "(): at least one array and its callback are required"}
	}
	arrays := args[:len(args)-callbacks]
	cbs := args[len(args)-callbacks:]
	valueEq := func(a, b any) (bool, error) { return phpval.String(a) == phpval.String(b), nil }
	keyEq := func(a, b any) (bool, error) { return phpval.String(a) == phpval.String(b), nil }
	next := 0
	if op.valueCB {
		valueEq = callbackEq(rt, cbs[next])
		next++
	}
	if op.keyCB {
		keyEq = callbackEq(rt, cbs[next])
	}
	out := model.NewArray()
	for _, e := range entriesOf(arrays[0]) {
		inAll, inAny := true, false
		for _, other := range arrays[1:] {
			found := false
			for _, o := range entriesOf(other) {
				var ok bool
				var err error
				switch op.match {
				case "value":
					ok, err = valueEq(e.val, o.val)
				case "key":
					ok, err = keyEq(e.key, o.key)
				default:
					if ok, err = keyEq(e.key, o.key); ok && err == nil {
						ok, err = valueEq(e.val, o.val)
					}
				}
				if err != nil {
					return nil, err
				}
				if ok {
					found = true
					break
				}
			}
			inAll = inAll && found
			inAny = inAny || found
		}
		if op.intersect && inAll || !op.intersect && !inAny {
			out.Set(e.key, e.val)
		}
	}
	return out, nil
}

// callbackEq adapts a php comparison callback, which answers 0 for equal.
func callbackEq(rt *runner.Runtime, cb any) func(a, b any) (bool, error) {
	return func(a, b any) (bool, error) {
		fn, ok := rt.Callable(cb)
		if !ok {
			return false, &valueError{message: "Argument must be a valid callback"}
		}
		r, err := fn(a, b)
		return phpval.Int(r) == 0, err
	}
}

// mergeRecursive is array_merge_recursive: string keys that collide merge into
// an array, integer keys are renumbered.
func mergeRecursive(arrays []any) any {
	out := model.NewArray()
	for _, arr := range arrays {
		for _, e := range entriesOf(arr) {
			if _, isInt := e.key.(int64); isInt {
				out.Append(e.val)
				continue
			}
			prev, exists := out.Get(e.key)
			if !exists {
				out.Set(e.key, e.val)
				continue
			}
			left := prev
			if !model.IsCollection(left) {
				l := model.NewArray()
				l.Append(prev)
				left = l
			}
			right := e.val
			if !model.IsCollection(right) {
				r := model.NewArray()
				r.Append(e.val)
				right = r
			}
			out.Set(e.key, mergeRecursive([]any{left, right}))
		}
	}
	return out
}
