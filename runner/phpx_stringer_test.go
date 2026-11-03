//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxStringOf reads the string form echo writes: a scalar as phpString
// has it and an object with no __toString as php's Error.
func TestPhpxStringOf(t *testing.T) {
	if got, err := phpxStringOf(int64(7)); err != nil || got != "7" {
		t.Fatalf("int: %q, %v", got, err)
	}
	bare := &model.Object{Class: &model.Class{Name: "B"}}
	if _, err := phpxStringOf(bare); err == nil {
		t.Fatal("an object with no __toString converted to a string")
	}
	c := &model.Class{Name: "A", Stringer: func(*model.Object) (string, error) { return "a", nil }}
	if got, err := phpxStringOf(&model.Object{Class: c}); err != nil || got != "a" {
		t.Fatalf("stringer: %q, %v", got, err)
	}
}
