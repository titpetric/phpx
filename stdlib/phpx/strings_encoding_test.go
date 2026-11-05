//go:build phpx

package phpx

import "testing"

func TestCheckEncoding(t *testing.T) {
	if !checkEncoding("é") || checkEncoding("\xff") {
		t.Fatal("utf-8 by default")
	}
	if checkEncoding("é", "ASCII") || !checkEncoding("abc", "ASCII") {
		t.Fatal("ascii allows the low half only")
	}
	if !checkEncoding("\xff", "ISO-8859-1") {
		t.Fatal("a single-byte encoding takes every byte")
	}
}
