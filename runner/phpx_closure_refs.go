//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxCaptureRef makes the variable of a by-reference `use` a shared cell. The
// enclosing frame and every closure that captures the name read and write the
// one cell, so a value assigned after the closure was made, or by the closure
// after the frame ended, is the one both see.
func phpxCaptureRef(refs map[string]*model.Cell, use model.ClosureUse, scope *Scope) map[string]*model.Cell {
	if !use.ByRef {
		return refs
	}
	if refs == nil {
		refs = map[string]*model.Cell{}
	}
	refs[use.Name] = scope.refCell(use.Name)
	return refs
}

// phpxSyncRefs binds the closure frame's by-reference names to their cells.
func (rt *Runtime) phpxSyncRefs(env closureEnv, scope *Scope) func() {
	for name, cell := range env.refs {
		scope.bindRef(name, cell)
	}
	return func() {}
}

// refCell returns the cell that holds name in this frame, moving the variable
// into one the first time it is shared.
func (s *Scope) refCell(name string) *model.Cell {
	if cell, ok := s.refs[name]; ok {
		return cell
	}
	var cell *model.Cell
	if bag, ok := s.statics[name]; ok {
		cell = &model.Cell{V: bag[name], Mirror: func(v any) { bag[name] = v }}
		delete(s.statics, name)
	} else {
		cell = &model.Cell{V: s.vars[name]}
		delete(s.vars, name)
	}
	s.bindRef(name, cell)
	return cell
}

// bindRef makes name the cell, for the rest of the frame.
func (s *Scope) bindRef(name string, cell *model.Cell) {
	if s.refs == nil {
		s.refs = map[string]*model.Cell{}
	}
	s.refs[name] = cell
}
