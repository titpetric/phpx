package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// A binary named phpx-test is this program built with go build -cover. phpx
// hands `phpx test` to it through upstream's <binary>-<tool> discovery. The
// coverage runtime reads GOCOVERDIR as the process starts, before any init runs,
// so a phpx-test started without it sets it and execs itself. It also takes
// `test` as its command when the arguments do not begin with it. The directory is
// cover/phpx-test unless GOCOVERDIR names another.
func init() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	if name != "phpx-test" {
		return
	}
	// A discovered tool is handed its arguments without its own name, and
	// phpx-test is the test command: `phpx test --matrix x` arrives as `--matrix x`.
	if len(os.Args) < 2 || os.Args[1] != "test" {
		os.Args = append([]string{os.Args[0], "test"}, os.Args[1:]...)
	}
	if os.Getenv("GOCOVERDIR") != "" {
		return
	}
	if self, err := os.Executable(); err == nil {
		_ = syscall.Exec(self, os.Args, phpxCoverEnv())
	}
}

// phpxCoverEnv is the environment with GOCOVERDIR set and its directory made.
func phpxCoverEnv() []string {
	env := os.Environ()
	dir := filepath.Join("cover", "phpx-test")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return env
	}
	return append(env, "GOCOVERDIR="+dir)
}
