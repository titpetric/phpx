//go:build !phpx

package parser

import (
	"testing"
)

// TestAltSyntaxUntagged pins that the untagged build reads no alternative syntax.
func TestAltSyntaxUntagged(t *testing.T) {
	t.Parallel()
	if _, ok, _ := (&parser{}).phpxAltBody(); ok {
		t.Error("read a body")
	}
}
