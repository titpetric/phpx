//go:build phpx

package runner

import "testing"

func TestScanSourceSkipsHeredocs(t *testing.T) {
	src := "<?php\nclass T {\n  const X = <<<'EOS'\n{{# a }}\n{{/ a }}\nEOS;\n\n  /** doc */\n  public function m() {}\n}\n"
	meta := phpxScanSource(src)
	if d, ok := meta.Decls["T::m"]; !ok || d.Doc != "/** doc */" {
		t.Fatalf("T::m lost behind a heredoc: %v", meta.Decls)
	}
}
