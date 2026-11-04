package expr

import (
	"testing"
)

func TestPhpxOffsetCallWithoutHelper(t *testing.T) {
	t.Parallel()
	got, err := phpxOffsetCall(&Env{Base: map[string]any{}}, "get", nil, "k")
	if got != nil || err != nil {
		t.Fatalf("got %v, %v; an environment with no helper answers nothing", got, err)
	}
}
