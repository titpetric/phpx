//go:build phpx

package runner

import (
	"errors"
	"fmt"
	"testing"
)

// valueErr is an error that is no pointer.
type valueErr string

func (e valueErr) Error() string { return string(e) }

// sliceErr holds a slice, so its dynamic type is not comparable and cannot key
// the site table.
type sliceErr struct{ parts []string }

func (e sliceErr) Error() string { return "slice" }

// TestPhpxSiteFirstWins pins that the first note of an error is kept, that a
// comparable value error is keyed, and that an error whose dynamic type holds a
// slice is left out without a panic.
func TestPhpxSiteFirstWins(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	err := &TypeError{}
	rt.phpxNoteSite(err, nil)
	rt.phpxSites[err] = phpxSite{file: "a.php", line: 3}
	rt.phpxNoteSite(err, nil)
	if got, _ := rt.phpxSiteMethod(err, "getLine"); got != int64(3) {
		t.Errorf("getLine = %v, want 3", got)
	}
	if got, _ := rt.phpxSiteMethod(err, "getFile"); got != "a.php" {
		t.Errorf("getFile = %v", got)
	}
	rt.phpxNoteSite(valueErr("v"), nil)
	if len(rt.phpxSites) != 2 {
		t.Fatalf("sites = %d, want 2", len(rt.phpxSites))
	}
	if got, _ := rt.phpxSiteMethod(valueErr("v"), "getLine"); got != int64(0) {
		t.Errorf("value error getLine = %v, want 0", got)
	}
	rt.phpxNoteSite(sliceErr{parts: []string{"a"}}, nil)
	if len(rt.phpxSites) != 2 {
		t.Fatalf("sites = %d after an uncomparable error, want 2", len(rt.phpxSites))
	}
}

// TestPhpxSiteUnwrapped pins that a wrapper and its cause both carry the site,
// because a catch clause binds the cause.
func TestPhpxSiteUnwrapped(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	cause := errors.New("broken pipe")
	wrapped := fmt.Errorf("load file: %w", cause)
	rt.phpxNoteSite(wrapped, nil)
	for _, err := range []error{wrapped, cause} {
		if _, ok := rt.phpxSiteMethod(err, "getLine"); !ok {
			t.Errorf("%v has no site", err)
		}
	}
}
