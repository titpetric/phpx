package model

import "testing"

// TestArrayCellAt pins the reference element lookup: only a bound element
// answers a cell.
func TestArrayCellAt(t *testing.T) {
	arr := NewArray()
	cell := &Cell{V: int64(1)}
	arr.Set(int64(0), cell)
	if got, ok := arr.CellAt(int64(0)); !ok || got != cell {
		t.Fatalf("CellAt = %v, %v; want the bound cell", got, ok)
	}
	if _, ok := arr.CellAt(int64(1)); ok {
		t.Fatal("CellAt answered a cell for a missing key")
	}
	plain := NewArray()
	plain.Append(int64(1))
	if _, ok := plain.CellAt(int64(0)); ok {
		t.Fatal("CellAt answered a cell for a plain element")
	}
}

// TestArrayElementCell pins the slot a by-reference foreach binds: made once,
// and it reads back through the element.
func TestArrayElementCell(t *testing.T) {
	arr := NewArray()
	arr.Append(int64(5))
	cell := arr.ElementCell(int64(0))
	cell.V = int64(9)
	if v, _ := arr.Get(int64(0)); v != int64(9) {
		t.Fatalf("Get = %v, want 9", v)
	}
	if again := arr.ElementCell(int64(0)); again != cell {
		t.Fatal("ElementCell made a second cell for one slot")
	}
}

// TestArrayRangeRaw pins the iterator that answers a reference element as its
// cell while Range answers its value.
func TestArrayRangeRaw(t *testing.T) {
	arr := NewArray()
	arr.Append(int64(1))
	arr.Append(&Cell{V: int64(2)})
	var raw int
	arr.RangeRaw(func(_, val any) bool {
		if _, ok := val.(*Cell); ok {
			raw++
		}
		return true
	})
	if raw != 1 {
		t.Fatalf("RangeRaw answered %d cells, want 1", raw)
	}
	arr.Range(func(_, val any) bool {
		if _, ok := val.(*Cell); ok {
			t.Fatal("Range answered a cell")
		}
		return true
	})
}

// TestRefArgSetter pins the setter a Go binding takes for an output parameter.
func TestRefArgSetter(t *testing.T) {
	cell := &Cell{V: "before"}
	ra := &RefArg{Value: "before", Make: func() *Cell { return cell }}
	ra.Setter()("after")
	if cell.V != "after" {
		t.Fatalf("cell = %v, want after", cell.V)
	}
	(&RefArg{}).Setter()("ignored")
}
