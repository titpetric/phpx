package model

import (
	"fmt"
	"math"
	"strconv"
)

// Array is PHP's ordered hash map: insertion order, integer and string keys,
// list and dictionary in one value.
//
// It starts as a dense slice and promotes itself to a map on the first key that
// is not the next index, which callers cannot observe; see promote. Keys are opaque here, and normalising "1" to 1 is the caller's job
// (runner.normalizeKey).
//
// A copy is a copy-on-write pair: Clone answers a second handle over the same
// storage and both handles read it, while the first write to either, through a
// method here, splits the storage in two. PHP makes the copy at assignment and
// defers the duplication the same way, so `$b = $a` costs one handle until one
// side writes.
type Array struct {
	list   []any // list mode only: element i is the value of key int64(i)
	keys   []any // map mode only: keys in insertion order
	values map[any]any
	nextID int64
	shared bool // the storage may be read through another handle; split first
	// cells records that an element is a php reference (a Cell). An ordinary
	// array never sets it, so the hot read and iteration paths skip the
	// per-element type check.
	cells bool
	// version counts the writes to this handle's entries. The destructor's
	// reachability cache reads it to tell a subtree it has already walked from
	// one a write has changed; a bump over-approximates, which only costs a
	// walk.
	version uint32
}

// Version counts the writes made to the array through this handle.
func (a *Array) Version() uint32 { return a.version }

// noIntKey is nextID's value for a map-mode array holding no integer key.
const noIntKey = math.MinInt64

// NewArray returns an empty ordered array.
func NewArray() *Array {
	return &Array{nextID: noIntKey}
}

// NewArraySize returns an empty ordered array with room for n entries. The
// presize drops the backing slice's growth reallocations in list mode.
func NewArraySize(n int) *Array {
	if n <= 0 {
		return NewArray()
	}
	return &Array{list: make([]any, 0, n), nextID: noIntKey}
}

// Clone returns a second handle over a's storage: PHP's array copy before
// either side writes. Both handles are marked shared, so the first write to
// either splits it. A nil array answers nil.
func (a *Array) Clone() *Array {
	if a == nil {
		return nil
	}
	out := *a
	a.shared = true
	out.shared = true
	return &out
}

// Shared reports whether the storage may be read through another handle.
func (a *Array) Shared() bool { return a != nil && a.shared }

// Split gives a its own copy of the storage when another handle shares it, so
// the write that follows touches this handle alone. A nested array becomes a
// handle of its own over shared storage: its split happens at the write that
// needs it.
func (a *Array) Split() {
	if !a.shared {
		return
	}
	a.shared = false
	a.list = splitValues(a.list)
	a.keys = splitValues(a.keys)
	if a.values != nil {
		a.values = splitMap(a.values)
	}
}

// splitValues copies a list of entries into a slice of its own, giving each
// nested array a handle of its own over the storage it shares.
func splitValues(in []any) []any {
	if in == nil {
		return nil
	}
	out := make([]any, len(in))
	for i, v := range in {
		if arr, ok := v.(*Array); ok {
			out[i] = arr.Clone()
			continue
		}
		out[i] = v
	}
	return out
}

// splitMap copies the entries of a map-mode array into a map of its own, one
// handle per nested array.
func splitMap(in map[any]any) map[any]any {
	out := make(map[any]any, len(in))
	for k, v := range in {
		if arr, ok := v.(*Array); ok {
			out[k] = arr.Clone()
			continue
		}
		out[k] = v
	}
	return out
}

// Reset empties the array, keeping the storage the next request refills into.
func (a *Array) Reset() {
	a.version++
	a.Split()
	clear(a.list)
	a.list = a.list[:0]
	clear(a.keys)
	a.keys = a.keys[:0]
	clear(a.values)
	// The sentinel, not 0, makes the claim above true: a reset array that held
	// an integer key would otherwise append at 0. A fresh array appends from
	// the first key the caller sets.
	a.nextID = noIntKey
	a.cells = false
}

// isList reports whether the array is still in list mode.
func (a *Array) isList() bool { return a.values == nil }

// promote switches list mode to map mode, materialising the implicit integer keys.
func (a *Array) promote() {
	n := len(a.list)
	size := n
	if c := cap(a.list); c > size {
		size = c
	}
	a.values = make(map[any]any, size)
	if n == 0 {
		a.keys = a.list[:0]
	} else {
		a.keys = make([]any, n, size)
		// One box per key serves both the map key and the order slice: the
		// interface value is immutable, so sharing it changes nothing a caller
		// can observe and halves the boxing allocations for keys past 255.
		for i, v := range a.list {
			k := any(int64(i))
			a.keys[i] = k
			a.values[k] = v
		}
	}
	a.list = nil
}

// Set assigns key=val, appending the key if new.
//
// An element that holds a Cell is a php reference (`$a[0] = &$x`). Assigning a
// value to it writes through the cell, so every other name bound to it sees the
// write, as php does. Assigning a Cell rebinds the element to that slot, which
// is `$a[0] = &$y`: the old reference is broken, the new one takes its place.
func (a *Array) Set(key, val any) {
	a.version++
	a.Split()
	if _, isCell := val.(*Cell); isCell {
		a.cells = true
	}
	if a.isList() {
		// In list mode nextID is always len(list) and an int64 key inside
		// [0, len] either overwrites an element or extends the list by one.
		if i, ok := key.(int64); ok && i >= 0 && i <= int64(len(a.list)) {
			if i == int64(len(a.list)) {
				a.list = append(a.list, val)
				a.nextID = i + 1
			} else if a.cells {
				if cell, isCell := a.list[i].(*Cell); isCell {
					if bound, isBound := val.(*Cell); isBound {
						a.list[i] = bound
					} else {
						cell.Set(val)
					}
				} else {
					a.list[i] = val
				}
			} else {
				a.list[i] = val
			}
			return
		}
		a.promote()
	}
	// One map write and a length check in place of a lookup and a write: the
	// length moves exactly when the key is new, which is when the order slice
	// takes it.
	if a.cells {
		if cur, ok := a.values[key]; ok {
			if cell, isCell := cur.(*Cell); isCell {
				if bound, isBound := val.(*Cell); isBound {
					a.values[key] = bound
				} else {
					cell.Set(val)
				}
				return
			}
		}
	}
	size := len(a.values)
	a.values[key] = val
	if len(a.values) != size {
		a.keys = append(a.keys, key)
	}
	if i, ok := key.(int64); ok && i >= a.nextID {
		a.nextID = i + 1
	}
}

// Append adds val at the next integer index (PHP `$a[] = v`).
func (a *Array) Append(val any) {
	a.version++
	a.Split()
	if _, isCell := val.(*Cell); isCell {
		a.cells = true
	}
	if a.isList() {
		a.list = append(a.list, val)
		// Assigned without increment, since a list's next index equals its
		// length by definition. The counter may still hold the sentinel.
		a.nextID = int64(len(a.list))
		return
	}
	if a.nextID == noIntKey {
		// The index holds no value yet and this is the first integer key and PHP
		// answers 0.
		a.Set(int64(0), val)
		return
	}
	a.Set(a.nextID, val)
}

// NextIndex is the integer key the next Append lands on.
func (a *Array) NextIndex() int64 {
	return a.nextID
}

// AppendAll adds every item at the next integer index, in one copy where the
// array is still a list.
//
// A Go slice is already a list and the fast path pays one allocation for the
// whole copy where Append per element pays on each growth step.
func (a *Array) AppendAll(items []any) {
	if len(items) == 0 {
		return
	}
	a.version++
	a.Split()
	if !a.isList() {
		for _, item := range items {
			a.Append(item)
		}
		return
	}
	a.list = append(a.list, items...)
	a.nextID = int64(len(a.list))
}

// Get returns the value for key and whether it existed. An element bound to a
// reference answers the value the cell holds.
func (a *Array) Get(key any) (any, bool) {
	v, ok := a.raw(key)
	if !ok {
		return nil, false
	}
	if a.cells {
		return CellValue(v), true
	}
	return v, true
}

// raw answers the value stored under key, a Cell for a reference element.
func (a *Array) raw(key any) (any, bool) {
	if a.isList() {
		if i, ok := key.(int64); ok && i >= 0 && i < int64(len(a.list)) {
			return a.list[i], true
		}
		return nil, false
	}
	v, ok := a.values[key]
	return v, ok
}

// CellAt answers the cell key is bound to, when the element is a reference
// (`$a[0] = &$x`). An element holding a plain value answers false.
func (a *Array) CellAt(key any) (*Cell, bool) {
	if !a.cells {
		return nil, false
	}
	v, ok := a.raw(key)
	if !ok {
		return nil, false
	}
	cell, ok := v.(*Cell)
	return cell, ok
}

// ElementCell answers the cell behind key, making one around the current value
// the first time a reference is taken: a by-reference foreach binds the loop
// variable and the element to one slot that outlives the loop.
func (a *Array) ElementCell(key any) *Cell {
	if cell, ok := a.CellAt(key); ok {
		return cell
	}
	v, _ := a.Get(key)
	cell := &Cell{V: v}
	a.Set(key, cell)
	return cell
}

// Delete removes key, preserving the order of the entries around it (PHP's
// unset). A list-mode array promotes itself first: dropping an element from
// the middle would leave the remaining keys sparse, which list mode cannot
// express, and dropping the last element would still leave nextID past the
// end, matching PHP, since unset never renumbers.
func (a *Array) Delete(key any) {
	a.version++
	a.Split()
	if a.isList() {
		i, ok := key.(int64)
		if !ok || i < 0 || i >= int64(len(a.list)) {
			return
		}
		if i == int64(len(a.list))-1 {
			a.list = a.list[:i]
			return
		}
		a.promote()
	}
	if _, ok := a.values[key]; !ok {
		return
	}
	delete(a.values, key)
	for i, k := range a.keys {
		if k == key {
			a.keys = append(a.keys[:i], a.keys[i+1:]...)
			break
		}
	}
}

// Pop removes the last entry and returns its key and value, PHP's array_pop.
// The append index drops only when the removed key equals the next key it
// would hand out: popping 9 from [5 => a, 9 => c] leaves the next append at 9.
func (a *Array) Pop() (any, any, bool) {
	if a.Len() == 0 {
		return nil, nil, false
	}
	keys := a.Keys()
	key := keys[len(keys)-1]
	value, _ := a.Get(key)
	a.Delete(key)
	// An array holding an integer key has set the index. The two cannot both
	// hold, and nextID-1 on the sentinel would wrap. The guard excludes the
	// sentinel.
	if i, ok := key.(int64); ok && a.nextID != noIntKey && i == a.nextID-1 {
		a.nextID = i
	}
	return key, value, true
}

// Len reports the number of entries.
func (a *Array) Len() int {
	if a.isList() {
		return len(a.list)
	}
	return len(a.keys)
}

// Keys returns keys in insertion order. A list-mode array materialises them on
// each call, since it does not store them.
func (a *Array) Keys() []any {
	if a.isList() {
		keys := make([]any, len(a.list))
		for i := range keys {
			keys[i] = int64(i)
		}
		return keys
	}
	return a.keys
}

// Range iterates entries in insertion order. An ordinary array, the common
// case, runs the small loop the compiler can inline; an array holding a
// reference element takes the slower loop, which answers the value its cell
// holds.
func (a *Array) Range(fn func(key, val any) bool) {
	// The flag is read once: an ordinary array, the common case, pays one
	// predicted branch per element and keeps the small loop the compiler
	// inlines into the caller.
	cells := a.cells
	if a.isList() {
		for i, v := range a.list {
			if cells {
				v = CellValue(v)
			}
			if !fn(int64(i), v) {
				return
			}
		}
		return
	}
	for _, k := range a.keys {
		v := a.values[k]
		if cells {
			v = CellValue(v)
		}
		if !fn(k, v) {
			return
		}
	}
}

// RangeRaw iterates entries in insertion order, answering the stored value: a
// Cell for a reference element, which a by-reference loop binds to.
func (a *Array) RangeRaw(fn func(key, val any) bool) {
	if a.isList() {
		for i, v := range a.list {
			if !fn(int64(i), v) {
				return
			}
		}
		return
	}
	for _, k := range a.keys {
		if !fn(k, a.values[k]) {
			return
		}
	}
}

// Map returns the array as a string-keyed map for Go APIs that accept named
// values. A PHP integer key becomes its decimal string form.
func (a *Array) Map() map[string]any {
	if a.isList() {
		result := make(map[string]any, len(a.list))
		for i, v := range a.list {
			result[strconv.Itoa(i)] = a.plain(v)
		}
		return result
	}
	result := make(map[string]any, len(a.keys))
	for _, key := range a.keys {
		result[fmt.Sprint(key)] = a.plain(a.values[key])
	}
	return result
}

// plain answers the value a stored entry reads, unwrapping a reference element.
func (a *Array) plain(v any) any {
	if a.cells {
		return CellValue(v)
	}
	return v
}

// Int64List reports whether a is a dense list of int64 values and returns a
// copy of those values.
func (a *Array) Int64List() ([]int64, bool) {
	if a == nil || !a.isList() {
		return nil, false
	}
	out := make([]int64, len(a.list))
	for i, v := range a.list {
		n, ok := a.plain(v).(int64)
		if !ok {
			return nil, false
		}
		out[i] = n
	}
	return out, true
}

// ReplaceInt64List puts vals back as a dense int64 list.
func (a *Array) ReplaceInt64List(vals []int64) {
	a.version++
	a.keys = nil
	a.values = nil
	a.list = make([]any, len(vals))
	for i, v := range vals {
		a.list[i] = v
	}
	a.nextID = int64(len(vals))
	if len(vals) == 0 {
		a.nextID = noIntKey
	}
	// The storage is this handle's alone from here, whatever it shared before.
	a.shared = false
}

// Clear removes all entries and returns the array to list mode.
func (a *Array) Clear() {
	a.version++
	a.list = nil
	a.keys = nil
	a.values = nil
	a.nextID = 0
	a.shared = false
	a.cells = false
}
