//go:build phpx

package phpx

import "testing"

// TestPhpxErrorHandlerMask pins the default mask of a handler registration.
func TestPhpxErrorHandlerMask(t *testing.T) {
	t.Parallel()
	h := phpxErrorHandler{levels: 32767}
	if h.levels&512 == 0 {
		t.Error("E_ALL does not include E_USER_WARNING")
	}
}
