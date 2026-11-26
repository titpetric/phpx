//go:build phpx

package runner

import (
	"fmt"
	"os"
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
// A mark remembers the root it found a candidate under, and the value that
// root held then. While the root holds the same value the next flush takes the
// candidate as reachable without walking the graph. A mockery proxy is dropped
// by thousands of statements and is reachable from the container's static
// every time; without the witness each of those statements walked a whole
// laravel application's object graph.
//
// `__destruct` is a magic method php has and upstream's design leaves out.
// Laravel's `Route::resource()` returns a `PendingResourceRegistration` that
// registers its routes in the destructor, so half of a laravel app's routes are
// missing without one; phpx/testdata/classes/destruct.phpt settles the rest.

// phpxDtorState is the destructor bookkeeping of one program: the instances
// still alive in creation order, the same set as a map, the ones the destructor
// already ran on, the instances a dropped reference may have freed, and the
// root each of those was last found reachable under.
type phpxDtorState struct {
	order   []*model.Object
	live    map[*model.Object]struct{}
	done    map[*model.Object]struct{}
	pending []*model.Object
	witness map[*model.Object]phpxWitness
	// walkEpoch counts tracked instances created or destroyed. walkObj and
	// walkArr remember, per container, the epoch when a walk found no tracked
	// instance inside it: while the epoch holds, the subtree is skipped, which
	// is what keeps a note of the application object from walking it again at
	// every frame that copies it to a local.
	walkEpoch int
	walkObj   map[*model.Object]int
	walkArr   map[*model.Array]int
	// markObj, markArr and markCell cache the tracked instances reachable
	// under a container at a version of it. A flush that drops one instance
	// then replays the parts of the graph no write has touched instead of
	// walking them, which is what keeps a suite from walking the whole test
	// framework on every assertion.
	markObj  map[*model.Object]phpxMarkEntry
	markArr  map[*model.Array]phpxMarkEntry
	markCell map[*model.Cell]phpxMarkEntry
	// closing is set while the program's end destroys what is left: every
	// instance is already noted there, so a statement inside a destructor
	// must not flush a candidate out of creation order.
	closing bool
	classes map[string]int
}

// dtors returns the state, creating it on first use.
func (rt *Runtime) dtors() *phpxDtorState {
	if rt.phpxDtors == nil {
		rt.phpxDtors = &phpxDtorState{
			live:    map[*model.Object]struct{}{},
			done:    map[*model.Object]struct{}{},
			witness: map[*model.Object]phpxWitness{},
			walkObj:  map[*model.Object]int{},
			walkArr:  map[*model.Array]int{},
			markObj:  map[*model.Object]phpxMarkEntry{},
			markArr:  map[*model.Array]phpxMarkEntry{},
			markCell: map[*model.Cell]phpxMarkEntry{},
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
	if state.classes == nil {
		state.classes = map[string]int{}
	}
	state.classes[obj.Class.Name]++
	state.live[obj] = struct{}{}
	state.order = append(state.order, obj)
	state.phpxEpochNext()
}

// phpxEpochNext starts a new walk epoch. The negative caches are keyed by
// containers a walk visited, and a key keeps its container reachable; a suite
// that walks millions of them would hold every one, so the maps are dropped
// once they have outgrown what a single test's walk reaches.
func (state *phpxDtorState) phpxEpochNext() {
	state.walkEpoch++
	if len(state.walkObj)+len(state.walkArr) > 1<<15 {
		clear(state.walkObj)
		clear(state.walkArr)
	}
}

// phpxNoteDropped notes every tracked instance reachable from a value whose
// last reference may just have gone away. The flush decides whether it did.
func (rt *Runtime) phpxNoteDropped(v any) {
	if rt.phpxDtors == nil || rt.phpxDtors.closing || v == nil {
		// The program's end destroys every instance in creation order; a
		// walk that collected candidates there would read the whole object
		// store again for each one.
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
		if state.walkObj[x] == state.walkEpoch {
			return
		}
		before := len(state.pending)
		for _, val := range x.Props {
			rt.phpxCollectDropped(val, seen)
		}
		if len(state.pending) == before {
			state.walkObj[x] = state.walkEpoch
		}
	case *model.Array:
		if x == nil {
			return
		}
		if _, ok := seen.arrays[x]; ok {
			return
		}
		seen.arrays[x] = struct{}{}
		if state.walkArr[x] == state.walkEpoch {
			return
		}
		before := len(state.pending)
		x.Range(func(key, val any) bool {
			rt.phpxCollectDropped(key, seen)
			rt.phpxCollectDropped(val, seen)
			return true
		})
		if len(state.pending) == before {
			state.walkArr[x] = state.walkEpoch
		}
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
	// Only a local can lose its last reference when the frame pops: the
	// caller still holds the receiver, the arguments and anything a static or
	// a reference cell names. Noting those walked the caller's whole graph on
	// every iterator call, which is quadratic over a suite.
	for name, val := range scope.vars {
		if name == "this" {
			continue
		}
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
// phpxDtorStats reads PHPX_DTOR_STATS once: a run that sets it prints the
// destructor's counters before the program's instances are destroyed.
var phpxDtorStats = func() func() bool {
	on := os.Getenv("PHPX_DTOR_STATS") != ""
	return func() bool { return on }
}()

var phpxDtorFlushes, phpxDtorNodes, phpxDtorPending, phpxDtorMaxNodes, phpxDtorRootNodes, phpxDtorFrames int

func (rt *Runtime) phpxFlush(extra any) error {
	state := rt.phpxDtors
	if state == nil {
		return nil
	}
	phpxDtorFlushes++
	phpxDtorPending += len(state.pending)
	for pass := 0; pass < 32; pass++ {
		if len(state.pending) == 0 {
			return nil
		}
		pending := state.pending
		state.pending = nil
		need := rt.phpxCandidates(state, pending)
		if len(need) == 0 {
			return nil
		}
		candidates := make(map[*model.Object]bool, len(need))
		for _, obj := range need {
			candidates[obj] = true
		}
		seen, _ := rt.phpxMarkReachable(extra, candidates)
		wanted := map[*model.Object]bool{}
		for _, obj := range need {
			if _, ok := seen.objects[obj]; ok {
				wanted[obj] = true
			}
		}
		witnessed := rt.phpxWitnessFor(wanted)
		progress := false
		for _, obj := range need {
			if w, ok := witnessed[obj]; ok {
				state.witness[obj] = w
			}
			if _, reachable := seen.objects[obj]; reachable {
				continue
			}
			delete(state.witness, obj)
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

// phpxCandidates answers the instances a flush must walk for: the pending ones
// still alive, not already destroyed, and not known reachable from a root that
// still holds what it held when the mark found it there.
func (rt *Runtime) phpxCandidates(state *phpxDtorState, pending []*model.Object) []*model.Object {
	seen := make(map[*model.Object]bool, len(pending))
	need := make([]*model.Object, 0, len(pending))
	for _, obj := range pending {
		if _, done := state.done[obj]; done {
			continue
		}
		if seen[obj] {
			continue
		}
		seen[obj] = true
		if w, ok := state.witness[obj]; ok && rt.phpxWitnessHolds(w) {
			continue
		}
		need = append(need, obj)
	}
	return need
}

// phpxWitness is one root that reached a tracked instance at the last mark,
// with the value the root held then. A later flush reads that root again: while
// it holds the same value the instance is taken as reachable and the graph is
// not walked. The approximation errs one way only - a stale witness keeps an
// instance alive until the program ends and never destroys one early.
type phpxWitness struct {
	kind  phpxWitnessKind
	name  string
	class string
	stat  *model.StaticVar
	value any
}

// phpxWitnessKind names the table a witness reads back from.
type phpxWitnessKind uint8

const (
	phpxWitnessNone phpxWitnessKind = iota
	phpxWitnessGlobal
	phpxWitnessAuto
	phpxWitnessClassStatic
	phpxWitnessFuncStatic
)

// phpxWitnessFor finds a root that reaches each wanted instance, walking each
// root a witness can name until all are found. A root outside those tables - a
// scope, an argument - names nothing a later flush can read back, so an
// instance only they reach gets no witness and is walked again.
func (rt *Runtime) phpxWitnessFor(wanted map[*model.Object]bool) map[*model.Object]phpxWitness {
	found := make(map[*model.Object]phpxWitness, len(wanted))
	if len(wanted) == 0 {
		return found
	}
	left := len(wanted)
	rt.phpxWitnessRoots(func(w phpxWitness, val any) bool {
		if left == 0 {
			return false
		}
		m := newPHPMark(rt.phpxDtors, wanted)
		m.value(val)
		for obj := range wanted {
			if _, ok := found[obj]; ok {
				continue
			}
			if _, hit := m.seen.objects[obj]; hit {
				found[obj] = w
				left--
			}
		}
		return true
	})
	return found
}

// phpxWitnessRoots visits every root a witness can name, with the value it
// holds now. visit answers false to stop.
func (rt *Runtime) phpxWitnessRoots(visit func(phpxWitness, any) bool) {
	for name, val := range rt.globals {
		if !visit(phpxWitness{kind: phpxWitnessGlobal, name: name, value: val}, val) {
			return
		}
	}
	stop := false
	rt.auto.Range(func(name string, val any) {
		if stop {
			return
		}
		if !visit(phpxWitness{kind: phpxWitnessAuto, name: name, value: val}, val) {
			stop = true
		}
	})
	if stop {
		return
	}
	for class, bag := range rt.classStatics {
		for name, val := range bag {
			if !visit(phpxWitness{kind: phpxWitnessClassStatic, class: class, name: name, value: val}, val) {
				return
			}
		}
	}
	for stat, bag := range rt.funcStatics {
		for name, val := range bag {
			if !visit(phpxWitness{kind: phpxWitnessFuncStatic, stat: stat, name: name, value: val}, val) {
				return
			}
		}
	}
}

// phpxWitnessHolds reports whether the root a witness names still holds the
// value it held when the mark found the instance beneath it.
func (rt *Runtime) phpxWitnessHolds(w phpxWitness) bool {
	var current any
	var ok bool
	switch w.kind {
	case phpxWitnessGlobal:
		current, ok = rt.globals[w.name]
	case phpxWitnessAuto:
		current, ok = rt.auto.Lookup(w.name)
	case phpxWitnessClassStatic:
		if bag, found := rt.classStatics[w.class]; found {
			current, ok = bag[w.name]
		}
	case phpxWitnessFuncStatic:
		if bag, found := rt.funcStatics[w.stat]; found {
			current, ok = bag[w.name]
		}
	}
	return ok && phpxSameValue(current, w.value)
}

// phpxSameValue reports two root values as one reference. A pair it cannot
// compare answers false, so the graph is walked again rather than trusted.
func phpxSameValue(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case *model.Object:
		y, ok := b.(*model.Object)
		return ok && x == y
	case *model.Array:
		y, ok := b.(*model.Array)
		return ok && x == y
	case *model.Cell:
		y, ok := b.(*model.Cell)
		return ok && x == y
	}
	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	if !ra.IsValid() || !rb.IsValid() || ra.Kind() != rb.Kind() {
		return false
	}
	switch ra.Kind() {
	case reflect.Slice:
		return ra.Pointer() == rb.Pointer()
	case reflect.Map:
		return ra.UnsafePointer() == rb.UnsafePointer()
	}
	return false
}

// phpxDestroyLive destroys every instance still alive when the program ends,
// in reverse creation order, as php tears its object store down after the
// shutdown functions.
func (rt *Runtime) phpxDestroyLive() {
	if phpxDtorStats() {
		fmt.Fprintf(os.Stderr, "dtor: flushes=%d nodes=%d maxnodes=%d pending=%d rootnodes=%d frames=%d hit=%d miss=%d arrwalk=%d taint=%d\n", phpxDtorFlushes, phpxDtorNodes, phpxDtorMaxNodes, phpxDtorPending, phpxDtorRootNodes, phpxDtorFrames, phpxMarkHit, phpxMarkMiss, phpxMarkArrWalk, phpxMarkTaint)
	}
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
	delete(state.witness, obj)
	state.phpxEpochNext()
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

// phpxMarkEntry caches what one container reaches: the tracked instances under
// it and the version of the container the walk saw. A later walk that meets the
// same version replays the list instead of descending.
type phpxMarkEntry struct {
	version uint32
	tracked []*model.Object
}

// phpxMark is one reachability walk: the containers it entered, the tracked
// instances under the container it is in, and the candidates it has yet to
// find.
type phpxMark struct {
	state *phpxDtorState
	seen  *phpMarkSet
	// stack holds the containers being walked, innermost last.
	stack []phpxMarkFrame
	// groups holds, per cycle root, the containers the cycle it closed passed
	// through. Every container in one strongly connected component reaches the
	// same tracked instances, so one list serves for all of them.
	groups map[int][]any
	// tracked collects the tracked instances under the container being walked.
	tracked []*model.Object
	left    map[*model.Object]bool
}

func newPHPMark(state *phpxDtorState, candidates map[*model.Object]bool) *phpxMark {
	left := make(map[*model.Object]bool, len(candidates))
	for obj := range candidates {
		left[obj] = true
	}
	return &phpxMark{state: state, seen: newPHPMarkSet(), left: left}
}

// record notes a tracked instance the walk reached. left is the candidate set,
// so the record is also the hit test a flush reads.
func (m *phpxMark) record(obj *model.Object) {
	m.seen.objects[obj] = struct{}{}
	delete(m.left, obj)
}

// replay takes a cached container's tracked instances into this walk.
func (m *phpxMark) replay(tracked []*model.Object) {
	for _, obj := range tracked {
		if _, ok := m.seen.objects[obj]; ok {
			continue
		}
		m.tracked = append(m.tracked, obj)
		m.record(obj)
	}
}

// phpxMarkFrame is one container on the walk's path.
type phpxMarkFrame struct {
	container any
	before    int
	// group is the stack index of the outermost container of the cycle this
	// frame belongs to, or -1.
	group int
}

var phpxMarkHit, phpxMarkMiss, phpxMarkArrWalk, phpxMarkTaint int

// enter records a container the walk descends into.
func (m *phpxMark) enter(c any) {
	m.stack = append(m.stack, phpxMarkFrame{container: c, before: len(m.tracked), group: -1})
}

// backEdge notes a walk that met a container it is already inside: the path
// from that container down to here is one cycle, and every container on it
// reaches what the outermost one reaches.
func (m *phpxMark) backEdge(c any) {
	for i := len(m.stack) - 1; i >= 0; i-- {
		if m.stack[i].container != c {
			continue
		}
		phpxMarkTaint++
		if m.groups == nil {
			m.groups = map[int][]any{}
		}
		for j := i; j < len(m.stack); j++ {
			if m.stack[j].group == i {
				continue
			}
			m.stack[j].group = i
			m.groups[i] = append(m.groups[i], m.stack[j].container)
		}
		return
	}
}

// leave pops the container. A container the walk reached without a cycle
// answers its own tracked list. One inside a cycle answers the cycle's members
// at the root and nothing for the frames under it; the root's list is theirs.
func (m *phpxMark) leave(c any) (members []any, tracked []*model.Object, cycle bool) {
	frame := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	if frame.group < 0 {
		return nil, m.tracked[frame.before:], false
	}
	if frame.group != len(m.stack) {
		return nil, nil, true
	}
	members = m.groups[frame.group]
	delete(m.groups, frame.group)
	return members, m.tracked[frame.before:], true
}

// storeGroup caches the cycle's tracked list for the root and every member.
func (m *phpxMark) storeGroup(c any, members []any, tracked []*model.Object) {
	m.state.storeMark(c, tracked)
	for _, member := range members {
		m.state.storeMark(member, tracked)
	}
}

// value marks v and everything it reaches.
func (m *phpxMark) value(v any) {
	switch x := v.(type) {
	case *model.Object:
		if x == nil {
			return
		}
		if entry, ok := m.state.markObj[x]; ok && entry.version == x.Version() {
			phpxMarkHit++
			m.replay(entry.tracked)
			return
		}
		phpxMarkMiss++
		if _, ok := m.seen.objects[x]; ok {
			m.backEdge(x)
			return
		}
		m.seen.objects[x] = struct{}{}
		m.enter(x)
		if _, live := m.state.live[x]; live {
			m.tracked = append(m.tracked, x)
			delete(m.left, x)
		}
		for _, val := range x.Props {
			m.value(val)
		}
		members, tracked, cycle := m.leave(x)
		if cycle {
			m.storeGroup(x, members, tracked)
			return
		}
		m.state.storeMarkObj(x, tracked)
	case *model.Array:
		if x == nil {
			return
		}
		if entry, ok := m.state.markArr[x]; ok && entry.version == x.Version() {
			phpxMarkHit++
			m.replay(entry.tracked)
			return
		}
		phpxMarkMiss++
		if _, ok := m.seen.arrays[x]; ok {
			m.backEdge(x)
			return
		}
		m.seen.arrays[x] = struct{}{}
		m.enter(x)
		phpxMarkArrWalk++
		x.Range(func(key, val any) bool {
			m.value(key)
			m.value(val)
			return true
		})
		members, tracked, cycle := m.leave(x)
		if cycle {
			m.storeGroup(x, members, tracked)
			return
		}
		m.state.storeMarkArr(x, tracked)
	case *model.Cell:
		if x == nil {
			return
		}
		if entry, ok := m.state.markCell[x]; ok && entry.version == x.Version() {
			m.replay(entry.tracked)
			return
		}
		if _, ok := m.seen.cells[x]; ok {
			m.backEdge(x)
			return
		}
		m.seen.cells[x] = struct{}{}
		m.enter(x)
		m.value(x.V)
		members, tracked, cycle := m.leave(x)
		if cycle {
			m.storeGroup(x, members, tracked)
			return
		}
		m.state.storeMarkCell(x, tracked)
	default:
		phpxEachValue(v, m.value)
	}
}

// markCaches is the number of cached entries past which the caches are dropped.
// A key keeps its container reachable, so a suite that walks millions of them
// would hold every one; the cap keeps what one test's walk reaches.
const markCaches = 1 << 16

// storeMark caches a container's tracked instances at the version it was
// walked at, whatever kind of container it is.
func (state *phpxDtorState) storeMark(c any, tracked []*model.Object) {
	switch x := c.(type) {
	case *model.Object:
		state.storeMarkObj(x, tracked)
	case *model.Array:
		state.storeMarkArr(x, tracked)
	case *model.Cell:
		state.storeMarkCell(x, tracked)
	}
}

// storeMarkObj caches the tracked instances under c at its current version.
func (state *phpxDtorState) storeMarkObj(c *model.Object, tracked []*model.Object) {
	state.markObj[c] = phpxMarkEntry{version: c.Version(), tracked: copyTracked(tracked)}
	state.capMarks()
}

// storeMarkArr caches the tracked instances under c at its current version.
func (state *phpxDtorState) storeMarkArr(c *model.Array, tracked []*model.Object) {
	state.markArr[c] = phpxMarkEntry{version: c.Version(), tracked: copyTracked(tracked)}
	state.capMarks()
}

// storeMarkCell caches the tracked instances under c at its current version.
func (state *phpxDtorState) storeMarkCell(c *model.Cell, tracked []*model.Object) {
	state.markCell[c] = phpxMarkEntry{version: c.Version(), tracked: copyTracked(tracked)}
	state.capMarks()
}

// capMarks drops the caches once they outgrow what one walk reaches.
func (state *phpxDtorState) capMarks() {
	if len(state.markObj)+len(state.markArr)+len(state.markCell) <= markCaches {
		return
	}
	clear(state.markObj)
	clear(state.markArr)
	clear(state.markCell)
}

// copyTracked answers a list the cache can hold without aliasing the walk's
// growing slice.
func copyTracked(tracked []*model.Object) []*model.Object {
	if len(tracked) == 0 {
		return nil
	}
	out := make([]*model.Object, len(tracked))
	copy(out, tracked)
	return out
}

// phpxMarkReachable collects every object the live roots reach: the frames,
// the globals and auto globals, the class and function statics, the bytecode
// engine's live values and the statement's own result. It stops once every
// candidate is marked, since a mark that finds them all destroys nothing.
func (rt *Runtime) phpxMarkReachable(extra any, candidates map[*model.Object]bool) (*phpMarkSet, map[*model.Object]phpxWitness) {
	m := newPHPMark(rt.phpxDtors, candidates)
	defer func() {
		n := len(m.seen.objects) + len(m.seen.arrays) + len(m.seen.cells)
		phpxDtorNodes += n
		if n > phpxDtorMaxNodes {
			phpxDtorMaxNodes = n
		}
	}()
	mark := func(v any) bool {
		m.tracked = m.tracked[:0]
		before := len(m.seen.objects) + len(m.seen.arrays) + len(m.seen.cells)
		m.value(v)
		phpxDtorRootNodes += len(m.seen.objects) + len(m.seen.arrays) + len(m.seen.cells) - before
		return len(candidates) > 0 && len(m.left) == 0
	}
	markAll := func(values map[string]any) bool {
		for _, val := range values {
			if mark(val) {
				return true
			}
		}
		return false
	}
	phpxDtorFrames++
	for _, scope := range rt.frames {
		if markAll(scope.vars) {
			return m.seen, nil
		}
		for _, cell := range scope.refs {
			if mark(cell) {
				return m.seen, nil
			}
		}
		for _, bag := range scope.statics {
			if markAll(bag) {
				return m.seen, nil
			}
		}
		for _, val := range scope.args {
			if mark(val) {
				return m.seen, nil
			}
		}
		for _, val := range scope.deferred {
			if mark(val) {
				return m.seen, nil
			}
		}
	}
	if markAll(rt.globals) {
		return m.seen, nil
	}
	stop := false
	rt.auto.Range(func(_ string, val any) {
		if stop {
			return
		}
		stop = mark(val)
	})
	if stop {
		return m.seen, nil
	}
	for _, bag := range rt.classStatics {
		if markAll(bag) {
			return m.seen, nil
		}
	}
	for _, bag := range rt.funcStatics {
		if markAll(bag) {
			return m.seen, nil
		}
	}
	for _, walk := range rt.vmWalkers {
		walk(func(v any) {
			if stop {
				return
			}
			stop = mark(v)
		})
		if stop {
			return m.seen, nil
		}
	}
	mark(extra)
	return m.seen, nil
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
