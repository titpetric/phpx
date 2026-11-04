package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestPhpxToolArgs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want []string
	}{
		{[]string{"/usr/local/bin/phpx-test", "--matrix", "x"}, []string{"/usr/local/bin/phpx-test", "test", "--matrix", "x"}},
		{[]string{"phpx-test", "test", "--matrix"}, []string{"phpx-test", "test", "--matrix"}},
		{[]string{"phpx-test"}, []string{"phpx-test", "test"}},
		{[]string{"phpx-gen", "a"}, []string{"phpx-gen", "gen", "a"}},
		{[]string{"phpx", "run", "x.php"}, []string{"phpx", "run", "x.php"}},
		{[]string{"phpx-linux-amd64", "run"}, []string{"phpx-linux-amd64", "run"}},
		{[]string{"phpscript", "test"}, []string{"phpscript", "test"}},
	}
	for _, tc := range cases {
		if got := phpxToolArgs(tc.in); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("phpxToolArgs(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPhpxCoverEnvSetsGOCOVERDIR(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, kv := range phpxCoverEnv() {
		if strings.HasPrefix(kv, "GOCOVERDIR=") && strings.HasSuffix(kv, "phpx-test") {
			return
		}
	}
	t.Fatal("GOCOVERDIR is not set to cover/phpx-test")
}
