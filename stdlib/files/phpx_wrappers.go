package files

import (
	"io"
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

// phpxStreamContents is stream_get_contents with its $length and $offset, which
// upstream ignores: sebastian/diff writes php://memory and reads it back from 0.
func phpxStreamContents(stream io.Reader, rest []any) (string, error) {
	length, offset := int64(-1), int64(-1)
	if len(rest) > 0 {
		if n, ok := rest[0].(int64); ok {
			length = n
		}
	}
	if len(rest) > 1 {
		if n, ok := rest[1].(int64); ok {
			offset = n
		}
	}
	if seeker, ok := stream.(io.Seeker); ok && offset >= 0 {
		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return "", err
		}
	}
	if length >= 0 {
		stream = io.LimitReader(stream, length)
	}
	b, err := io.ReadAll(stream)
	return string(b), err
}
