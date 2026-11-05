package parser

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxHooksLowerToMethods(t *testing.T) {
	prog, err := ParseFile("t.php", `<?php class A {
		public private(set) int $n = 1 { get => $this->n; set { $this->n = $value; } }
	}`)
	if err != nil {
		t.Fatal(err)
	}
	cd := prog.Stmts[0].(*model.ClassDecl)
	if len(cd.Fields) != 1 || cd.Fields[0].Visibility != "public" {
		t.Fatalf("fields = %+v, want one public field", cd.Fields)
	}
	names := []string{}
	for _, m := range cd.Methods {
		names = append(names, m.Name)
	}
	if len(names) != 2 || names[0] != "__phpx_get_n" || names[1] != "__phpx_set_n" {
		t.Fatalf("methods = %v, want the get and set hooks", names)
	}
	if got := cd.Methods[1].Params; len(got) != 1 || got[0].Name != "value" {
		t.Fatalf("set params = %+v, want $value", got)
	}
}
