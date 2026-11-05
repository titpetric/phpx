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

func TestMbEncodingsBeyondLatin1(t *testing.T) {
	if got := phpxConvertEncoding("\xe2", "UTF-8", "Windows-1252"); got != "â" {
		t.Errorf("cp1252 â = %q", got)
	}
	if got := phpxConvertEncoding("€", "Windows-1252", "UTF-8"); got != "\x80" {
		t.Errorf("euro to cp1252 = %q", got)
	}
	if got := phpxConvertEncoding("aé", "UTF-16BE", "UTF-8"); got != "\x00a\x00\xe9" {
		t.Errorf("utf-16be = %q", got)
	}
	if got := phpxConvertEncoding("日", "ISO-8859-1", "UTF-8"); got != "?" {
		t.Errorf("unmappable = %q", got)
	}
	if phpxValidIn("\xe9", "UTF-8") || !phpxValidIn("é", "UTF-8") {
		t.Error("UTF-8 validity")
	}
}
