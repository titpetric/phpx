//go:build phpx

package runner

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"
)

// TestPhpxIncludeAnswersFalse pins php's failure answers: include and
// include_once warn and return false when the file cannot be read, while
// require stays fatal and a parse error still ends the program.
func TestPhpxIncludeAnswersFalse(t *testing.T) {
	missing := fmt.Errorf("include %q: %w", "x.php", fs.ErrNotExist)
	cases := []struct {
		keyword string
		err     error
		want    bool
	}{
		{"include", missing, true},
		{"include_once", missing, true},
		{"require", missing, false},
		{"require_once", missing, false},
		{"include", fs.ErrPermission, true},
		{"include", errors.New("parse error"), false},
	}
	for _, tc := range cases {
		if got := phpxIncludeAnswersFalse(tc.keyword, tc.err); got != tc.want {
			t.Errorf("phpxIncludeAnswersFalse(%q, %v) = %v, want %v", tc.keyword, tc.err, got, tc.want)
		}
	}
}
