package parser

import "testing"

func TestQualifyCall(t *testing.T) {
	p := &parser{namespace: "App", imports: map[string]string{"Config": "Foo\\Config", "helper": "Foo\\helper"}}
	for name, want := range map[string]string{"config": "App\\config", "Config": "Foo\\Config", "helper": "Foo\\helper", "other": "App\\other"} {
		if got := p.qualifyCall(name); got != want {
			t.Errorf("qualifyCall(%q) = %q, want %q", name, got, want)
		}
	}
}
