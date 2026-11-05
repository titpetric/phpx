//go:build phpx

package phpx

import "testing"

func TestMbConvertCase(t *testing.T) {
	cases := []struct {
		in   string
		mode int
		want string
	}{
		{"Straße", mbCaseUpper, "STRASSE"},
		{"ÀB", mbCaseLower, "àb"},
		{"hello wORLD", mbCaseTitle, "Hello World"},
		{"o'neil", mbCaseTitle, "O'neil"},
	}
	for _, c := range cases {
		if got := mbConvertCase(c.in, int64(c.mode)); got != c.want {
			t.Errorf("mbConvertCase(%q, %d) = %q, want %q", c.in, c.mode, got, c.want)
		}
	}
}

func TestMbTrimAndCut(t *testing.T) {
	if got := mbTrim("  é ", true, true, nil); got != "é" {
		t.Errorf("mbTrim = %q", got)
	}
	if got := mbStrcut("héllo", 2, []any{int64(3)}); got != "él" && got != "é" {
		t.Errorf("mbStrcut = %q", got)
	}
}
