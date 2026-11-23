//go:build phpx

package runner

import (
	"reflect"

	"github.com/titpetric/phpx/model"
)

// phpx_destruct.go is the destructor: php runs a class's `__destruct` when the
// last reference to an instance goes away, and runs what is still alive when
// the program ends.
//
// The runtime keeps no reference count, so an instance whose class declares
// `__destruct` is remembered at `new`, and a *candidate* is noted wherever a
// reference can go away: a statement whose value is discarded, an `unset()`, a
// variable written over, a frame that returns. A candidate is destroyed when a
// walk of every live root cannot reach it. The walk runs at the end of the
// statement that dropped something, and only while an instance with a
// destructor is alive, so a program that declares none pays one nil check.
//
// `__destruct` is a magic method php has and upstream's design leaves out.
// Laravel's `Route::resource()` returns a `PendingResourceRegistration` that
// registers its routes in the destructor, so half of a laravel app's routes are
// missing without one; phpx/testdata/classes/destruct.phpt settles the rest.

// phpxDtorState is the destructor bookkeeping of one program: the instances
// still alive in creation order, the same set as a map, the ones the destructor
// already ran on, and the instances a dropped reference may have freed.
type phpxDtorState struct {
	order   []*model.Object
	live    map[*model.Object]struct{}
	done    map[*model.Object]struct{}
	pending []*model.Object
	// closing is set while the program's end destroys what is left: every
	// instance is already noted there, so a statement inside a destructor
	// must not flush a candidate out of creation order.
	closing bool
}

// dtors returns the state, creating it on first use.
func (rt *Runtime) dtors() *phpxDtorState {
	if rt.phpxDtors == nil {
		rt.phpxDtors = &phpxDtorState{
			live: map[*model.Object]struct{}{},
			done: map[*model.Object]struct{}{},
		}
	}
	return rt.phpxDtors
}

// phpxTrackInstance remembers a freshly constructed instance whose class
// declares `__destruct`, so the program's end destroys it even if nothing
// drops the last reference.
func (rt *Runtime) phpxTrackInstance(obj *model.Object) {
	if obj == nil || obj.Class == nil {
		return
	}
	if _, ok := lookupPHPMethod(obj.Class, "__destruct"); !ok {
		return
	}
	state := rt.dtors()
	state.live[obj] = struct{}{}
	state.order = append(state.order, obj)
}

// phpxNoteDropped notes every tracked instance reachable from a value whose
// last reference may just have gone away. The flush decides whether it did.
func (rt *Runtime) phpxNoteDropped(v any) {
	if rt.phpxDtors == nil || v == nil {
		return
	}
	rt.phpxCollectDropped(v, newPHPMarkSet())
}

// phpxCollectDropped appends the tracked instances under v to the pending list.
func (rt *Runtime) phpxCollectDropped(v any, seen *phpMarkSet) {
	state := rt.phpxDtors
	switch x := v.(type) {
	case *model.Object:
		if x == nil {
			return
		}
		if _, ok := seen.objects[x]; ok {
			return
		}
		seen.objects[x] = struct{}{}
		if _, live := state.live[x]; live {
			state.pending = append(state.pending, x)
		}
		for _, val := range x.Props {
			rt.phpxCollectDropped(val, seen)
		}
	case *model.Array:
		if x == nil {
			return
		}
		if _, ok := seen.arrays[x]; ok {
			return
		}
		seen.arrays[x] = struct{}{}
		x.Range(func(key, val any) bool {
			rt.phpxCollectDropped(key, seen)
			rt.phpxCollectDropped(val, seen)
			return true
		})
	case *model.Cell:
		if x == nil {
			return
		}
		if _, ok := seen.cells[x]; ok {
			return
		}
		seen.cells[x] = struct{}{}
		rt.phpxCollectDropped(x.V, seen)
	default:
		phpxEachValue(v, func(item any) { rt.phpxCollectDropped(item, seen) })
	}
}

// phpxSetVar writes name while instances with a destructor are alive: it notes
// the value written over first, and marks the name when the new value could
// hold one. It is the same store as Scope.Set, with the bookkeeping in one
// call the caller makes only while such an instance exists.
func (rt *Runtime) phpxSetVar(scope *Scope, name string, val any) {
	held := phpxHeld(scope, name)
	if held {
		phpxUnhold(scope, name)
		old, ok := scope.vars[name]
		if !ok && (scope.refs != nil || scope.statics != nil) {
			old, ok = scope.Get(name)
		}
		if ok {
			rt.phpxNoteDropped(old)
		}
	}
	scope.Set(name, val)
	switch val.(type) {
	case *model.Object, *model.Array, *model.Cell:
		// The name was unmarked above when it was marked, so it goes back on
		// the list either way: the new value may hold another instance.
		scope.phpxHolds = append(scope.phpxHolds, name)
	}
}

// phpxHeld reports whether the frame marked name.
func phpxHeld(scope *Scope, name string) bool {
	for _, held := range scope.phpxHolds {
		if held == name {
			return true
		}
	}
	return false
}

// phpxUnhold drops name from the frame's marked names.
func phpxUnhold(scope *Scope, name string) {
	for i, held := range scope.phpxHolds {
		if held == name {
			scope.phpxHolds = append(scope.phpxHolds[:i], scope.phpxHolds[i+1:]...)
			return
		}
	}
}

// phpxFramePopped notes what a returning frame held. The instance is destroyed
// at the end of the statement whose call returned it, when nothing the caller
// kept reaches it.
func (rt *Runtime) phpxFramePopped(scope *Scope) {
	if rt.phpxDtors == nil || scope == nil {
		return
	}
	for _, val := range scope.vars {
		rt.phpxNoteDropped(val)
	}
	for _, cell := range scope.refs {
		rt.phpxNoteDropped(cell)
	}
	for _, bag := range scope.statics {
		for _, val := range bag {
			rt.phpxNoteDropped(val)
		}
	}
	for _, val := range scope.args {
		rt.phpxNoteDropped(val)
	}
}

// phpxHasPending reports whether a dropped reference may have freed an
// instance the flush should look at.
func (rt *Runtime) phpxHasPending() bool {
	return rt.phpxDtors != nil && !rt.phpxDtors.closing && rt.phpxDtors.pending != nil
}

// phpxFlush destroys every pending instance the live roots cannot reach, then
// the instances that became unreachable when those went. extra is a value the
// running statement still holds, such as the statement's own result.
func (rt *Runtime) phpxFlush(extra any) error {
	state := rt.phpxDtors
	for pass := 0; pass < 32; pass++ {
		if state == nil || len(state.pending) == 0 {
			return nil
		}
		pending := state.pending
		state.pending = nil
		seen := rt.phpxMarkReachable(extra)
		progress := false
		for _, obj := range pending {
			if _, done := state.done[obj]; done {
				continue
			}
			if _, reachable := seen.objects[obj]; reachable {
				continue
			}
			progress = true
			if err := rt.phpxDestroy(obj); err != nil {
				return err
			}
		}
		if !progress {
			return nil
		}
	}
	return nil
}

// phpxDestroyLive destroys every instance still alive when the program ends,
// in reverse creation order, as php tears its object store down after the
// shutdown functions.
func (rt *Runtime) phpxDestroyLive() {
	state := rt.phpxDtors
	if state == nil {
		return
	}
	state.closing = true
	state.pending = nil
	for i := len(state.order) - 1; i >= 0; i-- {
		if err := rt.phpxDestroy(state.order[i]); err != nil {
			rt.RecordError(err)
			return
		}
	}
	rt.phpxDtors = nil
}

// phpxDestroy runs the instance's destructor once and notes what its
// properties held, which php frees after the body returns.
func (rt *Runtime) phpxDestroy(obj *model.Object) error {
	if obj == nil || rt.phpxDtors == nil {
		return nil
	}
	state := rt.phpxDtors
	if _, done := state.done[obj]; done {
		return nil
	}
	state.done[obj] = struct{}{}
	delete(state.live, obj)
	defer func() {
		for _, val := range obj.Props {
			rt.phpxNoteDropped(val)
		}
	}()
	call, ok := rt.boundMethod(obj, "__destruct", nil)
	if !ok {
		return nil
	}
	_, err := call()
	return err
}

// phpMarkSet dedups the containers a reachability walk visits by identity.
type phpMarkSet struct {
	objects map[*model.Object]struct{}
	arrays  map[*model.Array]struct{}
	cells   map[*model.Cell]struct{}
}

func newPHPMarkSet() *phpMarkSet {
	return &phpMarkSet{
		objects: map[*model.Object]struct{}{},
		arrays:  map[*model.Array]struct{}{},
		cells:   map[*model.Cell]struct{}{},
	}
}

// phpxMarkReachable collects every object the live roots reach: the frames,
// the globals and auto globals, the class and function statics, the bytecode
// engine's live values and the statement's own result.
func (rt *Runtime) phpxMarkReachable(extra any) *phpMarkSet {
	seen := newPHPMarkSet()
	mark := func(v any) { phpxMarkValue(v, seen) }
	for _, scope := range rt.frames {
		for _, val := range scope.vars {
			mark(val)
		}
		for _, cell := range scope.refs {
			phpxMarkValue(cell, seen)
		}
		for _, bag := range scope.statics {
			for _, val := range bag {
				mark(val)
			}
		}
		for _, val := range scope.args {
			mark(val)
		}
		for _, val := range scope.deferred {
			mark(val)
		}
	}
	for _, val := range rt.globals {
		mark(val)
	}
	rt.auto.Range(func(_ string, val any) {
		mark(val)
	})
	for _, bag := range rt.classStatics {
		for _, val := range bag {
			mark(val)
		}
	}
	for _, bag := range rt.funcStatics {
		for _, val := range bag {
			mark(val)
		}
	}
	for _, walk := range rt.vmWalkers {
		walk(mark)
	}
	mark(extra)
	return seen
}

// phpxMarkValue marks v and everything it reaches.
func phpxMarkValue(v any, seen *phpMarkSet) {
	switch x := v.(type) {
	case *model.Object:
		if x == nil {
			return
		}
		if _, ok := seen.objects[x]; ok {
			return
		}
		seen.objects[x] = struct{}{}
		for _, val := range x.Props {
			phpxMarkValue(val, seen)
		}
	case *model.Array:
		if x == nil {
			return
		}
		if _, ok := seen.arrays[x]; ok {
			return
		}
		seen.arrays[x] = struct{}{}
		x.Range(func(key, val any) bool {
			phpxMarkValue(key, seen)
			phpxMarkValue(val, seen)
			return true
		})
	case *model.Cell:
		if x == nil {
			return
		}
		if _, ok := seen.cells[x]; ok {
			return
		}
		seen.cells[x] = struct{}{}
		phpxMarkValue(x.V, seen)
	default:
		phpxEachValue(v, func(item any) { phpxMarkValue(item, seen) })
	}
}

// phpxEachValue calls fn for every element of a Go collection a binding
// returned. A script value of another shape holds nothing the runtime can
// follow; a Go closure's captures are one such shape, recorded as a limit.
func phpxEachValue(v any, fn func(any)) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if item := rv.Index(i); item.CanInterface() {
				fn(item.Interface())
			}
		}
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			if key := iter.Key(); key.CanInterface() {
				fn(key.Interface())
			}
			if val := iter.Value(); val.CanInterface() {
				fn(val.Interface())
			}
		}
	}
}
