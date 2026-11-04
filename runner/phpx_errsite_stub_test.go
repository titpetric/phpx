//go:build !phpx

package runner

import (
	"errors"
	"testing"
)

// TestPhpxSiteUntagged pins that the untagged build records nothing.
func TestPhpxSiteUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	err := errors.New("x")
	rt.phpxNoteSite(err, nil)
	if _, ok := rt.phpxSiteMethod(err, "getLine"); ok {
		t.Error("untagged build answered a site")
	}
}
