//go:build phpx

package phpx

import "testing"

func TestExtensionLoaded(t *testing.T) {
	exists := func(name string) bool { return name == "mb_strlen" }
	hasClass := func(name string) bool { return name == "DOMDocument" || name == "PDO" }
	if !extensionLoaded(exists, hasClass, "MBString") || !extensionLoaded(exists, hasClass, "spl") {
		t.Fatal("mbstring and spl should be loaded")
	}
	if !extensionLoaded(exists, hasClass, "dom") {
		t.Fatal("dom names a class marker and the class is present")
	}
	if !extensionLoaded(exists, hasClass, "pdo") {
		t.Fatal("pdo names a class marker and the class is present")
	}
	if extensionLoaded(exists, hasClass, "xmlwriter") {
		t.Fatal("xmlwriter names XMLWriter, which this runtime does not carry")
	}
	if extensionLoaded(exists, hasClass, "libxml") {
		t.Fatal("libxml stays off the table, its stub function is not a surface")
	}
	if extensionLoaded(exists, hasClass, "ctype") || extensionLoaded(exists, hasClass, "imagick") {
		t.Fatal("ctype has no marker function here and imagick is unknown")
	}
}
