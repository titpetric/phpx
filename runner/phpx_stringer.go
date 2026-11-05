//go:build phpx

package runner

import (
	"fmt"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
)

// phpxBindStringer gives a class with a __toString method the converter
// phpval reads in a string context. The method arrives flattened, so an
// inherited or trait-supplied one binds the same way.
func (rt *Runtime) phpxBindStringer(c *model.Class) {
	if _, ok := lookupPHPMethod(c, "__toString"); !ok {
		return
	}
	// php adds the Stringable interface to a class that declares the method.
	c.Implements = phpxMergeNames(c.Implements, "stringable")
	c.Stringer = func(o *model.Object) (string, error) {
		call, ok := rt.boundMethod(o, "__toString", nil)
		if !ok {
			return "", &phpval.ConversionError{Class: c.Name}
		}
		out, err := call()
		if err != nil {
			return "", err
		}
		s, isStr := out.(string)
		if !isStr {
			return "", fmt.Errorf("%s::__toString(): Return value must be of type string", c.Name)
		}
		return s, nil
	}
}

// phpxStringOf is the string form echo and concatenation write: an object
// answers through StringContext, which raises php's Error for one that has no
// __toString, and every other value keeps phpString.
func phpxStringOf(v any) (string, error) {
	if o, ok := v.(*model.Object); ok {
		return phpval.StringContext(o)
	}
	if PhpxIsResource(v) {
		return phpxResourceString(v), nil
	}
	return phpString(v), nil
}

// phpxCastString is the (string) cast: an object with no string form raises
// php's Error. helperCast returns no error, so the refusal crosses its frame as
// the callbackPanic the call boundary and the expression engine unwrap.
func phpxCastString(v any) any {
	s, err := phpxStringOf(v)
	if err != nil {
		panic(callbackPanic{err: err})
	}
	return s
}
