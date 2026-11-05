//go:build !phpx

package compat

import "testing"

// TestPhpxBuffersUntagged pins that the untagged stub is callable.
func TestPhpxBuffersUntagged(t *testing.T) {
	t.Parallel()
	phpxRegisterBuffers(nil, nil)
}
