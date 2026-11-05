//go:build !phpx

package runner

import (
	"testing"
)

// TestPhpxRefExprUntagged pins that the untagged build installs no helper.
func TestPhpxRefExprUntagged(t *testing.T) {
	t.Parallel()
	env := map[string]any{}
	(&Runtime{}).phpxInstallRefExpr(env, nil)
	if len(env) != 0 {
		t.Error("untagged build installed a helper")
	}
}

// TestPhpxHandleIDUntagged pins that the untagged build gives no identity.
func TestPhpxHandleIDUntagged(t *testing.T) {
	t.Parallel()
	if _, ok := phpxHandleID(&struct{}{}); ok {
		t.Error("untagged build gave an identity")
	}
}
