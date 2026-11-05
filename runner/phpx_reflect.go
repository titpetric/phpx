//go:build phpx

package runner

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
)

// phpxDecls keeps the declaration a hoisted class was built from, which carries
// what model.Class drops: the order methods were written in, `final`, and the
// parameter lists Reflection reads.
var phpxDecls sync.Map // *model.Class -> *model.ClassDecl

// phpxRecordDecl notes cd as the declaration behind c.
func phpxRecordDecl(c *model.Class, cd *model.ClassDecl) { phpxDecls.Store(c, cd) }

func phpxDeclOf(c *model.Class) *model.ClassDecl {
	if v, ok := phpxDecls.Load(c); ok {
		return v.(*model.ClassDecl)
	}
	return nil
}

func phpxMap(pairs ...any) *model.Array {
	out := model.NewArray()
	for i := 0; i+1 < len(pairs); i += 2 {
		out.Set(pairs[i], pairs[i+1])
	}
	return out
}

func phpxList[T any](items []T, f func(T) any) *model.Array {
	out := model.NewArray()
	for _, it := range items {
		out.Append(f(it))
	}
	return out
}

// PhpxReflectClass describes a class, an interface or an object's class as the
// array the Reflection prelude reads, or false when no declaration answers.
func (rt *Runtime) PhpxReflectClass(v any, autoload bool) any {
	var name string
	switch t := v.(type) {
	case *model.Object:
		if t.Class == nil {
			return false
		}
		name = t.Class.Name
	case string:
		name = strings.TrimPrefix(t, "\\")
	default:
		return false
	}
	if ok, _ := rt.ClassExists(name, autoload); ok {
		if c, found := rt.lookupClass(name); found {
			return rt.phpxDescribeClass(c)
		}
	}
	if ok, _ := rt.InterfaceExists(name, autoload); ok {
		if id, found := rt.lookupInterface(name); found {
			return rt.phpxDescribeInterface(id)
		}
	}
	if _, ok := rt.lookupConstructor(name); ok {
		short := name
		if i := strings.LastIndexByte(name, '\\'); i >= 0 {
			short = name[i+1:]
		}
		return phpxMap("name", name, "short", short, "namespace", "", "abstract", false, "final", false,
			"interface", false, "internal", true, "parent", nil, "interfaces", model.NewArray(),
			"file", false, "methods", model.NewArray(), "props", model.NewArray(), "consts", model.NewArray())
	}
	return false
}

func phpxSplitName(name string) (short, ns string) {
	if i := strings.LastIndexByte(name, '\\'); i >= 0 {
		return name[i+1:], name[:i]
	}
	return name, ""
}

func phpxParamInfo(fd *model.FuncDecl) *model.Array {
	out := model.NewArray()
	for i, p := range fd.Params {
		var typ any
		if p.Type != "" {
			typ = phpxPrefer(p.Resolved, p.Type)
		}
		out.Append(phpxMap(
			"name", p.Name, "position", int64(i), "type", typ,
			"hasDefault", p.Default != nil, "variadic", p.Variadic, "byRef", p.ByRef,
			"promoted", p.Modifiers != "",
		))
	}
	return out
}

func phpxMethodInfo(fd *model.FuncDecl, owner string, abstract bool) *model.Array {
	vis := fd.Visibility
	if vis == "" {
		vis = "public"
	}
	var ret any
	if fd.ReturnType != "" {
		ret = phpxPrefer(fd.ResolvedReturn, fd.ReturnType)
	}
	var file any = false
	if fd.Filename != "" {
		file = fd.Filename
	}
	return phpxMap("name", fd.Name, "class", owner, "visibility", vis, "static", fd.Static,
		"abstract", abstract || fd.Abstract, "params", phpxParamInfo(fd), "returnType", ret, "file", file)
}

func (rt *Runtime) phpxDescribeClass(c *model.Class) *model.Array {
	short, ns := phpxSplitName(c.Name)
	cd := phpxDeclOf(c)
	final, file, readonly := false, any(false), false
	methods, seen := model.NewArray(), map[string]bool{}
	addDecl := func(decl *model.ClassDecl, owner string) {
		for _, m := range decl.Methods {
			key := strings.ToLower(m.Name)
			if seen[key] {
				continue
			}
			seen[key] = true
			methods.Set(m.Name, phpxMethodInfo(m, owner, false))
			if file == false && m.Filename != "" {
				file = m.Filename
			}
		}
	}
	if cd != nil {
		final, readonly = cd.Final, cd.Readonly
		addDecl(cd, c.Name)
	}
	for _, anc := range c.Ancestors {
		if pc, ok := rt.lookupClass(anc); ok {
			if pd := phpxDeclOf(pc); pd != nil {
				addDecl(pd, pc.Name)
			}
		}
	}
	// What a trait contributed sits in the flattened table and not in a decl.
	var rest []string
	for name := range c.Methods {
		if !seen[strings.ToLower(name)] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	for _, name := range rest {
		owner := c.Name
		if o, ok := c.MethodOwner[name]; ok {
			owner = o
		}
		methods.Set(name, phpxMethodInfo(c.Methods[name], owner, false))
	}

	props := model.NewArray()
	for _, f := range c.Fields {
		props.Set(f.Name, rt.phpxPropInfo(c, f, false))
	}
	for _, f := range c.Statics {
		props.Set(f.Name, rt.phpxPropInfo(c, f, true))
	}
	consts := model.NewArray()
	for _, f := range c.Consts {
		consts.Set(f.Name, f.Name)
	}
	var parent any
	if len(c.Ancestors) > 0 {
		parent = c.Ancestors[0]
		if pc, ok := rt.lookupClass(c.Ancestors[0]); ok {
			parent = pc.Name
		}
	}
	traits := model.NewArray()
	if cd != nil {
		for _, t := range cd.Traits {
			traits.Append(strings.TrimPrefix(t, "\\"))
		}
	}
	ifaces := model.NewArray()
	if v, ok := rt.PhpxClassImplements(c.Name, true).(*model.Array); ok {
		v.Range(func(k, _ any) bool { ifaces.Append(k); return true })
	}
	return phpxMap("name", c.Name, "short", short, "namespace", ns, "abstract", c.Abstract, "final", final,
		"interface", false, "internal", false, "readonly", readonly, "parent", parent, "interfaces", ifaces, "traits", traits,
		"file", file, "methods", methods, "props", props, "consts", consts)
}

func (rt *Runtime) phpxPropInfo(c *model.Class, f model.Field, static bool) *model.Array {
	vis := f.Visibility
	if vis == "" {
		vis = "public"
	}
	var typ any
	if f.Type != "" {
		typ = phpxPrefer(f.Resolved, f.Type)
	}
	return phpxMap("name", f.Name, "class", c.Name, "visibility", vis, "static", static,
		"readonly", f.Readonly, "type", typ, "hasDefault", f.Default != nil)
}

func (rt *Runtime) phpxDescribeInterface(id *model.InterfaceDecl) *model.Array {
	short, ns := phpxSplitName(id.Name)
	methods := model.NewArray()
	for _, m := range id.Methods {
		methods.Set(m.Name, phpxMethodInfo(m, id.Name, true))
	}
	consts := model.NewArray()
	for _, f := range id.Consts {
		consts.Set(f.Name, f.Name)
	}
	return phpxMap("name", id.Name, "short", short, "namespace", ns, "abstract", true, "final", false,
		"interface", true, "internal", false, "parent", nil,
		"interfaces", phpxList(rt.phpxInterfaceChain(id.Name, nil), func(s string) any { return s }),
		"traits", model.NewArray(), "file", false, "methods", methods, "props", model.NewArray(), "consts", consts)
}

// PhpxNewWithoutConstructor builds an instance of class with its declared
// defaults and no constructor call, as ReflectionClass does.
func (rt *Runtime) PhpxNewWithoutConstructor(class string) (any, error) {
	class = strings.TrimPrefix(class, "\\")
	if ok, err := rt.ClassExists(class, true); !ok {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Class \"%s\" does not exist", class)
	}
	c, ok := rt.lookupClass(class)
	if !ok {
		return nil, fmt.Errorf("Class \"%s\" does not exist", class)
	}
	if err := phpxInstantiable(c); err != nil {
		return nil, err
	}
	scope := rt.newScope()
	obj := model.NewObject(c)
	defaults := rt.classDefaultScope(c, scope)
	for _, f := range c.Fields {
		if phpxUninitialized(f) {
			continue
		}
		var def any
		if f.Default != nil {
			v, err := rt.Eval(f.Default, defaults)
			if err != nil {
				return nil, err
			}
			def = v
		}
		obj.SetProp(f.Name, def)
	}
	return obj, nil
}

// PhpxParamDefault evaluates the default of a parameter of a method of class,
// or of the function when class is empty.
func (rt *Runtime) PhpxParamDefault(class, method string, index int) (any, error) {
	var fd *model.FuncDecl
	scope := rt.newScope()
	if class != "" {
		name := strings.TrimPrefix(class, "\\")
		if c, ok := rt.lookupClass(name); ok {
			for mname, m := range c.Methods {
				if strings.EqualFold(mname, method) {
					fd = m
				}
			}
			scope = rt.classDefaultScope(c, scope)
		} else if id, ok := rt.lookupInterface(name); ok {
			for _, m := range id.Methods {
				if strings.EqualFold(m.Name, method) {
					fd = m
				}
			}
		} else {
			return nil, fmt.Errorf("Class \"%s\" does not exist", class)
		}
	}
	if fd == nil || index < 0 || index >= len(fd.Params) || fd.Params[index].Default == nil {
		return nil, fmt.Errorf("Internal error: Failed to retrieve the default value")
	}
	return rt.Eval(fd.Params[index].Default, scope)
}

// PhpxGetProp reads a property whatever its visibility.
func PhpxGetProp(obj *model.Object, name string) any { return model.Deref(obj.Props[name]) }

// PhpxSetProp writes a property whatever its visibility.
func PhpxSetProp(obj *model.Object, name string, v any) { obj.SetProp(name, v) }

// phpxFuncDecls maps a declared function's name to its declaration.
var phpxFuncDecls sync.Map // lower-cased name -> *model.FuncDecl

// phpxRecordFunc notes the declaration behind a user function.
func phpxRecordFunc(fd *model.FuncDecl) { phpxFuncDecls.Store(strings.ToLower(fd.Name), fd) }

// PhpxFuncInfo describes a user function or a closure for ReflectionFunction,
// or answers false for a name no declaration holds.
func (rt *Runtime) PhpxFuncInfo(f any) any {
	var fd *model.FuncDecl
	name := "{closure}"
	switch t := f.(type) {
	case string:
		if v, ok := phpxFuncDecls.Load(strings.ToLower(strings.TrimPrefix(t, "\\"))); ok {
			fd = v.(*model.FuncDecl)
			name = fd.Name
		}
	case *Callable:
		switch {
		case t.closure != nil:
			fd = &model.FuncDecl{Params: t.closure.Params, Body: t.closure.Body}
		case t.fn != nil:
			fd, name = t.fn, t.fn.Name
		}
	}
	if fd == nil {
		return false
	}
	var ret any
	if fd.ReturnType != "" {
		ret = phpxPrefer(fd.ResolvedReturn, fd.ReturnType)
	}
	return phpxMap("name", name, "params", phpxParamInfo(fd), "returnType", ret, "file", fd.Filename)
}

// PhpxStaticProp reads a static property, and writes it when set is given as
// the last of rest.
func (rt *Runtime) PhpxStaticProp(class, name string, rest ...any) (any, error) {
	bag, err := rt.staticStorageFor(class, name, rt.newScope())
	if err != nil {
		return nil, err
	}
	if len(rest) == 2 {
		bag[name] = rest[0]
	}
	return bag[name], nil
}

// PhpxClassConst reads Class::NAME the way `Class::NAME` does, inherited
// constants included.
func (rt *Runtime) PhpxClassConst(class, name string) (any, error) {
	return rt.helperClassConst(&scopeRef{scope: rt.newScope()})(strings.TrimPrefix(class, "\\"), name)
}

// phpxPrefer answers the resolved spelling of a type when the parser recorded
// one, and the spelling as written otherwise.
func phpxPrefer(resolved, written string) string {
	if resolved != "" {
		return resolved
	}
	return written
}

// PhpxMethodNames answers get_class_methods(): the public methods of a class or
// of an object's class, declaration order, inherited ones after their own.
func (rt *Runtime) PhpxMethodNames(class any) ([]string, error) {
	info, ok := rt.PhpxReflectClass(class, true).(*model.Array)
	if !ok {
		return nil, fmt.Errorf("get_class_methods(): Argument #1 ($object_or_class) must be an object or a valid class name, %s given", phpDebugType(class))
	}
	methods, _ := info.Get("methods")
	var names []string
	model.RangeValues(methods, func(_, m any) bool {
		entry, ok := m.(*model.Array)
		if !ok {
			return true
		}
		if vis, _ := entry.Get("visibility"); vis != nil && phpString(vis) != "public" {
			return true
		}
		if name, ok := entry.Get("name"); ok {
			names = append(names, phpString(name))
		}
		return true
	})
	return names, nil
}
