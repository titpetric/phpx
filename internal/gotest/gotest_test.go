package gotest

import (
	"testing"
)

// TestRun covers the helper on a script that needs no namespace.
func TestRun(t *testing.T) {
	Run(t, "none", []Case{{Src: `echo 1 + 1;`, Want: "2"}})
}
