//go:build phpx

package phpx

import "testing"

func TestSscanfRun(t *testing.T) {
	vals, n, _ := sscanfRun("age: 25 name: Bob 3.5", "age: %d name: %s %f")
	if n != 3 || len(vals) != 3 || vals[0] != int64(25) || vals[1] != "Bob" || vals[2] != 3.5 {
		t.Fatalf("vals %v matched %d", vals, n)
	}
	vals, _, _ = sscanfRun("id-42", "id-%d")
	if len(vals) != 1 || vals[0] != int64(42) {
		t.Fatalf("vals %v", vals)
	}
	vals, n, _ = sscanfRun("abc", "%d")
	if n != 0 || len(vals) != 1 || vals[0] != nil {
		t.Fatalf("a failed conversion gives null: %v %d", vals, n)
	}
	vals, _, _ = sscanfRun("12:34:56", "%d:%d:%d")
	if len(vals) != 3 || vals[2] != int64(56) {
		t.Fatalf("vals %v", vals)
	}
	vals, _, _ = sscanfRun("hello world", "%[a-z] %s")
	if vals[0] != "hello" || vals[1] != "world" {
		t.Fatalf("set %v", vals)
	}
}
