//go:build phpx

package phpx

import "testing"

// TestPhpxEncodingDetect pins the names and the strict detection order.
func TestPhpxEncodingDetect(t *testing.T) {
	t.Parallel()
	if got := phpxEncodingName("utf8"); got != "UTF-8" {
		t.Errorf("utf8 = %q", got)
	}
	if !phpxValidIn("abc", "ASCII") || phpxValidIn("é", "ASCII") {
		t.Error("ASCII validity wrong")
	}
	if got := phpxConvertEncoding("é", "ISO-8859-1", "UTF-8"); got != "\xe9" {
		t.Errorf("latin1 = %q", got)
	}
	if got := phpxConvertEncoding("\xe9", "UTF-8", "ISO-8859-1"); got != "é" {
		t.Errorf("utf8 = %q", got)
	}
}
