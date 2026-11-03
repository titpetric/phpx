//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxStringOfUntagged pins upstream's answer: an object with no string
// form echoes as the empty string.
func TestPhpxStringOfUntagged(t *testing.T) {
	got, err := phpxStringOf(&model.Object{Class: &model.Class{Name: "B"}})
	if err != nil || got != "" {
		t.Fatalf("got %q, %v", got, err)
	}
}
