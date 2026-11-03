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
