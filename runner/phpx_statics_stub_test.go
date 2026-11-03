//go:build !phpx

package runner

import (
	"testing"
)

// TestPhpxStaticStorageForUntagged pins upstream's answer: the property name
// takes no part in finding the bag, so each class keeps a copy.
func TestPhpxStaticStorageForUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	if _, err := rt.staticStorageFor("", "n", NewScope()); err == nil {
		t.Errorf("staticStorageFor on an empty class name must fail like staticStorage")
	}
}
