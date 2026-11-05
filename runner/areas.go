package runner

import (
	"fmt"
	"strings"
)

// The namespaces a binding area lives under. Runtime\PS is phpscript's own
// surface, Runtime\PX the phpx overlay that installs over it, and Runtime\Go
// the opt-in bindings of go's standard library, which a script imports.
const (
	NamespacePS = `Runtime\PS`
	NamespacePX = `Runtime\PX`
	NamespaceGo = `Runtime\Go`

	// NamespaceUnnamed holds an installer contributed without a name.
	NamespaceUnnamed = NamespacePS + `\Unnamed`
)

// RegisterArea contributes an installer under ProfileSecure in the named area.
func RegisterArea(name string, installer func(*Runtime)) {
	RegisterProfileArea(name, ProfileSecure, installer)
}

// RegisterProfileArea contributes an installer in the named area under a profile.
func RegisterProfileArea(name string, profile Profile, installer func(*Runtime)) {
	if installer == nil {
		return
	}
	if name == "" {
		name = NamespaceUnnamed
	}
	bindings.Lock()
	defer bindings.Unlock()
	bindings.installers = append(bindings.installers, binding{name: name, layer: areaLayer(name), profile: profile, install: installer})
}

// AreasFor returns the areas whose profile overlaps the mask, Runtime\PS first.
func AreasFor(profile Profile) []Area {
	bindings.Lock()
	defer bindings.Unlock()
	areas := make([]Area, 0, len(bindings.installers))
	for layer := 0; layer <= areaLayers; layer++ {
		for _, b := range bindings.installers {
			if b.layer == layer && b.profile&profile != 0 {
				areas = append(areas, Area{Name: b.name, Profile: b.profile, Install: b.install})
			}
		}
	}
	return areas
}

// SelectAreas keeps the areas that sit under one of the namespaces. An empty
// list keeps them all: the default runtime carries every area.
func SelectAreas(areas []Area, namespaces []string) []Area {
	if len(namespaces) == 0 {
		return areas
	}
	out := make([]Area, 0, len(areas))
	for _, area := range areas {
		for _, ns := range namespaces {
			if InNamespace(area.Name, ns) {
				out = append(out, area)
				break
			}
		}
	}
	return out
}

// InNamespace reports whether name is the namespace or sits below it, as php compares names.
func InNamespace(name, namespace string) bool {
	name, namespace = foldName(name), foldName(namespace)
	return name == namespace || strings.HasPrefix(name, namespace+`\`)
}

// areaLayers is the highest layer: 0 is Runtime\PS and the rest, 1 is Runtime\PX.
const areaLayers = 1

// areaLayer orders the areas: the overlay installs after what it overlays.
func areaLayer(name string) int {
	if InNamespace(name, NamespacePX) {
		return 1
	}
	return 0
}

// Areas reports the namespaces the runtime was built to install; empty means all.
func (rt *Runtime) Areas() []string { return rt.opts.Areas }

// CheckAreas refuses a namespace that names no registered area or opt-in
// namespace. A misspelling would otherwise select nothing and leave a runtime
// without bindings. It passes everything when no binding package is linked in.
func CheckAreas(namespaces []string) error {
	registered := AreasFor(^Profile(0))
	opt := Namespaces()
	if len(registered) == 0 {
		return nil
	}
	for _, ns := range namespaces {
		found := false
		for _, area := range registered {
			found = found || InNamespace(area.Name, ns)
		}
		for _, name := range opt {
			found = found || InNamespace(name, ns)
		}
		if !found {
			return fmt.Errorf("runner.areas: %q names no area; phpscript list --areas lists them", ns)
		}
	}
	return nil
}
