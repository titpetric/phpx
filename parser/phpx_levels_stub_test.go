//go:build !phpx

package parser

import (
	"testing"
)

// TestPhpxLoopLevelsUntagged pins that the untagged build reads no level.
func TestPhpxLoopLevelsUntagged(t *testing.T) {
	t.Parallel()
	if got := (&parser{}).phpxLoopLevels(); got != 0 {
		t.Errorf("got %d", got)
	}
}
