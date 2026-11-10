package expr

import (
	"reflect"
	"sort"

	"github.com/titpetric/phpx/model"
)

// goPairs reads a Go slice or string-keyed map a binding returned as the
// array php would hold: a map spreads in key order, since Go keeps none.
func goPairs(v any) (*model.Array, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		out := model.NewArray()
		for i := 0; i < rv.Len(); i++ {
			out.Append(rv.Index(i).Interface())
		}
		return out, true
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return nil, false
		}
		keys := rv.MapKeys()
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		out := model.NewArray()
		for _, k := range keys {
			out.Set(k.String(), rv.MapIndex(k).Interface())
		}
		return out, true
	}
	return nil, false
}
