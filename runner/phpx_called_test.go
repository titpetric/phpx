package runner

import "testing"

func TestPhpxForwardedClass(t *testing.T) {
	scope := NewScope()
	if got := phpxForwardedClass("self", "A", scope); got != "A" {
		t.Fatalf("no called class recorded: got %q, want A", got)
	}
	phpxSetCalled(scope, "B")
	for keyword, want := range map[string]string{"self": "B", "parent": "B", "static": "B", "A": "A"} {
		if got := phpxForwardedClass(keyword, "A", scope); got != want {
			t.Errorf("%s: got %q, want %q", keyword, got, want)
		}
	}
	if name, err := resolveCalledClass("static", scope); err != nil || name != "B" {
		t.Errorf("static resolved to %q, %v", name, err)
	}
}
