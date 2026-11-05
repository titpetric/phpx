//go:build phpx

package runner

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxGuardClears pins that a guard marks one accessor and clears with it.
func TestPhpxGuardClears(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	obj := model.NewObject(&model.Class{Name: "A"})
	done := rt.phpxGuard(obj, "x", "get")
	if !rt.phpxGuarded(obj, "x", "get") || rt.phpxGuarded(obj, "x", "set") || rt.phpxGuarded(obj, "y", "get") {
		t.Error("guard marks the wrong accessor")
	}
	done()
	if rt.phpxGuarded(obj, "x", "get") {
		t.Error("guard did not clear")
	}
	rt.phpxQuietLeave()
	if rt.phpxQuiet != 0 {
		t.Error("quiet went negative")
	}
}

func TestPhpxInQuietTracksIssetDepth(t *testing.T) {
	rt := New(io.Discard, Options{})
	if rt.phpxInQuiet() {
		t.Fatal("quiet before isset")
	}
	rt.phpxQuietEnter(false)
	if !rt.phpxInQuiet() {
		t.Fatal("not quiet inside isset")
	}
	rt.phpxQuietLeave()
	if rt.phpxInQuiet() {
		t.Fatal("quiet after isset")
	}
}
