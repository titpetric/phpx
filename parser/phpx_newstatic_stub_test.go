//go:build !phpx

package parser

import (
	"testing"
)

// TestNewStaticPropertyUntagged pins that the untagged parser refuses the form.
func TestNewStaticPropertyUntagged(t *testing.T) {
	t.Parallel()
	if _, ok, err := (&parser{}).phpxNewStaticProp("A"); ok || err != nil {
		t.Error("untagged parser read a static property after new")
	}
}
