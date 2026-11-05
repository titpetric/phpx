//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\SPL`, registerPhpxSPL)
	runner.RegisterArea(`Runtime\PX\SPL\TemporaryArrays`, registerPhpxTemporaryArrays)
}

// registerPhpxSPL gives Countable and Traversable script objects the functions
// that read them: count, iterator_to_array, iterator_count and iterator_apply.
func registerPhpxSPL(rt *runner.Runtime) {
	if prev, ok := rt.LookupFunc("count"); ok {
		if upstream, ok := prev.(func(any) int64); ok {
			count := func(v any, _ ...any) (int64, error) {
				if obj, ok := v.(*model.Object); ok {
					n, found, err := rt.PhpxCallMethod(obj, "count")
					if err != nil {
						return 0, err
					}
					if found {
						switch i := n.(type) {
						case int64:
							return i, nil
						case int:
							return int64(i), nil
						case float64:
							return int64(i), nil
						}
						return 0, nil
					}
				}
				return upstream(v), nil
			}
			rt.RegisterFunc("count", count)
			rt.RegisterFunc("sizeof", count)
		}
	}
	// walk visits a Traversable (an Iterator, an IteratorAggregate or an array)
	// in order.
	walk := func(v any, visit func(k, val any) bool) error {
		switch t := v.(type) {
		case *model.Array:
			t.Range(visit)
		case *model.Object:
			if done, err := rt.PhpxIterate(t, visit); err != nil || done {
				return err
			}
			t.Range(func(name string, val any) bool { return visit(name, val) })
		case runner.PhpxHostIterator:
			return runner.PhpxWalkHost(t, visit)
		}
		return nil
	}
	rt.RegisterFunc("iterator_to_array", func(it any, preserve ...bool) (*model.Array, error) {
		out := model.NewArray()
		keep := len(preserve) == 0 || preserve[0]
		err := walk(it, func(k, val any) bool {
			if keep {
				out.Set(k, val)
			} else {
				out.Append(val)
			}
			return true
		})
		return out, err
	})
	rt.RegisterFunc("iterator_count", func(it any) (int64, error) {
		var n int64
		err := walk(it, func(_, _ any) bool { n++; return true })
		return n, err
	})
}

// registerPhpxTemporaryArrays lets array_pop and array_shift take the Go slice
// a call answered, `array_pop(explode('.', $name))`. php passes the temporary
// and discards what the call leaves of it.
func registerPhpxTemporaryArrays(rt *runner.Runtime) {
	for _, name := range []string{"array_pop", "array_shift"} {
		prev, ok := rt.LookupFunc(name)
		if !ok {
			continue
		}
		upstream, ok := prev.(func(any) (any, error))
		if !ok {
			continue
		}
		rt.RegisterFunc(name, func(v any) (any, error) {
			switch s := v.(type) {
			case []string:
				arr := model.NewArraySize(len(s))
				for _, e := range s {
					arr.Append(e)
				}
				v = arr
			case []any:
				arr := model.NewArraySize(len(s))
				for _, e := range s {
					arr.Append(e)
				}
				v = arr
			}
			return upstream(v)
		})
	}
}
