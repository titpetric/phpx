//go:build phpx

package phpx

import "testing"

// The first words of php's mt_srand(1) sequence, from php 8.4.
func TestMtRandMatchesPhp(t *testing.T) {
	m := &mtState{}
	m.seed(1)
	want := []int64{895547922, 2141438069, 1546885062, 2002651684, 245631}
	for i, w := range want {
		got, err := m.rand(nil)
		if err != nil || got != w {
			t.Fatalf("mt_rand #%d = %d, %v; want %d", i, got, err, w)
		}
	}
}

func TestMtRandRange(t *testing.T) {
	m := &mtState{}
	m.seed(42)
	for i := 0; i < 1000; i++ {
		if v := m.between(5, 9); v < 5 || v > 9 {
			t.Fatalf("out of range: %d", v)
		}
	}
	if _, err := m.rand([]any{int64(3), int64(2)}); err == nil {
		t.Fatal("max below min should fail")
	}
}
