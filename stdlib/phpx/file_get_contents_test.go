//go:build phpx

package phpx

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestFileGetContentsWindow(t *testing.T) {
	rt := runner.New(io.Discard, runner.Options{})
	rt.RegisterFunc("file_get_contents", func(name string) (any, error) { return "0123456789", nil })
	registerFileGetContents(rt)
	read, _ := rt.Callable("file_get_contents")
	cases := []struct {
		args []any
		want string
	}{
		{[]any{"f"}, "0123456789"},
		{[]any{"f", false, nil, int64(3)}, "3456789"},
		{[]any{"f", false, nil, int64(3), int64(2)}, "34"},
		{[]any{"f", false, nil, int64(-2)}, "89"},
		{[]any{"f", false, nil, int64(99)}, ""},
	}
	for _, c := range cases {
		got, err := read(c.args...)
		if err != nil || got != c.want {
			t.Errorf("%v = %q, %v; want %q", c.args, got, err, c.want)
		}
	}
}
