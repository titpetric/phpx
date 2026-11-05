//go:build phpx

package runner

import "testing"

func TestPhpxBuiltinTypes(t *testing.T) {
	if !PhpxBuiltinClass("\\Closure") || PhpxBuiltinClass("Countable") || PhpxBuiltinClass("Nope") {
		t.Fatal("Closure is a class, Countable an interface, Nope nothing")
	}
	info, ok := phpxDescribeBuiltin("countable")
	if !ok {
		t.Fatal("Countable has no record")
	}
	if v, _ := info.Get("interface"); v != true {
		t.Fatalf("interface = %v", v)
	}
}

func TestPhpxInternal(t *testing.T) {
	for _, name := range []string{"stdClass", "ArrayObject", "Exception", "\\RuntimeException", "UnitEnum"} {
		if !phpxInternal(name) {
			t.Errorf("%s is declared by the engine", name)
		}
	}
	if phpxInternal("App\\Model") {
		t.Error("a script's class is not internal")
	}
}
