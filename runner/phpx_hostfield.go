//go:build phpx

package runner

import (
	"reflect"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
)

// phpxHostField reads a public property of the Go value behind an object that
// extends a host class (DateInterval's y, m, d, h, i, s, f, invert and days).
func phpxHostField(host any, name string) (any, bool) {
	f, ok := hostFieldOf(host, name)
	if !ok {
		return nil, false
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return f.Int(), true
	case reflect.Float32, reflect.Float64:
		return f.Float(), true
	}
	return f.Interface(), true
}

// phpxHostFieldSet writes such a property and reports whether the host has it.
func phpxHostFieldSet(host any, name string, val any) bool {
	f, ok := hostFieldOf(host, name)
	if !ok || !f.CanSet() {
		return false
	}
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.SetInt(phpval.Int(val))
	case reflect.Float32, reflect.Float64:
		f.SetFloat(phpval.Float(val))
	case reflect.Interface:
		if val == nil {
			f.Set(reflect.Zero(f.Type()))
		} else {
			f.Set(reflect.ValueOf(val))
		}
	default:
		return false
	}
	return true
}

func hostFieldOf(host any, name string) (reflect.Value, bool) {
	rv := reflect.ValueOf(host)
	if rv.Kind() != reflect.Pointer || rv.IsNil() || rv.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	st := rv.Elem()
	for i := 0; i < st.NumField(); i++ {
		sf := st.Type().Field(i)
		if sf.IsExported() && !sf.Anonymous && strings.EqualFold(sf.Name, name) {
			return st.Field(i), true
		}
	}
	return reflect.Value{}, false
}
