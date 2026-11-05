//go:build phpx

package phpx

import "testing"

func TestSettype(t *testing.T) {
	var got any
	set := func(v any) { got = v }
	if !settype(set, "12abc", "int") || got != int64(12) {
		t.Fatalf("int: %v", got)
	}
	if !settype(set, int64(0), "boolean") || got != false {
		t.Fatalf("bool: %v", got)
	}
	if !settype(set, 1, "null") || got != nil {
		t.Fatalf("null: %v", got)
	}
	if settype(set, 1, "resource") {
		t.Fatal("resource is no settype target")
	}
}
