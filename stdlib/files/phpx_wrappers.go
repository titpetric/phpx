package files

import (
	"io"
	"os"

	"github.com/titpetric/phpx/runner"
)

func init() {
	RegisterWrapper("stdout", func(*runner.Runtime, string) (any, error) { return os.Stdout, nil })
	RegisterWrapper("stderr", func(*runner.Runtime, string) (any, error) { return os.Stderr, nil })
	scratch := func(*runner.Runtime, string) (any, error) {
		f, err := os.CreateTemp("", "phpx-stream-")
		if err != nil {
			return false, nil
		}
		_ = os.Remove(f.Name())
		return f, nil
	}
	RegisterWrapper("memory", scratch)
	RegisterWrapper("temp", scratch)
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
