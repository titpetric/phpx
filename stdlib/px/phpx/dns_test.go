//go:build phpx

package phpx

import "testing"

func TestDNSKinds(t *testing.T) {
	if got := dnsKinds(dnsMX | dnsA); len(got) != 2 || got[0] != dnsA || got[1] != dnsMX {
		t.Fatalf("kinds = %v", got)
	}
	if got := dnsKinds(dnsANY); len(got) != 6 {
		t.Fatalf("ANY = %v", got)
	}
}

func TestDNSRecordShape(t *testing.T) {
	r := dnsRecord("example.test", "MX", "pri", int64(10), "target", "mail.example.test")
	if v, _ := r.Get("type"); v != "MX" {
		t.Fatalf("type = %v", v)
	}
	if v, _ := r.Get("pri"); v != int64(10) {
		t.Fatalf("pri = %v", v)
	}
}
