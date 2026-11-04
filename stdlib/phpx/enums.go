//go:build phpx

package phpx

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxEnums)
}

// valueError is the ValueError a failed from() raises.
type valueError struct{ message string }

func (e *valueError) Error() string          { return e.message }
func (e *valueError) ThrowableClass() string { return "ValueError" }

// registerPhpxEnums installs what the parser lowers an enum to: the function
// that builds a case once, and from()/tryFrom().
func registerPhpxEnums(rt *runner.Runtime) {
	cases := map[string]any{}
	rt.RegisterFunc("__phpx_enum_case", func(enum, name string, value any) (any, error) {
		key := strings.ToLower(enum) + "::" + name
		if c, ok := cases[key]; ok {
			return c, nil
		}
		obj, err := rt.PhpxNewWithoutConstructor(enum)
		if err != nil {
			return nil, err
		}
		o := obj.(*model.Object)
		runner.PhpxSetProp(o, "name", name)
		runner.PhpxSetProp(o, "value", value)
		cases[key] = o
		return o, nil
	})
	rt.RegisterFunc("__phpx_enum_from", func(enum string, value any, strict bool) (any, error) {
		all, _, callErr := rt.PhpxStaticCall(enum, "cases")
		if callErr != nil {
			return nil, callErr
		}
		if arr, ok := all.(*model.Array); ok {
			var found any
			arr.Range(func(_, c any) bool {
				if o, ok := c.(*model.Object); ok && phpxSameBacking(o.Props["value"], value) {
					found = o
					return false
				}
				return true
			})
			if found != nil {
				return found, nil
			}
		}
		if !strict {
			return nil, nil
		}
		shown := fmt.Sprint(value)
		if s, ok := value.(string); ok {
			shown = `"` + s + `"`
		}
		return nil, &valueError{message: shown + " is not a valid backing value for enum " + enum}
	})
}

// phpxSameBacking compares a case's value with an argument as php's from()
// does: an int-backed enum takes a numeric string, a string-backed one a string.
func phpxSameBacking(have, want any) bool {
	switch h := have.(type) {
	case int64:
		switch w := want.(type) {
		case int64:
			return h == w
		case int:
			return h == int64(w)
		case string:
			return fmt.Sprint(h) == w
		}
	case string:
		if w, ok := want.(string); ok {
			return h == w
		}
		return h == fmt.Sprint(want)
	}
	return have == want
}
