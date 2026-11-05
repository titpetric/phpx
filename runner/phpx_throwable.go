//go:build phpx

package runner

import (
	"fmt"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/parser"
)

// phpxThrowableParents is php's hierarchy of the built-in throwables, name to
// parent. A name absent here is not one, and `extends` on it keeps the
// host-class answer.
var phpxThrowableParents = [][2]string{
	{"Exception", ""},
	{"Error", ""},
	{"ErrorException", "Exception"},
	{"RuntimeException", "Exception"},
	{"LogicException", "Exception"},
	{"JsonException", "Exception"},
	{"InvalidArgumentException", "LogicException"},
	{"DomainException", "LogicException"},
	{"LengthException", "LogicException"},
	{"OutOfRangeException", "LogicException"},
	{"BadFunctionCallException", "LogicException"},
	{"BadMethodCallException", "BadFunctionCallException"},
	{"OutOfBoundsException", "RuntimeException"},
	{"RangeException", "RuntimeException"},
	{"OverflowException", "RuntimeException"},
	{"UnderflowException", "RuntimeException"},
	{"UnexpectedValueException", "RuntimeException"},
	{"TypeError", "Error"},
	{"ValueError", "Error"},
	{"ArithmeticError", "Error"},
	{"DivisionByZeroError", "ArithmeticError"},
	{"ArgumentCountError", "TypeError"},
	{"UnhandledMatchError", "Error"},
	{"AssertionError", "Error"},
	{"CompileError", "Error"},
	{"ParseError", "CompileError"},
}

// phpxThrowableBody is the member set a throwable root declares. The Go
// constructor of `new Exception` builds a value with no members a subclass
// could reach, so `extends Exception` gets this declaration as its parent.
const phpxThrowableBody = `{
	protected $message = "";
	protected $code = 0;
	protected $previous = null;
	protected $file = "";
	protected $line = 0;
	public function getMessage() { return $this->message; }
	public function getCode() { return $this->code; }
	public function getPrevious() { return $this->previous; }
	public function getFile() { return $this->file; }
	public function getLine() { return $this->line; }
	public function getTrace() { return []; }
	public function getTraceAsString() { return "#0 {main}"; }
	public function __toString() { return get_class($this) . ": " . $this->message; }
`

const phpxThrowableCtor = `	public function __construct($message = "", $code = 0, $previous = null) {
		$this->message = $message;
		$this->code = $code;
		$this->previous = $previous;
	}
}`

const phpxErrorExceptionCtor = `	protected $severity = 1;
	public function __construct($message = "", $code = 0, $severity = 1, $filename = null, $line = null, $previous = null) {
		$this->message = $message;
		$this->code = $code;
		$this->severity = $severity;
		$this->previous = $previous;
	}
	public function getSeverity() { return $this->severity; }
}`

var (
	phpxThrowableOnce  sync.Once
	phpxThrowableDecls map[string]*model.ClassDecl
)

// phpxThrowableDeclarations parses the built-in throwables once; the
// declarations are read-only after that.
func phpxThrowableDeclarations() map[string]*model.ClassDecl {
	phpxThrowableOnce.Do(func() {
		var src strings.Builder
		src.WriteString("<?php\n")
		for _, pair := range phpxThrowableParents {
			name, parent := pair[0], pair[1]
			switch {
			case parent == "":
				fmt.Fprintf(&src, "class %s implements Throwable %s%s\n", name, phpxThrowableBody, phpxThrowableCtor)
			case name == "ErrorException":
				fmt.Fprintf(&src, "class ErrorException extends Exception %s%s\n", phpxThrowableBody, phpxErrorExceptionCtor)
			default:
				fmt.Fprintf(&src, "class %s extends %s {}\n", name, parent)
			}
		}
		phpxThrowableDecls = map[string]*model.ClassDecl{}
		prog, err := parser.Parse(src.String())
		if err != nil {
			panic("phpx: throwable declarations: " + err.Error())
		}
		for _, stmt := range prog.Stmts {
			if cd, ok := stmt.(*model.ClassDecl); ok {
				phpxThrowableDecls[strings.ToLower(cd.Name)] = cd
			}
		}
	})
	return phpxThrowableDecls
}

// phpxThrowableParent builds the parent class for `extends` on a built-in
// throwable, or reports false for any other host class.
func (rt *Runtime) phpxThrowableParent(name string, traits map[string]*model.TraitDecl, stmts []model.Stmt, chain map[string]bool) (*model.Class, bool, error) {
	decls := phpxThrowableDeclarations()
	cd, ok := decls[strings.ToLower(strings.TrimPrefix(name, "\\"))]
	if !ok {
		return nil, false, nil
	}
	c, err := rt.phpxClassChain(cd, decls, traits, stmts, chain)
	if err != nil {
		return nil, true, err
	}
	// `parent::__construct()` resolves the parent by name through the class
	// table, so the class and every ancestor of it are registered. A script
	// that asks `class_exists('RuntimeException')` hears yes, which is php's.
	rt.classes[c.Name] = c
	for _, ancestor := range c.Ancestors {
		if _, registered := rt.classes[ancestor]; !registered {
			if _, _, err := rt.phpxThrowableParent(ancestor, traits, stmts, chain); err != nil {
				return nil, true, err
			}
		}
	}
	return c, true, nil
}

// phpxIsErrorClass reports whether the throwable descends from Error. A thrown
// object answers from its recorded chain, because a class named Fault that
// extends Error is an Error; any other value keeps upstream's name suffix.
func phpxIsErrorClass(class string, err error) bool {
	if thrown, ok := err.(*objectError); ok && thrown.object != nil && thrown.object.Class != nil {
		for _, name := range thrown.object.Class.Ancestors {
			if strings.EqualFold(name, "Error") {
				return true
			}
		}
		if len(thrown.object.Class.Ancestors) > 0 {
			return strings.EqualFold(class, "Error")
		}
	}
	return strings.HasSuffix(class, "Error")
}
