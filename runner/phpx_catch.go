//go:build phpx

package runner

import "strings"

// phpxHostErrorCatches reports whether a clause naming want takes an error a Go
// binding returned that carries no php class. php raises such a failure as one
// of its own throwable classes, so the clauses that name one of those take it;
// a class a script or a library declared (a test framework's IncompleteTest, a
// validator's ValidationException) is never what an internal failure is.
func phpxHostErrorCatches(want string) bool {
	_, ok := phpxBuiltinThrowables[strings.ToLower(strings.TrimPrefix(want, "\\"))]
	return ok
}

var phpxBuiltinThrowables = func() map[string]struct{} {
	out := map[string]struct{}{}
	for _, name := range []string{
		"exception", "error", "errorexception", "typeerror", "valueerror", "arithmeticerror",
		"divisionbyzeroerror", "argumentcounterror", "assertionerror", "unhandledmatcherror",
		"runtimeexception", "logicexception", "invalidargumentexception", "domainexception",
		"lengthexception", "outofrangeexception", "outofboundsexception", "rangeexception",
		"overflowexception", "underflowexception", "unexpectedvalueexception",
		"badfunctioncallexception", "badmethodcallexception", "jsonexception",
		"reflectionexception", "stringable", "throwable",
	} {
		out[name] = struct{}{}
	}
	return out
}()
