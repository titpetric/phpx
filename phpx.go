package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// A tool binary is named <binary>-<tool>: phpx-test is phpx built with go build
// -cover. Upstream's discovery hands it its arguments without its own name, so
// `phpx test --matrix x` arrives as `--matrix x`. The name carries the command:
// everything after the first dash is it, and it goes in as the first argument
// before anything reads os.Args.
//
// The coverage runtime reads GOCOVERDIR as the process starts, before any init
// runs, so a phpx-test started without it sets it and execs itself. The
// directory is cover/phpx-test unless GOCOVERDIR names another.
func init() {
	os.Args = phpxToolArgs(os.Args)
	name := phpxBaseName(os.Args[0])
	if name != "phpx-test" || os.Getenv("GOCOVERDIR") != "" {
		return
	}
	if self, err := os.Executable(); err == nil {
		_ = syscall.Exec(self, os.Args, phpxCoverEnv())
	}
}

// phpxBaseName is the file name of argv0 without its extension.
func phpxBaseName(argv0 string) string {
	base := filepath.Base(argv0)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// phpxToolArgs puts the command named by a <binary>-<tool> argv[0] in front of
// the arguments: phpx-test --matrix x becomes phpx-test test --matrix x. A
// name with no dash, or arguments that already begin with the command, stay as
// they are.
func phpxToolArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	_, tool, found := strings.Cut(phpxBaseName(args[0]), "-")
	// phpx-linux-amd64 is a release file name and no tool.
	if !found || tool == "" || strings.Contains(tool, "-") {
		return args
	}
	if len(args) > 1 && args[1] == tool {
		return args
	}
	return append([]string{args[0], tool}, args[1:]...)
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
