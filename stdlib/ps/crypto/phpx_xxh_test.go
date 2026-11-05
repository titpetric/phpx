//go:build phpx

package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

// TestXXHDigests pins the four algorithms against the values php 8.4 prints.
func TestXXHDigests(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 40)
	cases := []struct{ algo, in, want string }{
		{"xxh32", "", "02cc5d05"},
		{"xxh32", "abc", "32d153ff"},
		{"xxh64", "", "ef46db3751d8e999"},
		{"xxh64", "abc", "44bc2cf5ad770999"},
		{"xxh3", "abc", "78af5f94892f3950"},
		{"xxh3", "", "2d06800538d394c2"},
		{"xxh128", "abc", "06b05ab6733a618578af5f94892f3950"},
		{"xxh128", "", "99aa06d3014798d86001c324468d497f"},
		{"xxh128", long, "82137ef846680bfa7372e42cfb671208"},
		{"xxh64", long, "754481191093fcd7"},
		{"xxh32", long, "f84bccfe"},
	}
	for _, c := range cases {
		h := digests[c.algo]()
		_, _ = h.Write([]byte(c.in))
		if got := hex.EncodeToString(h.Sum(nil)); got != c.want {
			t.Errorf("%s(%.10q) = %s, want %s", c.algo, c.in, got, c.want)
		}
	}
}
