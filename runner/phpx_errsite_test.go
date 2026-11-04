//go:build phpx

package runner

import "testing"

// valueErr is an error that is no pointer.
type valueErr string

func (e valueErr) Error() string { return string(e) }

// TestPhpxSiteFirstWins pins that the first note of an error is kept and a
// value error is ignored.
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
	if len(rt.phpxSites) != 1 {
		t.Errorf("sites = %d, want 1", len(rt.phpxSites))
	}
}
