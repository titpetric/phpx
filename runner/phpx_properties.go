//go:build phpx

package runner

import (
	"fmt"
	"github.com/titpetric/phpx/modelutil"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
)

// phpxShadowSlot is one private property an ancestor declared, which a
// descendant's declaration of the same name hides from every scope but that
// ancestor's own. php keeps the two in two slots; the fork keeps the hidden
// one under php's mangled key beside the property the descendant sees.
type phpxShadowSlot struct {
	owner string
	key   string
	field model.Field
}

// phpxPropInfo is what the visibility check needs of one non-public property:
// its visibility, the class that declared it, and the hidden slots an ancestor
// left beside it.
type phpxPropInfo struct {
	visibility string
	owner      string
	readonly   bool
	decl       model.Field
	shadows    []phpxShadowSlot
}

// phpxShadowKey is the key php gives the private property owner in an object
// whose class redeclares the name; `(array)` reads the slot back under it.
func phpxShadowKey(owner, name string) string { return "\x00" + owner + "\x00" + name }

// phpxIsShadowKey reports whether name is a hidden slot's key rather than a
// property name a script wrote.
func phpxIsShadowKey(name string) bool { return len(name) > 0 && name[0] == 0 }

// phpxHiddenProps holds, per class, the non-public properties it carries. A
// class with none has no entry, so a read of an ordinary object costs one map
// miss. It is keyed by the hoisted class and built once at hoist.
var phpxHiddenProps sync.Map // *model.Class -> map[string]phpxPropInfo

// phpxPropHidden is the error a refused access raises. Quiet marks it for
// isset() and empty(), which answer for a property out of reach without one.
type phpxPropHidden struct{ msg string }

func (e *phpxPropHidden) Error() string { return e.msg }

// ThrowableClass makes the refusal an Error, which a catch clause matches.
func (e *phpxPropHidden) ThrowableClass() string { return "Error" }

// Quiet reports that isset and empty swallow this error.
func (e *phpxPropHidden) Quiet() bool { return true }

// phpxBindProps records the visibility and the declaring class of every
// non-public property of c. Called once c has merged its parent's fields:
// own declarations (promoted and trait ones included) belong to c, the rest
// keep the owner the parent recorded.
func phpxBindProps(c *model.Class, own []model.Field, parent *model.Class) {
	info := map[string]phpxPropInfo{}
	if parent != nil {
		if inherited, ok := phpxHiddenProps.Load(parent); ok {
			for name, p := range inherited.(map[string]phpxPropInfo) {
				info[name] = p
			}
		}
	}
	for _, f := range c.Fields {
		declared := phpxHasField(own, f.Name)
		if !declared {
			continue
		}
		prev, had := info[f.Name]
		next := phpxPropInfo{visibility: f.Visibility, owner: c.Name, readonly: f.Readonly, decl: f}
		if had {
			// A private property of an ancestor keeps its own slot when the
			// child declares the name again; php reads it from the ancestor's
			// scope alone and `(array)` shows the mangled key.
			next.shadows = append(next.shadows, prev.shadows...)
			if prev.visibility == "private" {
				next.shadows = append(next.shadows, phpxShadowSlot{
					owner: prev.owner,
					key:   phpxShadowKey(prev.owner, f.Name),
					field: prev.decl,
				})
			}
		}
		if next.visibility == "public" && !next.readonly && len(next.shadows) == 0 {
			delete(info, f.Name)
			continue
		}
		info[f.Name] = next
	}
	if len(info) > 0 {
		phpxHiddenProps.Store(c, info)
	}
}

// phpxPropVerdict says how an access to obj->name from scope plays out: nil
// when it is allowed, hidden when php sees no such property (a private one of
// an ancestor), and an error when php refuses it.
func (rt *Runtime) phpxPropVerdict(obj *model.Object, name string, scope *Scope) (hidden bool, err error) {
	if obj.Class == nil {
		return false, nil
	}
	table, ok := phpxHiddenProps.Load(obj.Class)
	if !ok {
		return false, nil
	}
	p, ok := table.(map[string]phpxPropInfo)[name]
	if !ok {
		return false, nil
	}
	from := phpxCallerClass(scope)
	if p.visibility == "" || p.visibility == "public" {
		return false, nil
	}
	if p.visibility == "protected" {
		if from != "" && rt.phpxMayCall("protected", p.owner, from) {
			return false, nil
		}
	} else {
		if from == p.owner {
			return false, nil
		}
		for _, slot := range p.shadows {
			if strings.EqualFold(from, slot.owner) {
				return false, nil
			}
		}
		if p.owner != obj.Class.Name {
			return true, nil
		}
	}
	return false, &phpxPropHidden{msg: fmt.Sprintf("Cannot access %s property %s::$%s", p.visibility, obj.Class.Name, name)}
}

// phpxSlotKey answers the storage key an access to obj->name uses. When the
// running body belongs to an ancestor whose private property the object's
// class declared again, php reads that ancestor's own slot; the fork keeps it
// under the mangled key beside the visible property.
func (rt *Runtime) phpxSlotKey(obj *model.Object, name string, scope *Scope) string {
	if obj == nil || obj.Class == nil || phpxIsShadowKey(name) {
		return name
	}
	from := phpxCallerClass(scope)
	if from == "" {
		return name
	}
	held, ok := phpxHiddenProps.Load(obj.Class)
	if !ok {
		return name
	}
	p, ok := held.(map[string]phpxPropInfo)[name]
	if !ok || len(p.shadows) == 0 {
		return name
	}
	for _, slot := range p.shadows {
		if strings.EqualFold(slot.owner, from) {
			return slot.key
		}
	}
	return name
}

// phpxSeedShadow writes each hidden slot's default into a new object. The
// declaration's default evaluates in the scope of the class that wrote it, and
// a property with no default starts null, as php initializes it.
func (rt *Runtime) phpxSeedShadow(obj *model.Object, c *model.Class, caller *Scope) error {
	if c == nil {
		return nil
	}
	held, ok := phpxHiddenProps.Load(c)
	if !ok {
		return nil
	}
	for _, p := range held.(map[string]phpxPropInfo) {
		for _, slot := range p.shadows {
			var def any
			if slot.field.Default != nil {
				owner := c
				if oc, ok := rt.lookupClass(slot.owner); ok {
					owner = oc
				}
				v, err := rt.Eval(slot.field.Default, rt.classDefaultScope(owner, caller))
				if err != nil {
					return err
				}
				def = v
			}
			obj.SetProp(slot.key, def)
		}
	}
	return nil
}

// phpxReadable filters the properties of obj to those scope may read, in the
// object's order. A hidden slot stays out unless scope is the ancestor that
// declared it, which then reads it under the plain name, as php does.
func (rt *Runtime) phpxReadable(obj *model.Object, scope *Scope, yield func(name string, v any) bool) {
	info, _ := phpxHiddenProps.Load(obj.Class)
	table, _ := info.(map[string]phpxPropInfo)
	from := phpxCallerClass(scope)
	obj.Range(func(name string, v any) bool {
		if phpxIsShadowKey(name) {
			return true
		}
		if p, ok := table[name]; ok && from != "" {
			for _, slot := range p.shadows {
				if !strings.EqualFold(slot.owner, from) {
					continue
				}
				// The ancestor reads its own slot and never the visible one:
				// an unset slot is a property it does not see at all.
				v, ok := obj.Prop(slot.key)
				if !ok {
					return true
				}
				if hidden, err := rt.phpxPropVerdict(obj, name, scope); hidden || err != nil {
					return true
				}
				return yield(name, v)
			}
		}
		if hidden, err := rt.phpxPropVerdict(obj, name, scope); hidden || err != nil {
			return true
		}
		return yield(name, v)
	})
}

// PhpxCallerScope is the frame that called the running binding.
func (rt *Runtime) PhpxCallerScope() *Scope {
	if len(rt.frames) == 0 {
		return nil
	}
	return rt.frames[len(rt.frames)-1]
}

// PhpxVisibleProps lists the properties of obj the calling body may read.
func (rt *Runtime) PhpxVisibleProps(obj *model.Object, yield func(name string, v any) bool) {
	rt.phpxReadable(obj, rt.PhpxCallerScope(), yield)
}

// phpxPropAccess refuses a write or an unset php refuses. A private property
// of an ancestor is out of sight and the write lands on the shared slot, the
// one divergence from php, which would make a dynamic property.
func (rt *Runtime) phpxPropAccess(obj *model.Object, name string, scope *Scope) error {
	hidden, err := rt.phpxPropVerdict(obj, name, scope)
	if err != nil || hidden || obj.Class == nil {
		return err
	}
	return phpxReadonlyWrite(obj, name, scope)
}

// phpxReadonlyWrite refuses a write to a readonly property that is already
// set, and the first write from outside the declaring class. Both raise an
// Error, as php does.
func phpxReadonlyWrite(obj *model.Object, name string, scope *Scope) error {
	table, ok := phpxHiddenProps.Load(obj.Class)
	if !ok {
		return nil
	}
	p, ok := table.(map[string]phpxPropInfo)[name]
	if !ok || !p.readonly {
		return nil
	}
	if _, set := obj.Prop(name); set {
		return &phpxPropHidden{msg: fmt.Sprintf("Cannot modify readonly property %s::$%s", p.owner, name)}
	}
	from := phpxCallerClass(scope)
	if from == p.owner {
		return nil
	}
	where := "global scope"
	if from != "" {
		where = "scope " + from
	}
	return &phpxPropHidden{msg: fmt.Sprintf("Cannot modify protected(set) readonly property %s::$%s from %s", p.owner, name, where)}
}

// phpxUninitialized reports that f has no value until the constructor writes
// it: a readonly property is absent from the object, as php's uninitialized
// state is.
func phpxUninitialized(f model.Field) bool { return f.Readonly }

// phpxEachProp visits the properties of obj the body running in scope may see.
func (rt *Runtime) phpxEachProp(obj *model.Object, scope *Scope, yield func(name string, v any) bool) {
	rt.phpxReadable(obj, scope, yield)
}

// phpxPropUnset refuses an unset php refuses: the visibility rules of a write,
// and any readonly property that holds a value.
func (rt *Runtime) phpxPropUnset(obj *model.Object, name string, scope *Scope) error {
	hidden, err := rt.phpxPropVerdict(obj, name, scope)
	if err != nil || hidden || obj.Class == nil {
		return err
	}
	if table, ok := phpxHiddenProps.Load(obj.Class); ok {
		if p, ok := table.(map[string]phpxPropInfo)[name]; ok && p.readonly {
			if _, set := obj.Prop(name); set || phpxCallerClass(scope) != p.owner {
				return &phpxPropHidden{msg: fmt.Sprintf("Cannot unset readonly property %s::$%s", p.owner, name)}
			}
		}
	}
	return nil
}

// phpxIndirectWrite refuses `$o->list[] = v` and `$o->map['k']['j'] = v` when
// the property is a readonly array that holds a value. Only a variable base is
// looked at, so the base is not evaluated twice.
func (rt *Runtime) phpxIndirectWrite(base model.Expr, scope *Scope) error {
	for {
		idx, ok := modelutil.UnwrapParenthesized(base).(*model.Index)
		if !ok {
			break
		}
		base = idx.Base
	}
	pa, ok := modelutil.UnwrapParenthesized(base).(*model.PropAccess)
	if !ok {
		return nil
	}
	v, ok := pa.Base.(*model.Var)
	if !ok {
		return nil
	}
	holder, _ := scope.Get(v.Name)
	obj, ok := holder.(*model.Object)
	if !ok || obj.Class == nil {
		return nil
	}
	table, ok := phpxHiddenProps.Load(obj.Class)
	if !ok {
		return nil
	}
	name, err := rt.propName(pa, scope)
	if err != nil {
		return nil
	}
	if p, ok := table.(map[string]phpxPropInfo)[name]; ok && p.readonly {
		if cur, set := obj.Prop(rt.phpxSlotKey(obj, name, scope)); set {
			if _, isArray := cur.(*model.Array); isArray {
				return &phpxPropHidden{msg: fmt.Sprintf("Cannot indirectly modify readonly property %s::$%s", p.owner, name)}
			}
		}
	}
	return nil
}

// phpxUninitializedRead refuses a read of a readonly property the object does
// not hold yet, which is php's uninitialized typed property.
func (rt *Runtime) phpxUninitializedRead(obj *model.Object, name string, scope *Scope) error {
	if obj.Class == nil {
		return nil
	}
	table, ok := phpxHiddenProps.Load(obj.Class)
	if !ok {
		return nil
	}
	if p, ok := table.(map[string]phpxPropInfo)[name]; ok && p.readonly {
		if _, set := obj.Prop(rt.phpxSlotKey(obj, name, scope)); set {
			return nil
		}
		return &phpxPropHidden{msg: fmt.Sprintf("Typed property %s::$%s must not be accessed before initialization", p.owner, name)}
	}
	return nil
}

// phpxCoerceProperty stores a Stringable as its string for a property declared string.
func (rt *Runtime) phpxCoerceProperty(obj *model.Object, name string, v any) any {
	o, ok := v.(*model.Object)
	if !ok || obj.Class == nil {
		return v
	}
	switch phpxDeclaredType(obj.Class, name) {
	case "string", "?string":
		if s, err := phpxStringOf(o); err == nil {
			return s
		}
	}
	return v
}

// phpxDeclaredType answers a property's declared type, as written.
func phpxDeclaredType(c *model.Class, name string) string {
	for _, f := range c.Fields {
		if strings.EqualFold(f.Name, name) {
			if f.Resolved != "" {
				return f.Resolved
			}
			return f.Type
		}
	}
	return ""
}
