//go:build phpx

package runner

import "testing"

func TestPhpxInstallVarVarAddsHelpers(t *testing.T) {
	rt := New(nil, Options{})
	env := map[string]any{}
	rt.phpxInstallVarVar(env, &scopeRef{scope: rt.newScope()})
	if env["__phpx_varvar"] == nil || env["__phpx_varvar_set"] == nil {
		t.Fatal("helpers missing")
	}
}
