//go:build !phpx

package run

import (
	"io/fs"
)

// phpxRunRoot keeps the working-directory sandbox: no answer, so the caller
// leaves the filesystem options alone.
func phpxRunRoot(string) (fs.FS, string, string, bool) { return nil, "", "", false }

// phpxEntrypoint keeps the script spelling the entrypoint resolution chose.
func phpxEntrypoint(arg string) string { return arg }
