//go:build !phpx

package compat

import "testing"

func TestPhpxBadUTF8StubAcceptsAll(t *testing.T) {
	if newRegexpCache().phpxBadUTF8("/./u", "\xe2") {
		t.Fatal("the untagged build refuses nothing")
	}
}
