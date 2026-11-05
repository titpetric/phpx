package expr

import (
	"errors"
	"strings"
)

// isQuietCall names the language constructs that answer for a value out of
// reach and never raise: isset and empty.
func isQuietCall(name string) bool {
	if i := strings.LastIndexByte(name, '\\'); i >= 0 {
		name = name[i+1:]
	}
	return strings.EqualFold(name, "isset") || strings.EqualFold(name, "empty")
}

// quietOnRefusal turns an error that marks itself Quiet, the refusal of a
// protected or private property, into null: isset() is false and empty() is
// true, as in php. Any other error passes through.
func quietOnRefusal(c closure) closure {
	return func(env *Env) (any, error) {
		v, err := c(env)
		var quiet interface{ Quiet() bool }
		if err != nil && errors.As(err, &quiet) && quiet.Quiet() {
			return nil, nil
		}
		return v, err
	}
}
