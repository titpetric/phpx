//go:build !phpx

package parser

import "testing"

// TestPhpxSyntaxUntagged pins that the untagged build reads none of the forms.
func TestPhpxSyntaxUntagged(t *testing.T) {
	t.Parallel()
	p := &parser{}
	if p.phpxByRefName() {
		t.Error("read a by-reference name")
	}
	if _, ok, _ := p.phpxStaticDynamicMember("A"); ok {
		t.Error("read a dynamic static member")
	}
	if (&lexer{}).phpxLeadingDot() {
		t.Error("lexed a leading-dot float")
	}
}
