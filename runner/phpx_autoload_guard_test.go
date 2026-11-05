//go:build phpx

package runner

import (
	"io"
	"testing"
)

func TestPhpxEnterAutoloadRefusesReentry(t *testing.T) {
	rt := New(io.Discard, Options{})
	release, ok := phpxEnterAutoload(rt, "Vendor\\Thing")
	if !ok {
		t.Fatal("first entry refused")
	}
	if _, again := phpxEnterAutoload(rt, "vendor\\thing"); again {
		t.Fatal("a class being loaded was entered again")
	}
	release()
	if release2, again := phpxEnterAutoload(rt, "Vendor\\Thing"); !again {
		t.Fatal("entry after release refused")
	} else {
		release2()
	}
}
