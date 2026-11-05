package parser

import "testing"

func TestIdentBytesAbove7F(t *testing.T) {
	if !isIdentStart(0xCA) || !isIdentPart(0x9F) {
		t.Fatal("bytes at or above 0x80 belong to identifiers")
	}
	if isIdentStart('$') || isIdentPart('-') {
		t.Fatal("operators are not identifier bytes")
	}
}
