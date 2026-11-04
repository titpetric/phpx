package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// A binary named phpx-test is this program built with go build -cover. Run as
// `phpx test ...`, phpx hands the command to it when it finds one. The coverage
// runtime reads GOCOVERDIR as the process starts, before any init runs, so both
// sides set it and exec: phpx before handing over, and phpx-test when started
// bare. The directory is cover/phpx-test unless GOCOVERDIR names another.
func init() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	switch name {
	case "phpx-test":
		if os.Getenv("GOCOVERDIR") == "" {
			if self, err := os.Executable(); err == nil {
				_ = syscall.Exec(self, os.Args, phpxCoverEnv())
			}
		}
	case "phpx":
		if path := phpxTestBinary(os.Args[1:]); path != "" {
			_ = syscall.Exec(path, append([]string{path}, os.Args[1:]...), phpxCoverEnv())
		}
	}
}

// phpxCoverEnv is the environment with GOCOVERDIR set and its directory made.
func phpxCoverEnv() []string {
	env := os.Environ()
	if os.Getenv("GOCOVERDIR") != "" {
		return env
	}
	dir := filepath.Join("cover", "phpx-test")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return env
	}
	return append(env, "GOCOVERDIR="+dir)
}

// phpxTestBinary finds phpx-test, beside this executable first and then on PATH,
// when the arguments are the test command. It answers "" for any other command.
func phpxTestBinary(args []string) string {
	command := ""
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			command = a
			break
		}
	}
	if command != "test" {
		return ""
	}
	if self, err := os.Executable(); err == nil {
		if candidate := filepath.Join(filepath.Dir(self), "phpx-test"); isExecutable(candidate) {
			return candidate
		}
	}
	if path, err := exec.LookPath("phpx-test"); err == nil {
		return path
	}
	return ""
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}
