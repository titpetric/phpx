//go:build phpx

package core

import "testing"

// TestPhpxHighlightSource checks the highlighter against PHP's own output for a
// source that exercises the colours: a cast, a member named by a keyword, and a
// heredoc with interpolation.
func TestPhpxHighlightSource(t *testing.T) {
	t.Parallel()
	src := "<?php\nfunction f(int $x): string { return (string) $x; }\n" +
		"$o->include(\"a\\tb\");\n$s = <<<EOT\nv $x\nEOT;\n"
	want := `<pre><code style="color: #000000"><span style="color: #0000BB">&lt;?php
</span><span style="color: #007700">function </span><span style="color: #0000BB">f</span><span style="color: #007700">(</span><span style="color: #0000BB">int $x</span><span style="color: #007700">): </span><span style="color: #0000BB">string </span><span style="color: #007700">{ return (string) </span><span style="color: #0000BB">$x</span><span style="color: #007700">; }
</span><span style="color: #0000BB">$o</span><span style="color: #007700">-&gt;</span><span style="color: #0000BB">include</span><span style="color: #007700">(</span><span style="color: #DD0000">"a\tb"</span><span style="color: #007700">);
</span><span style="color: #0000BB">$s </span><span style="color: #007700">= &lt;&lt;&lt;EOT
</span><span style="color: #DD0000">v </span><span style="color: #0000BB">$x</span><span style="color: #DD0000">
</span><span style="color: #007700">EOT;
</span></code></pre>`
	if got := phpxHighlightSource(src); got != want {
		t.Fatalf("highlight mismatch:\n got %q\nwant %q", got, want)
	}
}

// TestPhpxHighlightInlineHTML checks that text outside PHP prints escaped and
// outside a span, in the wrapper's colour.
func TestPhpxHighlightInlineHTML(t *testing.T) {
	t.Parallel()
	got := phpxHighlightSource("a & b\n<?= 1 ?>c")
	want := `<pre><code style="color: #000000">a &amp; b
<span style="color: #0000BB">&lt;?= 1 ?&gt;</span>c</code></pre>`
	if got != want {
		t.Fatalf("highlight mismatch:\n got %q\nwant %q", got, want)
	}
}
