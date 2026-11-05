//go:build phpx

package runner

import (
	"bytes"
	"testing"
)

func TestPhpxResources(t *testing.T) {
	a, b := &bytes.Buffer{}, &bytes.Buffer{}
	if !PhpxIsResource(a) || PhpxIsResource("x") || PhpxIsResource(nil) {
		t.Fatal("a buffer is a stream and a string is not")
	}
	ida, idb := PhpxResourceID(a), PhpxResourceID(b)
	if ida == idb || PhpxResourceID(a) != ida || ida < 5 {
		t.Fatalf("ids %d %d", ida, idb)
	}
	if PhpxResourceClosed(a) {
		t.Fatal("open handle reads closed")
	}
	PhpxCloseResource(a)
	if !PhpxResourceClosed(a) || PhpxResourceClosed(b) {
		t.Fatal("close should mark the one handle")
	}
	if phpxResourceString(a) == "" {
		t.Fatal("empty string form")
	}
}
