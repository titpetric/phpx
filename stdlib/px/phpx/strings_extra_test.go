//go:build phpx

package phpx

import "testing"

func TestTokState(t *testing.T) {
	tok := &tokState{s: "a,,b;c", pos: 0}
	var got []string
	for {
		v := tok.next(",;")
		s, ok := v.(string)
		if !ok {
			break
		}
		got = append(got, s)
	}
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("tokens = %v", got)
	}
}

func TestSimilarText(t *testing.T) {
	if n := similarText("World", "Word"); n != 4 {
		t.Fatalf("similar = %d", n)
	}
}

func TestStrIncrementDecrement(t *testing.T) {
	for in, want := range map[string]string{"a": "b", "z": "aa", "Az": "Ba", "zz": "aaa", "a9": "b0", "99": "100"} {
		if got, err := strIncrement(in); err != nil || got != want {
			t.Errorf("increment %q = %q, %v; want %q", in, got, err, want)
		}
	}
	for in, want := range map[string]string{"b": "a", "aa": "z", "Ba": "Az", "10": "9"} {
		if got, err := strDecrement(in); err != nil || got != want {
			t.Errorf("decrement %q = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := strDecrement("a"); err == nil {
		t.Error("decrementing a should fail")
	}
}

func TestSoundex(t *testing.T) {
	if soundex("Robert") != "R163" || soundex("Tymczak") != "T522" || soundex("Pfister") != "P236" {
		t.Fatalf("soundex: %s %s %s", soundex("Robert"), soundex("Tymczak"), soundex("Pfister"))
	}
}

func TestQuotedPrintable(t *testing.T) {
	enc := quotedPrintableEncode("héllo=")
	if enc != "h=C3=A9llo=3D" {
		t.Fatalf("encoded = %q", enc)
	}
	if quotedPrintableDecode(enc) != "héllo=" {
		t.Fatal("round trip")
	}
}
