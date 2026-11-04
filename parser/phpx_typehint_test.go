//go:build phpx

package parser

import "testing"

// TestPhpxResolveTypeHint pins the names a hint reads as in a namespace with an import.
func TestPhpxResolveTypeHint(t *testing.T) {
	t.Parallel()
	p := &parser{namespace: "Foo", imports: map[string]string{"Other": "Vendor\\Other"}}
	cases := map[string]string{
		"":           "",
		"int":        "int",
		"?Bar":       "?Foo\\Bar",
		"Other|null": "?Vendor\\Other",
		"\\DateTime": "DateTime",
		"self":       "self",
		"Foo|null":   "?Foo\\Foo",
		"null|int":   "?int",
		"int|string": "int|string",
		"A&B":        "Foo\\A&Foo\\B",
	}
	for in, want := range cases {
		if got := p.phpxResolveTypeHint(in); got != want {
			t.Errorf("%q = %q, want %q", in, got, want)
		}
	}
}
