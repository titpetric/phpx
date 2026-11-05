//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxCaptureRef makes the variable of a by-reference `use` a shared cell. The
// enclosing frame and every closure that captures the name read and write the
// one bag, so a value assigned after the closure was made, or by the closure
// after the frame ended, is the one both see.
func phpxCaptureRef(refs map[string]map[string]any, use model.ClosureUse, scope *Scope) map[string]map[string]any {
	if !use.ByRef {
		return refs
	}
	if refs == nil {
		refs = map[string]map[string]any{}
	}
	refs[use.Name] = scope.refCell(use.Name)
	return refs
}

// phpxSyncRefs binds the closure frame's by-reference names to their cells.
func (rt *Runtime) phpxSyncRefs(env closureEnv, scope *Scope) func() {
	for name, cell := range env.refs {
		scope.bindStatic(name, cell)
	}
	return func() {}
}

// refCell returns the bag that holds name in this frame, moving the variable
// into one the first time it is captured by reference.
func (s *Scope) refCell(name string) map[string]any {
	if bag, ok := s.statics[name]; ok {
		return bag
	}
	bag := map[string]any{name: s.vars[name]}
	delete(s.vars, name)
	s.bindStatic(name, bag)
	return bag
}
