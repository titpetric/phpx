//go:build phpx

package runner

import "testing"

type hostFieldFixture struct {
	Y int
	F float64
	D any
	t int
}

func TestPhpxHostFields(t *testing.T) {
	h := &hostFieldFixture{Y: 2, D: false}
	if v, ok := phpxHostField(h, "y"); !ok || v != int64(2) {
		t.Fatalf("y = %v %v", v, ok)
	}
	if !phpxHostFieldSet(h, "f", "1.5") || h.F != 1.5 {
		t.Fatalf("f = %v", h.F)
	}
	if !phpxHostFieldSet(h, "d", int64(7)) || h.D != int64(7) {
		t.Fatalf("d = %v", h.D)
	}
	if _, ok := phpxHostField(h, "t"); ok {
		t.Fatal("an unexported field is no property")
	}
	if phpxHostFieldSet(h, "nope", 1) {
		t.Fatal("an unknown property was set")
	}
}
