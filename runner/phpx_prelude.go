//go:build phpx

package runner

import (
	_ "embed"
	"strings"
	"sync"

	"github.com/titpetric/phpx/parser"
)

//go:embed phpx_prelude/reflection.php
var phpxReflectionSource string

// phpxPreludeClasses are the classes the embedded prelude declares, which a
// script reaches by name as it does a class of the standard library.
var phpxPreludeClasses = map[string]bool{
	"reflectionexception": true, "reflectionnamedtype": true, "reflectionparameter": true,
	"reflectionfunctionabstract": true, "reflectionmethod": true, "reflectionfunction": true,
	"reflectionproperty": true, "reflectionclass": true, "reflectionobject": true,
}

var phpxPreludeLoading sync.Mutex

// phpxPrelude declares the prelude classes the first time one is asked for.
// It reports true when the name was one of them and is declared now.
func (rt *Runtime) phpxPrelude(class string) bool {
	if !phpxPreludeClasses[strings.ToLower(strings.TrimPrefix(class, "\\"))] {
		return false
	}
	if _, ok := rt.lookupClass("ReflectionClass"); ok {
		return rt.hasClass(class)
	}
	prog, err := parser.ParseFile("phpx:reflection.php", phpxReflectionSource)
	if err != nil {
		return false
	}
	if err := rt.hoist(prog, "phpx:reflection.php"); err != nil {
		return false
	}
	return rt.hasClass(class)
}
