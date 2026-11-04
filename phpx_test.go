package main

import "testing"

func TestPhpxTestBinaryOnlyForTheTestCommand(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if got := phpxTestBinary([]string{"run", "x.php"}); got != "" {
		t.Fatalf("got %q for run", got)
	}
	if got := phpxTestBinary([]string{"--matrix", "test"}); got != "" {
		t.Fatalf("got %q with no phpx-test on the path", got)
	}
}
