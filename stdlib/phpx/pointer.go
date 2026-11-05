//go:build phpx

package phpx

import (
	"reflect"
	"sync"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Pointer`, registerPhpxPointer)
}

// pointers holds the internal position php keeps on every array, which
// model.Array does not carry. An array not listed is at its first element.
var pointers = struct {
	sync.Mutex
	at map[any]int
}{at: map[any]int{}}

// pointerKey identifies the array a position belongs to: the handle of a
// *model.Array, and the backing store of a Go slice a binding returned.
func pointerKey(v any) any {
	if arr, ok := v.(*model.Array); ok {
		return arr
	}
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice {
		return [2]uintptr{rv.Pointer(), uintptr(rv.Len())}
	}
	return v
}

// elementsOf lists the keys and values php walks for v.
func elementsOf(v any) (keys, vals []any) {
	if arr, ok := v.(*model.Array); ok {
		arr.Range(func(k, val any) bool {
			keys, vals = append(keys, k), append(vals, val)
			return true
		})
		return keys, vals
	}
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			keys, vals = append(keys, int64(i)), append(vals, rv.Index(i).Interface())
		}
	}
	return keys, vals
}

// pointerOf reads the position of arr, and the keys and values it walks.
func pointerOf(arr any) (pos int, keys, vals []any) {
	keys, vals = elementsOf(arr)
	pointers.Lock()
	defer pointers.Unlock()
	return pointers.at[pointerKey(arr)], keys, vals
}

func setPointer(arr any, pos int) {
	pointers.Lock()
	pointers.at[pointerKey(arr)] = pos
	pointers.Unlock()
}

// registerPhpxPointer installs current, key, next, prev, reset and end.
func registerPhpxPointer(rt *runner.Runtime) {
	valueAt := func(arr any, move func(pos, n int) int) any {
		pos, _, vals := pointerOf(arr)
		pos = move(pos, len(vals))
		setPointer(arr, pos)
		if pos < 0 || pos >= len(vals) {
			return false
		}
		return vals[pos]
	}
	stay := func(pos, _ int) int { return pos }
	rt.RegisterFunc("current", func(arr any) any { return valueAt(arr, stay) })
	rt.RegisterFunc("pos", func(arr any) any { return valueAt(arr, stay) })
	rt.RegisterFunc("next", func(arr any) any {
		return valueAt(arr, func(pos, n int) int { return min(pos+1, n) })
	})
	rt.RegisterFunc("prev", func(arr any) any {
		return valueAt(arr, func(pos, n int) int {
			if pos <= 0 || pos > n {
				return n // php leaves the pointer past the ends
			}
			return pos - 1
		})
	})
	rt.RegisterFunc("reset", func(arr any) any { return valueAt(arr, func(_, _ int) int { return 0 }) })
	rt.RegisterFunc("end", func(arr any) any {
		return valueAt(arr, func(_, n int) int { return max(n-1, 0) })
	})
	rt.RegisterFunc("key", func(arr any) any {
		pos, keys, _ := pointerOf(arr)
		if pos < 0 || pos >= len(keys) {
			return nil
		}
		return keys[pos]
	})
}
