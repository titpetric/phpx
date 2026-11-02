//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// newGlobalFrames returns a runtime whose global frame holds vars, and a fresh
// local frame to bind names into.
func newGlobalFrames(vars map[string]any) (*Runtime, *Scope, *Scope) {
	globals := NewScope()
	for name, value := range vars {
		globals.vars[name] = value
	}
	rt := &Runtime{frames: []*Scope{globals}}
	return rt, globals, NewScope()
}

func TestPhpxGlobalReadsAndWritesTheGlobalFrame(t *testing.T) {
	t.Parallel()
	rt, globals, local := newGlobalFrames(map[string]any{"counter": 1})

	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}

	got, ok := local.Get("counter")
	if !ok || got != 1 {
		t.Errorf("local read = %v, %v, want 1, true", got, ok)
	}

	local.Set("counter", 11)
	if got := globals.vars["counter"]; got != 11 {
		t.Errorf("global after a local write = %v, want 11", got)
	}
}

func TestPhpxGlobalCreatesAnUnsetNameAsNull(t *testing.T) {
	t.Parallel()
	rt, globals, local := newGlobalFrames(nil)

	if err := rt.phpxGlobal(&model.Global{Names: []string{"fresh"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}

	value, ok := globals.vars["fresh"]
	if !ok || value != nil {
		t.Errorf("global entry = %v, %v, want nil, true", value, ok)
	}
	if got, ok := local.Get("fresh"); !ok || got != nil {
		t.Errorf("local read = %v, %v, want nil, true", got, ok)
	}
}

func TestPhpxGlobalUnsetDropsTheBindingAndKeepsTheGlobal(t *testing.T) {
	t.Parallel()
	rt, globals, local := newGlobalFrames(map[string]any{"counter": 11})

	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}
	local.Unset("counter")

	if _, ok := local.Get("counter"); ok {
		t.Error("the local name stayed bound after unset")
	}
	if got := globals.vars["counter"]; got != 11 {
		t.Errorf("global after unset = %v, want 11", got)
	}
}

func TestPhpxGlobalLeavesALocalValueOutOfReach(t *testing.T) {
	t.Parallel()
	rt, globals, local := newGlobalFrames(map[string]any{"shadow": "global value"})
	local.Set("shadow", "local value")

	if err := rt.phpxGlobal(&model.Global{Names: []string{"shadow"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}

	if got, _ := local.Get("shadow"); got != "global value" {
		t.Errorf("local read = %v, want the global value", got)
	}
	if got := globals.vars["shadow"]; got != "global value" {
		t.Errorf("global = %v, want it untouched", got)
	}
}

func TestPhpxGlobalBindsEveryNameAndSkipsTheFileBody(t *testing.T) {
	t.Parallel()
	rt, globals, local := newGlobalFrames(map[string]any{"counter": 11, "name": "top"})

	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter", "name"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}
	for name, want := range map[string]any{"counter": 11, "name": "top"} {
		if got, ok := local.Get(name); !ok || got != want {
			t.Errorf("local read of %s = %v, %v, want %v, true", name, got, ok, want)
		}
	}

	// The file body's own frame names the global already, and binding it to
	// itself would be the same map twice.
	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter"}}, globals); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}
	if globals.statics != nil {
		t.Error("the global frame grew a binding to itself")
	}
}

func TestPhpxGlobalWithoutAFrameDoesNothing(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	local := NewScope()

	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}
	if _, ok := local.Get("counter"); ok {
		t.Error("a name bound with no global frame to bind it to")
	}
}
