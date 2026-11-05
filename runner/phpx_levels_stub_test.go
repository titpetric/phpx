//go:build !phpx

package runner

import "testing"

// TestPhpxLevelsUntagged pins that the untagged build passes nothing outward.
func TestPhpxLevelsUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	rt.phpxOuterLevels(3)
	if rt.phpxLeave(flowBreak) {
		t.Error("untagged build left a level")
	}
}
