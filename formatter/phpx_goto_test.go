package formatter

import "testing"

// TestPhpxLoopJump pins the spelling of break and continue with a level.
func TestPhpxLoopJump(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{"break": "break;", "continue": "continue;"} {
		if got := phpxLoopJump(in, 0); got != want {
			t.Errorf("%s = %q", in, got)
		}
	}
	if got := phpxLoopJump("break", 3); got != "break 3;" {
		t.Errorf("break 3 = %q", got)
	}
}
