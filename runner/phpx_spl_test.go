//go:build phpx

package runner

import (
	"io"
	"testing"
)

func TestPhpxCloneGoStruct(t *testing.T) {
	type host struct{ N int }
	rt := New(io.Discard, Options{})
	src := &host{N: 1}
	got, err := rt.PhpxClone(src)
	if err != nil {
		t.Fatal(err)
	}
	dup := got.(*host)
	dup.N = 2
	if src.N != 1 || dup == src {
		t.Fatal("clone of a Go struct shared its storage")
	}
	if _, err := rt.PhpxClone(5); err == nil {
		t.Fatal("clone of a scalar should fail")
	}
}
