package phpval

import (
	"testing"
)

func TestIsNumericString(t *testing.T) {
	for s, want := range map[string]bool{"12.5": true, "1e3": true, " 5": true, "5 ": true, "abc": false, "0x1A": false, "": false, ".5": true} {
		if got := IsNumericString(s); got != want {
			t.Errorf("IsNumericString(%q) = %v, want %v", s, got, want)
		}
	}
}
