package files

import (
	"io"
	"os"
	"strings"

	"github.com/titpetric/phpx/model"
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

// phpxPathinfo is pathinfo() with its $flags, which upstream does not take.
func phpxPathinfo(path string, flags ...int64) any {
	parts := model.NewArray()
	// A trailing slash names the directory: pathinfo('dir/') is 'dir'.
	if trimmed := strings.TrimRight(path, "/"); trimmed != "" {
		path = trimmed
	}
	var dir, base, ext, name string
	hasDir := strings.Contains(path, "/")
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		dir, base = path[:i], path[i+1:]
		if dir == "" {
			dir = "/"
		}
	} else {
		base = path
	}
	hasExt := false
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		ext, name, hasExt = base[i+1:], base[:i], true
	} else {
		name = base
	}
	mask := int64(15)
	if len(flags) > 0 {
		mask = flags[0]
	}
	if mask == 15 {
		if hasDir {
			parts.Set("dirname", dir)
		} else {
			parts.Set("dirname", ".")
		}
		parts.Set("basename", base)
		if hasExt {
			parts.Set("extension", ext)
		}
		parts.Set("filename", name)
		return parts
	}
	switch mask {
	case 1:
		if hasDir {
			return dir
		}
		return "."
	case 2:
		return base
	case 4:
		return ext
	case 8:
		return name
	}
	return ""
}
