package files

import (
	"os"
	"strings"

	"github.com/titpetric/phpx/runner"
)

// phpxWrapper opens the php:// stream scheme names. stdout and stderr are the
// process's own streams, and memory and temp are an unlinked scratch file, so
// every stream function that reads an *os.File works on them. It reports false for a
// scheme it does not hold.
func phpxWrapper(_ *runner.Runtime, scheme string) (any, bool) {
	switch strings.ToLower(scheme) {
	case "stdout":
		// The process's own stdout, as php://stdout is: not the script's output
		// buffer, and not captured by a host that captures that.
		return os.Stdout, true
	case "stderr":
		return os.Stderr, true
	case "memory", "temp":
		f, err := os.CreateTemp("", "phpx-stream-")
		if err != nil {
			return false, true
		}
		_ = os.Remove(f.Name())
		return f, true
	}
	return nil, false
}
