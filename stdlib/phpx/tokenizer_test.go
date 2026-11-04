//go:build phpx

package phpx

import "testing"

func TestPhpTokensEmitPhp84Ids(t *testing.T) {
	t.Parallel()
	l := &phpTokens{src: "<?php fn($x) => $x ?? throw new E;", line: 1}
	l.run()
	seen := map[int64]string{}
	for _, tok := range l.out {
		seen[tok.id] = tok.text
	}
	for name, text := range map[string]string{"T_OPEN_TAG": "<?php ", "T_FN": "fn", "T_COALESCE": "??", "T_THROW": "throw", "T_NEW": "new"} {
		if got := seen[phpxTokenConstants[name]]; got != text {
			t.Errorf("%s: got %q, want %q", name, got, text)
		}
	}
}
