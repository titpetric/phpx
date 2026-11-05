//go:build !phpx

package parser

import "testing"

// TestBracedNamespaceUntagged pins that an untagged build refuses the form.
func TestBracedNamespaceUntagged(t *testing.T) {
	t.Parallel()
	if _, err := Parse("<?php namespace A { class X {} }"); err == nil {
		t.Error("untagged build parsed a braced namespace")
	}
}
