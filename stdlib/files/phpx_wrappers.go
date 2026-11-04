package files

import (
	"os"
	"strings"
)

// phpxWrapper opens the php:// stream scheme names. stdout and stderr are the
// process's own streams. memory and temp stay unopened, as upstream pins. It reports false for a
// scheme it does not hold.
func phpxWrapper(scheme string) (any, bool) {
	switch strings.ToLower(scheme) {
	case "stdout":
		return os.Stdout, true
	case "stderr":
		return os.Stderr, true
	}
	return nil, false
}
