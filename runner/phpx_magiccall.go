//go:build phpx

package runner

import (
	"slices"

	"github.com/titpetric/phpx/model"
)

// phpxMagicCall sends a call to a method the class does not declare to its
// __call($name, $arguments), as php does.
func (rt *Runtime) phpxMagicCall(obj *model.Object, method string, args []any, scope *Scope) (any, bool, error) {
	decl, ok := lookupPHPMethod(obj.Class, "__call")
	if !ok {
		return nil, false, nil
	}
	list := model.NewArraySize(len(args))
	for _, a := range slices.Clone(args) {
		list.Append(a)
	}
	out, err := rt.invokeMethod(obj, decl, []any{method, list}, scope)
	return out, true, err
}
