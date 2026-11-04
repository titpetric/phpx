//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxCaptureRef notes the frame a by-reference `use` variable lives in.
func phpxCaptureRef(refs map[string]*Scope, use model.ClosureUse, scope *Scope) map[string]*Scope {
	if !use.ByRef {
		return refs
	}
	if refs == nil {
		refs = map[string]*Scope{}
	}
	refs[use.Name] = scope
	return refs
}

// phpxSyncRefs makes `use (&$x)` write back. The closure sees the enclosing
// frame's current value when it starts, and the value it leaves is stored there
// when it ends, if that frame is still running: a frame the pool took back
// belongs to someone else.
func (rt *Runtime) phpxSyncRefs(env closureEnv, scope *Scope) func() {
	if len(env.refs) == 0 {
		return func() {}
	}
	live := func(outer *Scope) bool {
		for _, f := range rt.frames {
			if f == outer {
				return true
			}
		}
		return false
	}
	for name, outer := range env.refs {
		if live(outer) {
			if v, ok := outer.Get(name); ok {
				scope.Set(name, v)
			}
		}
	}
	return func() {
		for name, outer := range env.refs {
			if live(outer) {
				if v, ok := scope.Get(name); ok {
					outer.Set(name, v)
				}
			}
		}
	}
}
