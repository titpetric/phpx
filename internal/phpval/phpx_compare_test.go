package phpval

import "testing"

type tick int

func (t tick) PhpxCompare(o any) (int, bool) {
	u, ok := o.(tick)
	if !ok {
		return 0, false
	}
	return int(t) - int(u), true
}

func TestHostCompare(t *testing.T) {
	if Compare(tick(1), tick(2)) >= 0 || Compare(tick(2), tick(1)) <= 0 || Compare(tick(3), tick(3)) != 0 {
		t.Fatal("a host value orders itself")
	}
}
