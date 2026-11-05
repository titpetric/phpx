//go:build !phpx

package runner

import "testing"

func TestPhpxAliasMatchesStubIsOff(t *testing.T) {
	if phpxAliasMatches("a", "a") {
		t.Fatal("the untagged build knows no aliases")
	}
}
