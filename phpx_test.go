package main

import (
	"strings"
	"testing"
)

func TestPhpxCoverEnvSetsGOCOVERDIR(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, kv := range phpxCoverEnv() {
		if strings.HasPrefix(kv, "GOCOVERDIR=") && strings.HasSuffix(kv, "phpx-test") {
			return
		}
	}
	t.Fatal("GOCOVERDIR is not set to cover/phpx-test")
}
