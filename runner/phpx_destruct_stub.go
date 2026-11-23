//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// The untagged build answers upstream's lifecycle: a class's `__destruct` is
// an ordinary method nothing calls, which is the design decision the fork
// reverses. Every hook on the path is a no-op here; see phpx_destruct.go.

// phpxDtorState is absent in the untagged build; the field needs a type.
type phpxDtorState struct {
	live map[*model.Object]struct{}
}

// phpxTrackInstance answers the untagged build, which tracks no instance.
func (rt *Runtime) phpxTrackInstance(_ *model.Object) {}

// phpxNoteDropped answers the untagged build, which destroys nothing.
func (rt *Runtime) phpxNoteDropped(_ any) {}

// phpxSetVar answers the untagged build, which stores through Scope.Set alone.
func (rt *Runtime) phpxSetVar(scope *Scope, name string, val any) { scope.Set(name, val) }

// phpxFramePopped answers the untagged build, which has no frame to check.
func (rt *Runtime) phpxFramePopped(_ *Scope) {}

// phpxHasPending answers the untagged build, which tracks nothing.
func (rt *Runtime) phpxHasPending() bool { return false }

// phpxFlush answers the untagged build, which has nothing pending.
func (rt *Runtime) phpxFlush(_ any) error { return nil }

// phpxDestroyLive answers the untagged build, which destroys nothing at exit.
func (rt *Runtime) phpxDestroyLive() {}
