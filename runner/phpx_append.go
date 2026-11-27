//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxAppendSlice grows a Go slice a binding returned where php writes
// `$a[] =`: its elements become the array php holds, and the array is stored
// back through target's base, because a slice cannot grow through the
// interface that holds it. A container that is neither is left to the caller.
func (rt *Runtime) phpxAppendSlice(target *model.Index, base, value any, scope *Scope) (any, bool, error) {
	arr, ok := phpxOwnedArray(base).(*model.Array)
	if !ok {
		return nil, false, nil
	}
	arr.Append(value)
	if err := rt.assignTo(target.Base, arr, scope); err != nil {
		return nil, true, err
	}
	return value, true, nil
}
