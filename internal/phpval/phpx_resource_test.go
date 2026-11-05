package phpval

import "testing"

func TestResourceHook(t *testing.T) {
	old := PhpxResourceID
	defer func() { PhpxResourceID = old }()
	type handle struct{}
	PhpxResourceID = func(v any) (int64, bool) {
		if _, ok := v.(*handle); ok {
			return 7, true
		}
		return 0, false
	}
	if got := String(&handle{}); got != "Resource id #7" {
		t.Fatalf("string = %q", got)
	}
	if got := Int(&handle{}); got != 7 {
		t.Fatalf("int = %d", got)
	}
	PhpxResourceID = nil
	if String(&handle{}) != "" {
		t.Fatal("without the hook a handle has no string form")
	}
}
