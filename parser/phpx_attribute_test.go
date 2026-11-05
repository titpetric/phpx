//go:build phpx

package parser

import "testing"

// TestParameterAttributes pins that an attribute inside a parameter list ends
// at its bracket and leaves the parameter to the parser.
func TestParameterAttributes(t *testing.T) {
	t.Parallel()
	for _, src := range []string{
		"<?php function f(#[A] $x, #[B('a]b')] int $y = 2) { return $x; }",
		"<?php class C { public function __construct(#[SensitiveParameter] private string $s) {} }",
		"<?php #[A] #[B] function g() {}",
	} {
		if _, err := Parse(src); err != nil {
			t.Errorf("%q: %v", src, err)
		}
	}
}
