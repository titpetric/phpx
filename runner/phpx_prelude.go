//go:build phpx

package runner

import (
	_ "embed"
	"strings"

	"github.com/titpetric/phpx/parser"
)

//go:embed phpx_prelude/reflection.php
var phpxReflectionSource string

//go:embed phpx_prelude/spl.php
var phpxSPLSource string

// phpxPreludeFile is one embedded file of classes, declared together the first
// time a script names one of them. A class of the standard library is reached
// by name and declares nothing in the script that uses it.
type phpxPreludeFile struct {
	name    string
	source  string
	sentry  string // a class the file declares, whose presence says it is loaded
	classes map[string]bool
}

var phpxPreludes = []phpxPreludeFile{
	{
		name: "phpx:reflection.php", source: phpxReflectionSource, sentry: "ReflectionClass",
		classes: map[string]bool{
			"reflectionexception": true, "reflectionnamedtype": true, "reflectionparameter": true,
			"reflectionfunctionabstract": true, "reflectionmethod": true, "reflectionfunction": true,
			"reflectionproperty": true, "reflectionclass": true, "reflectionobject": true, "reflectionattribute": true,
		},
	},
	{
		name: "phpx:spl.php", source: phpxSPLSource, sentry: "ArrayObject",
		classes: map[string]bool{
			"splobserver": true, "splsubject": true, "arrayiterator": true, "arrayobject": true,
			"splobjectstorage": true, "spldoublylinkedlist": true, "splqueue": true, "splstack": true,
			"splfixedarray": true,
		},
	},
}

// phpxPrelude declares the file that holds class the first time it is asked for.
// It reports true when the name was one of them and is declared now.
func (rt *Runtime) phpxPrelude(class string) bool {
	key := strings.ToLower(strings.TrimPrefix(class, "\\"))
	for _, file := range phpxPreludes {
		if !file.classes[key] {
			continue
		}
		if _, ok := rt.lookupClass(file.sentry); ok {
			return rt.hasClass(class) || rt.phpxHasInterface(class)
		}
		prog, err := parser.ParseFile(file.name, file.source)
		if err != nil {
			return false
		}
		if err := rt.hoist(prog, file.name); err != nil {
			return false
		}
		return rt.hasClass(class) || rt.phpxHasInterface(class)
	}
	return false
}

func (rt *Runtime) phpxHasInterface(name string) bool {
	_, ok := rt.lookupInterface(name)
	return ok
}
