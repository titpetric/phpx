//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxBuiltinType describes a class or interface the engine provides without a
// declaration, for reflection and class_exists.
type phpxBuiltinType struct {
	name       string
	iface      bool
	final      bool
	parent     string
	interfaces []string
	methods    []string
}

var phpxBuiltinTypes = func() map[string]phpxBuiltinType {
	list := []phpxBuiltinType{
		{name: "Closure", final: true, methods: []string{"bind", "bindTo", "call", "fromCallable", "__invoke"}},
		{name: "Generator", final: true, interfaces: []string{"Iterator", "Traversable"},
			methods: []string{"current", "getReturn", "key", "next", "rewind", "send", "throw", "valid"}},
		{name: "Traversable", iface: true},
		{name: "Iterator", iface: true, interfaces: []string{"Traversable"}, methods: []string{"current", "key", "next", "rewind", "valid"}},
		{name: "IteratorAggregate", iface: true, interfaces: []string{"Traversable"}, methods: []string{"getIterator"}},
		{name: "ArrayAccess", iface: true, methods: []string{"offsetExists", "offsetGet", "offsetSet", "offsetUnset"}},
		{name: "Countable", iface: true, methods: []string{"count"}},
		{name: "Stringable", iface: true, methods: []string{"__toString"}},
		{name: "JsonSerializable", iface: true, methods: []string{"jsonSerialize"}},
		{name: "Serializable", iface: true, methods: []string{"serialize", "unserialize"}},
		{name: "Throwable", iface: true, interfaces: []string{"Stringable"},
			methods: []string{"getMessage", "getCode", "getFile", "getLine", "getTrace", "getTraceAsString", "getPrevious", "__toString"}},
		{name: "UnitEnum", iface: true, methods: []string{"cases"}},
		{name: "BackedEnum", iface: true, interfaces: []string{"UnitEnum"}, methods: []string{"from", "tryFrom"}},
	}
	out := make(map[string]phpxBuiltinType, len(list))
	for _, t := range list {
		out[strings.ToLower(t.name)] = t
	}
	return out
}()

// PhpxBuiltinClass reports whether name is a class the engine provides itself.
func PhpxBuiltinClass(name string) bool {
	t, ok := phpxBuiltinTypes[strings.ToLower(strings.TrimPrefix(name, "\\"))]
	return ok && !t.iface
}

// phpxDescribeBuiltin answers the reflection record for an engine type.
func phpxDescribeBuiltin(name string) (*model.Array, bool) {
	t, ok := phpxBuiltinTypes[strings.ToLower(strings.TrimPrefix(name, "\\"))]
	if !ok {
		return nil, false
	}
	methods := model.NewArray()
	for _, m := range t.methods {
		methods.Set(m, phpxMap("name", m, "class", t.name, "visibility", "public", "static", false,
			"abstract", t.iface, "params", model.NewArray(), "returnType", nil, "file", false))
	}
	ifaces := model.NewArray()
	for _, i := range t.interfaces {
		ifaces.Append(i)
	}
	return phpxMap("name", t.name, "short", t.name, "namespace", "", "abstract", false, "final", t.final,
		"interface", t.iface, "internal", true, "readonly", false, "parent", nil, "interfaces", ifaces,
		"traits", model.NewArray(), "file", false, "methods", methods, "props", model.NewArray(), "consts", model.NewArray()), true
}

// phpxInternalNames is every class and interface the engine declares itself:
// the preludes' classes, the built-in throwables and stdClass.
var phpxInternalNames = func() map[string]bool {
	out := map[string]bool{"stdclass": true}
	for _, file := range phpxPreludes {
		for name := range file.classes {
			out[name] = true
		}
		out[strings.ToLower(file.sentry)] = true
	}
	for _, pair := range phpxThrowableParents {
		out[strings.ToLower(pair[0])] = true
		out[strings.ToLower(pair[1])] = true
	}
	return out
}()

// phpxInternal reports whether name is declared by the engine and not a script.
func phpxInternal(name string) bool {
	return phpxInternalNames[strings.ToLower(strings.TrimPrefix(name, "\\"))]
}

// phpxDescribeTrait answers the reflection record for a trait: its methods and
// properties, abstract false, flagged as a trait.
func (rt *Runtime) phpxDescribeTrait(td *model.TraitDecl) *model.Array {
	short, ns := phpxSplitName(td.Name)
	methods := model.NewArray()
	for _, m := range td.Methods {
		methods.Set(m.Name, phpxMethodInfo(m, td.Name, false))
	}
	props := model.NewArray()
	consts := model.NewArray()
	for _, f := range td.Consts {
		consts.Set(f.Name, f.Name)
	}
	file := any(false)
	for _, m := range td.Methods {
		if m.Filename != "" {
			file = m.Filename
			break
		}
	}
	return phpxMap("name", td.Name, "short", short, "namespace", ns, "abstract", false, "final", false,
		"interface", false, "trait", true, "internal", false, "readonly", false, "parent", nil,
		"interfaces", model.NewArray(), "traits", model.NewArray(), "file", file, "methods", methods,
		"props", props, "consts", consts)
}
