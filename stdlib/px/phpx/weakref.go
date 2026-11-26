//go:build phpx

package phpx

import (
	"weak"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\WeakReference`, registerPhpxWeakReference)
}

// registerPhpxWeakReference installs WeakReference, which holds an instance
// without keeping it alive: a parser stores a weak reference to itself in a
// static table so the table does not close a cycle.
func registerPhpxWeakReference(rt *runner.Runtime) {
	rt.RegisterConstructor("WeakReference", func() *phpxWeakReference { return &phpxWeakReference{} })
	rt.RegisterFunc("WeakReference::create", func(obj *model.Object) *phpxWeakReference {
		if obj == nil {
			return &phpxWeakReference{}
		}
		return &phpxWeakReference{obj: weak.Make(obj)}
	})
}

// phpxWeakReference is the object WeakReference::create answers.
type phpxWeakReference struct {
	obj weak.Pointer[model.Object]
}

// Get answers the instance while anything else still holds it, or nil.
func (w *phpxWeakReference) Get() any {
	if w == nil || w.obj == (weak.Pointer[model.Object]{}) {
		return nil
	}
	if obj := w.obj.Value(); obj != nil {
		return obj
	}
	return nil
}
