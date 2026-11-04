//go:build phpx

package phpx

import (
	"reflect"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxJSONEncode)
}

// registerPhpxJSONEncode makes json_encode read an object as php does: a
// JsonSerializable answers jsonSerialize(), and any other object is its public
// properties. Upstream encodes every property it holds.
func registerPhpxJSONEncode(rt *runner.Runtime) {
	fn, ok := rt.LookupFunc("json_encode")
	if !ok {
		return
	}
	ft := reflect.TypeOf(fn)
	if ft == nil || ft.Kind() != reflect.Func || ft.NumIn() < 1 {
		return
	}
	inner := reflect.ValueOf(fn)
	wrapped := reflect.MakeFunc(ft, func(args []reflect.Value) []reflect.Value {
		if !args[0].CanInterface() {
			return inner.Call(args)
		}
		prepared, err := phpxJSONValue(rt, args[0].Interface(), 0)
		if err == nil {
			arg := reflect.ValueOf(&prepared).Elem()
			if arg.Type().AssignableTo(ft.In(0)) {
				args[0] = arg
			}
		}
		var out []reflect.Value
		if ft.IsVariadic() {
			out = inner.CallSlice(args)
		} else {
			out = inner.Call(args)
		}
		if len(out) > 0 {
			v := out[0]
			if v.Kind() == reflect.Interface && !v.IsNil() {
				v = v.Elem()
			}
			if v.Kind() == reflect.String {
				fixed := reflect.ValueOf(phpxUnescapeHTML(v.String())).Convert(v.Type())
				if out[0].Kind() == reflect.Interface {
					out[0] = reflect.ValueOf(&[]any{fixed.Interface()}[0]).Elem()
				} else {
					out[0] = fixed
				}
			}
		}
		return out
	})
	rt.RegisterFunc("json_encode", wrapped.Interface())
}

// phpxJSONValue replaces each script object under v by what json_encode writes
// for it.
func phpxJSONValue(rt *runner.Runtime, v any, depth int) (any, error) {
	if depth > 64 {
		return v, nil
	}
	switch t := v.(type) {
	case *model.Array:
		out := model.NewArraySize(t.Len())
		var err error
		t.Range(func(k, val any) bool {
			var next any
			if next, err = phpxJSONValue(rt, val, depth+1); err != nil {
				return false
			}
			out.Set(k, next)
			return true
		})
		return out, err
	case *model.Object:
		if t.Class == nil {
			return v, nil
		}
		if res, found, err := rt.PhpxCallMethod(t, "jsonSerialize"); found || err != nil {
			if err != nil {
				return nil, err
			}
			return phpxJSONValue(rt, res, depth+1)
		}
		out := model.NewArray()
		var err error
		rt.PhpxPublicProps(t, func(name string, val any) bool {
			var next any
			if next, err = phpxJSONValue(rt, val, depth+1); err != nil {
				return false
			}
			out.Set(name, next)
			return true
		})
		if out.Len() == 0 {
			return v, err // an empty object stays {}
		}
		return out, err
	}
	return v, nil
}

// phpxUnescapeHTML undoes Go's \u003c, \u003e and \u0026, which php writes as
// the characters themselves unless JSON_HEX_TAG and JSON_HEX_AMP are given. A
// backslash that is itself escaped starts no sequence.
func phpxUnescapeHTML(s string) string {
	if !strings.Contains(s, `\u00`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+5 < len(s) && s[i+1] == 'u' {
			switch s[i : i+6] {
			case `\u003c`:
				b.WriteByte('<')
				i += 5
				continue
			case `\u003e`:
				b.WriteByte('>')
				i += 5
				continue
			case `\u0026`:
				b.WriteByte('&')
				i += 5
				continue
			}
		}
		b.WriteByte(s[i])
		if i+1 < len(s) {
			i++
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
