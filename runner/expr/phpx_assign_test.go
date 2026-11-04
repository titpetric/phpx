package expr

import (
	"testing"
)

func TestPhpxOrderNamedKeepsAnUnknownCallee(t *testing.T) {
	t.Parallel()
	if got := phpxOrderNamed("not_a_builtin", nil); got != nil {
		t.Fatalf("got %v, want the arguments unchanged", got)
	}
}
