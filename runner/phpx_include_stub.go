//go:build !phpx

package runner

// phpxIncludeAnswersFalse answers the untagged build, where a failed include ends the program.
func phpxIncludeAnswersFalse(_ string, _ error) bool { return false }

// phpxIncludeWarning answers the untagged build, which emits no include warning.
func phpxIncludeWarning(_ *Runtime, _, _ string, _ error) {}
