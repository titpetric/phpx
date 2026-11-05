//go:build phpx

package phpx

import "testing"

func TestExtensionLoaded(t *testing.T) {
	exists := func(name string) bool { return name == "mb_strlen" }
	if !extensionLoaded(exists, "MBString") || !extensionLoaded(exists, "spl") {
		t.Fatal("mbstring and spl should be loaded")
	}
	if extensionLoaded(exists, "ctype") || extensionLoaded(exists, "imagick") {
		t.Fatal("ctype has no marker function here and imagick is unknown")
	}
}
