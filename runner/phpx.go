//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxGlobal binds every name in `global $a, $b;` to the global frame, which is
// what php does: the name reads and writes the variable the file body holds.
//
// The binding is the static mechanism pointed somewhere else. bindStatic aims a
// name at a bag keyed by that name, Get and Set resolve through the bag, and
// the global frame's own table is a bag of that shape. Six php behaviours come
// out of it without another line: a read sees the global, a write reaches it, a
// name the statement already bound stays bound, unset() drops the binding and
// leaves the global, a local value assigned before the statement falls out of
// reach, and the statement's own name list needs no order.
//
// A name the global frame does not hold is created there as null, because php
// makes the global exist the moment the statement names it.
func (rt *Runtime) phpxGlobal(n *model.Global, scope *Scope) error {
	if len(rt.frames) == 0 {
		return nil
	}
	globals := rt.frames[0]
	if globals == scope {
		// In the file body the name is the global already.
		return nil
	}
	for _, name := range n.Names {
		if _, ok := globals.vars[name]; !ok {
			globals.vars[name] = nil
		}
		scope.bindStatic(name, globals.vars)
	}
	return nil
}
