//go:build phpx

package phpx

import "testing"

func TestCsvArgs(t *testing.T) {
	d, e, esc, err := csvArgs("str_getcsv", []any{";", "'", ""})
	if err != nil || d != ';' || e != '\'' || esc != -1 {
		t.Fatalf("got %c %c %d %v", d, e, esc, err)
	}
	if _, _, _, err := csvArgs("str_getcsv", []any{"ab"}); err == nil {
		t.Fatal("a two-byte separator should be refused")
	}
}

func TestCsvLine(t *testing.T) {
	arr := csvArray([]any{"a", "b c", "d\"e", "f,g", ""})
	if got := csvLine(arr, ',', '"', '\\', "\n"); got != "a,\"b c\",\"d\"\"e\",\"f,g\",\n" {
		t.Fatalf("csvLine = %q", got)
	}
}
