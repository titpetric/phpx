//go:build !phpx

package parser

import "testing"

// TestPhpxResolveTypeHintUntagged pins that an untagged build keeps the spelling.
func TestPhpxResolveTypeHintUntagged(t *testing.T) {
	t.Parallel()
	p := &parser{namespace: "Foo"}
	if got := p.phpxResolveTypeHint("?Bar"); got != "?Bar" {
		t.Errorf("got %q", got)
	}
}
