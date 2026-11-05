//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.PhpxRegisterLate(registerResources)
}

// registerResources teaches the type functions that a closed stream is still a
// resource of a kind: gettype says "resource (closed)", is_resource says no.
func registerResources(rt *runner.Runtime) {
	if closeFn, ok := rt.Callable("fclose"); ok {
		rt.RegisterFunc("fclose", func(h any) (any, error) {
			res, err := closeFn(h)
			if closed, isBool := res.(bool); isBool && closed && err == nil {
				runner.PhpxCloseResource(h)
			}
			return res, err
		})
	}
	prior := map[string]func(...any) (any, error){}
	for _, name := range []string{"is_resource", "gettype", "get_debug_type", "is_object"} {
		if fn, ok := rt.Callable(name); ok {
			prior[name] = fn
		}
	}
	rt.RegisterFunc("is_resource", func(v any) bool {
		return runner.PhpxIsResource(v) && !runner.PhpxResourceClosed(v)
	})
	rt.RegisterFunc("gettype", func(v any) (any, error) {
		switch {
		case runner.PhpxIsResource(v) && runner.PhpxResourceClosed(v):
			return "resource (closed)", nil
		case runner.PhpxIsResource(v):
			return "resource", nil
		}
		return prior["gettype"](v)
	})
	rt.RegisterFunc("get_debug_type", func(v any) (any, error) {
		switch {
		case runner.PhpxIsResource(v) && runner.PhpxResourceClosed(v):
			return "resource (closed)", nil
		case runner.PhpxIsResource(v):
			return "resource (stream)", nil
		}
		return prior["get_debug_type"](v)
	})
	if isObject, ok := prior["is_object"]; ok {
		rt.RegisterFunc("is_object", func(v any) (any, error) {
			if runner.PhpxIsResource(v) {
				return false, nil
			}
			return isObject(v)
		})
	}
	rt.RegisterFunc("get_resource_type", func(v any) string {
		if runner.PhpxResourceClosed(v) {
			return "Unknown"
		}
		return "stream"
	})
	rt.RegisterFunc("get_resource_id", func(v any) int64 { return runner.PhpxResourceID(v) })
}
