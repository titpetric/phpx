//go:build phpx

package runner

import (
	"errors"
	"io/fs"
	"strings"
)

// phpxIncludeAnswersFalse reports an include keyword that answers false on a read failure.
func phpxIncludeAnswersFalse(keyword string, err error) bool {
	if !strings.HasPrefix(strings.ToLower(keyword), "include") {
		return false
	}
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission)
}

// phpxIncludeWarning emits php's warning for an include that did not open.
func phpxIncludeWarning(rt *Runtime, keyword, path string, err error) {

	reason := "No such file or directory"
	if errors.Is(err, fs.ErrPermission) {
		reason = "Permission denied"
	}
	rt.PhpxWarning(2, keyword+"("+path+"): Failed to open stream: "+reason)
}
