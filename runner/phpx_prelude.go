//go:build phpx

package runner

import (
	_ "embed"
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
			"splfixedarray": true, "weakmap": true, "fiber": true, "datetimeinterface": true, "directoryiterator": true, "splfileinfo": true, "filesystemiterator": true, "recursivedirectoryiterator": true, "unitenum": true, "backedenum": true, "unhandledmatcherror": true, "outeriterator": true, "recursiveiterator": true, "iteratoriterator": true,
			"filteriterator": true, "callbackfilteriterator": true, "recursivefilteriterator": true,
			"limititerator": true, "appenditerator": true, "emptyiterator": true, "recursivearrayiterator": true, "recursiveiteratoriterator": true,
		},
	},
}

// PhpxRegisterPreludes hands each embedded file to the runtime, which declares
// it when autoload first asks for one of its classes.
func PhpxRegisterPreludes(rt *Runtime) {
	for _, file := range phpxPreludes {
		names := make([]string, 0, len(file.classes)+1)
		names = append(names, file.sentry)
		for name := range file.classes {
			names = append(names, name)
		}
		rt.RegisterClassSource(file.name, file.source, names...)
	}
}
