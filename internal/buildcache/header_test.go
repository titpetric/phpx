package buildcache

import (
	"crypto/sha256"
	"testing"
)

func TestHeaderRoundTrip(t *testing.T) {
	t.Parallel()
	want := header{Build: "abc", Name: "/src/a.php", Options: "go:false", Content: sha256.Sum256([]byte("x"))}
	data := append(want.header(), []byte("payload")...)
	got, payload, err := split(data)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("split header = %+v, want %+v", got, want)
	}
	if string(payload) != "payload" {
		t.Fatalf("payload = %q, want payload", payload)
	}
}

func TestSplitRefusesAnotherFormat(t *testing.T) {
	t.Parallel()
	h := header{Build: "abc", Name: "a", Options: "o"}
	data := h.header()
	data[len(magic)] = Format + 1
	if _, _, err := split(data); err == nil {
		t.Fatal("split read an entry written to another format")
	}
	data = h.header()
	data[0] = 'x'
	if _, _, err := split(data); err == nil {
		t.Fatal("split read an entry with no magic")
	}
}

func TestSplitRefusesShortHeader(t *testing.T) {
	t.Parallel()
	h := header{Build: "abc", Name: "a", Options: "o"}
	data := h.header()
	for i := range data {
		if _, _, err := split(data[:i]); err == nil {
			t.Fatalf("split(%d of %d bytes) = nil error", i, len(data))
		}
	}
}

func TestDefaultDirNamesTheCache(t *testing.T) {
	t.Setenv("PHPScript_CACHE", "/tmp/cachehere")
	if got := DefaultDir(); got != "/tmp/cachehere" {
		t.Fatalf("DefaultDir = %q, want the override", got)
	}
	t.Setenv("PHPScript_CACHE", "")
	if got := DefaultDir(); got == "" {
		t.Fatal("DefaultDir = \"\", want the user cache directory")
	}
}
