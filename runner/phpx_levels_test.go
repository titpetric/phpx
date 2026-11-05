//go:build phpx

package runner

import "testing"

// TestPhpxLeaveCountsLevels pins that `break 3` passes through two loops.
func TestPhpxLeaveCountsLevels(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	rt.phpxOuterLevels(3)
	if !rt.phpxLeave(flowBreak) || !rt.phpxLeave(flowBreak) {
		t.Fatal("did not pass the first two levels on")
	}
	if rt.phpxLeave(flowBreak) {
		t.Error("passed the last level on")
	}
	rt.phpxOuterLevels(1)
	if rt.phpxLeave(flowContinue) {
		t.Error("a plain continue left its loop")
	}
	rt.phpxOuterLevels(2)
	if rt.phpxLeave(flowReturn) {
		t.Error("a return was counted as a level")
	}
}
