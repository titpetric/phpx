//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.PhpxRegisterLate(registerFileGetContents)
}

// registerFileGetContents lets file_get_contents take its offset and length:
// the stream is read whole by the function it wraps and cut to the window.
func registerFileGetContents(rt *runner.Runtime) {
	read, ok := rt.Callable("file_get_contents")
	if !ok {
		return
	}
	rt.RegisterFunc("file_get_contents", func(name string, rest ...any) (any, error) {
		content, err := read(name)
		if err != nil || len(rest) < 3 {
			return content, err
		}
		s, isString := content.(string)
		if !isString {
			return content, nil
		}
		offset := phpval.Int(rest[2])
		if offset < 0 {
			offset = max(0, int64(len(s))+offset)
		}
		if offset > int64(len(s)) {
			return "", nil
		}
		s = s[offset:]
		if len(rest) > 3 && rest[3] != nil {
			if n := phpval.Int(rest[3]); n >= 0 && n < int64(len(s)) {
				s = s[:n]
			}
		}
		return s, nil
	})
}
