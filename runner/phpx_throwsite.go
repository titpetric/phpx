//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxStampThrowable records where a throwable was created, as php does at
// `new`: the file of the running scope and the line of the statement. Only an
// object carrying the root's `file` and `line` members is one.
func (rt *Runtime) phpxStampThrowable(obj *model.Object, scope *Scope) {
	if _, ok := obj.Prop("line"); !ok {
		return
	}
	if _, ok := obj.Prop("previous"); !ok {
		return
	}
	if scope != nil {
		if file, ok := scope.Get("__FILE__"); ok {
			obj.SetProp("file", file)
		}
	}
	obj.SetProp("line", int64(rt.currentSourceLine()))
}
