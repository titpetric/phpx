//go:build !phpx

package runner

import "testing"

// TestPhpxDivZeroUntagged pins that an untagged build raises nothing.
func TestPhpxDivZeroUntagged(t *testing.T) {
	t.Parallel()
	if err := phpxDivZero("/", int64(1), int64(0)); err != nil {
		t.Errorf("got %v", err)
	}
}
