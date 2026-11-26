//go:build !phpx

package runner

import (
	"io/fs"
	"testing"
)

// TestPhpxIncludeAnswersFalseStub pins the untagged build, where an include that fails ends the program.
func TestPhpxIncludeAnswersFalseStub(t *testing.T) {
	if phpxIncludeAnswersFalse("include", fs.ErrNotExist) {
		t.Error("untagged build answers false for a failed include")
	}
}
