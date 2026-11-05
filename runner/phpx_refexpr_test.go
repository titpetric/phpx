//go:build phpx

package runner

import "testing"

// TestPhpxRefExprInstalls pins the helper name the expression compiler emits.
func TestPhpxRefExprInstalls(t *testing.T) {
	t.Parallel()
	env := map[string]any{}
	(&Runtime{}).phpxInstallRefExpr(env, nil)
	if _, ok := env["__refexpr"]; !ok {
		t.Error("__refexpr missing")
	}
}

// TestPhpxHandleID pins that a Go pointer has an identity and an object, a
// scalar and nil have none from here.
func TestPhpxHandleID(t *testing.T) {
	t.Parallel()
	p := &struct{ x int }{}
	if a, ok := phpxHandleID(p); !ok || a == 0 {
		t.Error("pointer has no identity")
	}
	if _, ok := phpxHandleID(3); ok {
		t.Error("scalar has an identity")
	}
	if _, ok := phpxHandleID(nil); ok {
		t.Error("nil has an identity")
	}
}
