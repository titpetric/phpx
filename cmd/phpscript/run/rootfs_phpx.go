//go:build phpx

package run

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// phpxRunRoot roots a run at the host filesystem, as the php CLI does. A
// script includes an absolute path (/tmp/Mockery1234) and a path relative to
// the working directory, and both read the file php reads. Upstream keeps an
// include inside the working directory on purpose; the fork carries php's
// semantics because a vendor tree writes to /tmp and includes what it wrote.
//
// The answers are the source filesystem, the working directory inside it and
// the root a profile's paths are read from. The last answer is false for a
// build that keeps the sandbox.
func phpxRunRoot(string) (fs.FS, string, string, bool) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, "", "", false
	}
	return os.DirFS("/"), strings.TrimPrefix(filepath.ToSlash(wd), "/"), "/", true
}

// phpxEntrypoint spells the named script as an absolute path, which the host
// filesystem reads whichever directory the shell ran from.
func phpxEntrypoint(arg string) string {
	abs, err := filepath.Abs(arg)
	if err != nil {
		return arg
	}
	return filepath.ToSlash(abs)
}
