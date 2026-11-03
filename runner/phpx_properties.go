//go:build phpx

package runner

import (
	"fmt"
	"sync"

	"github.com/titpetric/phpx/model"
)

// phpxPropInfo is what the visibility check needs of one non-public property:
// its visibility and the class that declared it.
type phpxPropInfo struct {
	visibility string
	owner      string
}

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
		if f.Visibility == "private" || f.Visibility == "protected" {
			info[f.Name] = phpxPropInfo{visibility: f.Visibility, owner: c.Name}
		} else {
			delete(info, f.Name)
		}
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
	if p.visibility == "protected" {
		if from != "" && rt.phpxMayCall("protected", p.owner, from) {
			return false, nil
		}
	} else {
		if from == p.owner {
			return false, nil
		}
		if p.owner != obj.Class.Name {
			return true, nil
		}
	}
	return false, &phpxPropHidden{msg: fmt.Sprintf("Cannot access %s property %s::$%s", p.visibility, obj.Class.Name, name)}
}

// phpxReadable filters the properties of obj to those scope may read, in the
// object's order.
func (rt *Runtime) phpxReadable(obj *model.Object, scope *Scope, yield func(name string, v any) bool) {
	obj.Range(func(name string, v any) bool {
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
	_, err := rt.phpxPropVerdict(obj, name, scope)
	return err
}

// phpxEachProp visits the properties of obj the body running in scope may see.
func (rt *Runtime) phpxEachProp(obj *model.Object, scope *Scope, yield func(name string, v any) bool) {
	rt.phpxReadable(obj, scope, yield)
}
