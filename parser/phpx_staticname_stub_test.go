//go:build !phpx

package parser

import (
	"testing"
)

// TestStaticPropNamedByExpressionUntagged pins that the untagged build refuses
// a static property named by an expression, as it always has.
func TestStaticPropNamedByExpressionUntagged(t *testing.T) {
	t.Parallel()
	if _, err := Parse("<?php $a = A::${$p};"); err == nil {
		t.Error("untagged parser read a static property named by an expression")
	}
}
