//go:build !phpx

package runner

import "testing"

func TestPhpxEnterAutoloadStubAllows(t *testing.T) {
	if _, ok := phpxEnterAutoload(nil, "X"); !ok {
		t.Fatal("the untagged build refuses nothing")
	}
}
