//go:build !phpx

package runner

import "testing"

func TestPhpxHooksStubAnswersNothing(t *testing.T) {
	rt := New(nil, Options{})
	if _, handled, _ := rt.phpxHookGet(nil, "x", nil); handled {
		t.Error("untagged build has no hook")
	}
	if handled, _ := rt.phpxHookSet(nil, "x", 1, nil); handled {
		t.Error("untagged build has no hook")
	}
}
