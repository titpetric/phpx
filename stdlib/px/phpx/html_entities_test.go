//go:build phpx

package phpx

import "testing"

func TestHTMLEscape(t *testing.T) {
	cases := []struct {
		in          string
		flags       int64
		double, all bool
		want        string
	}{
		{`a<b>&"'`, entQuotes, true, false, "a&lt;b&gt;&amp;&quot;&#039;"},
		{`"'`, entCompat, true, false, "&quot;'"},
		{`"'`, entNoQuotes, true, false, `"'`},
		{"&amp; &lt;", entQuotes, false, false, "&amp; &lt;"},
		{"&amp;", entQuotes, true, false, "&amp;amp;"},
		{"café €", entQuotes, true, true, "caf&eacute; &euro;"},
		{"'", entQuotes | entHTML5, true, false, "&apos;"},
	}
	for _, c := range cases {
		if got := htmlEscape(c.in, c.flags, c.double, c.all); got != c.want {
			t.Errorf("htmlEscape(%q, %d) = %q, want %q", c.in, c.flags, got, c.want)
		}
	}
}

func TestHTMLDecode(t *testing.T) {
	if got := htmlDecode("&lt;b&gt; &amp;amp; &quot;x&quot; &#039;", entQuotes, false); got != `<b> &amp; "x" '` {
		t.Errorf("decode = %q", got)
	}
	if got := htmlDecode("&quot;&#039;", entCompat, false); got != `"&#039;` {
		t.Errorf("compat decode = %q", got)
	}
	if got := htmlDecode("caf&eacute; &euro; &#8364; &bogus;", entQuotes, true); got != "café € € &bogus;" {
		t.Errorf("entity decode = %q", got)
	}
}
