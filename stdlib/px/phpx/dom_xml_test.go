//go:build phpx

package phpx

import "testing"

const xmlFixture = `<?xml version="1.0" encoding="UTF-8"?>
<!-- a comment before the root -->
<phpunit bootstrap="vendor/autoload.php" cacheDirectory=".phpunit.cache">
    <testsuites>
        <testsuite name="one">
            <directory suffix=".php">tests/unit</directory>
        </testsuite>
    </testsuites>
</phpunit>`

// TestLoadXMLKeepsCase reads the element and attribute case php reads: the
// HTML parse folds both, and a configuration file spells them.
func TestLoadXMLKeepsCase(t *testing.T) {
	doc := newDOMDocument()
	if !doc.LoadXML(xmlFixture) {
		t.Fatal("loadXML answered false")
	}
	root := doc.DocumentElement
	if root == nil || root.NodeName != "phpunit" {
		t.Fatalf("documentElement = %v", root)
	}
	if got := root.GetAttribute("cacheDirectory"); got != ".phpunit.cache" {
		t.Fatalf("cacheDirectory = %q", got)
	}
	if got := root.GetAttribute("bootstrap"); got != "vendor/autoload.php" {
		t.Fatalf("bootstrap = %q", got)
	}
}

// TestXPathReadsConfiguration asks the queries the readers ask: a relative
// path from the root element, a descendant, a predicate and a comment.
func TestXPathReadsConfiguration(t *testing.T) {
	doc := newDOMDocument()
	doc.LoadXML(xmlFixture)
	xpath := &DOMXPath{Document: doc}

	for _, tc := range []struct {
		query string
		want  int64
	}{
		{"testsuites/testsuite", 1},
		{"//testsuite", 1},
		{"testsuites/testsuite[@name='one']", 1},
		{"testsuites/testsuite[@name='two']", 0},
		{"testsuites/testsuite/directory", 1},
		{"//comment()", 1},
		{"testsuite", 0},
	} {
		if got := xpath.Query(tc.query).Length; got != tc.want {
			t.Errorf("query %q = %d nodes, want %d", tc.query, got, tc.want)
		}
	}
}

// TestXPathNodeListIterates walks a result the way a foreach does.
func TestXPathNodeListIterates(t *testing.T) {
	doc := newDOMDocument()
	doc.LoadXML(xmlFixture)
	list := (&DOMXPath{Document: doc}).Query("testsuites/testsuite")
	var names []string
	if err := list.Rewind(); err != nil {
		t.Fatal(err)
	}
	for {
		ok, err := list.Valid()
		if err != nil || !ok {
			break
		}
		node, err := list.Current()
		if err != nil {
			t.Fatal(err)
		}
		if n := nodeOf(node); n != nil {
			names = append(names, n.NodeName)
		}
		list.Next()
	}
	if len(names) != 1 || names[0] != "testsuite" {
		t.Fatalf("iterated %v", names)
	}
}

// TestElementGetElementsByTagName answers the call the configuration reader
// makes on an element and not only on the document.
func TestElementGetElementsByTagName(t *testing.T) {
	doc := newDOMDocument()
	doc.LoadXML(xmlFixture)
	suites := doc.DocumentElement.GetElementsByTagName("testsuite")
	if suites.Length != 1 {
		t.Fatalf("testsuite nodes = %d", suites.Length)
	}
	suite, ok := suites.Item(0).(*DOMElement)
	if !ok {
		t.Fatalf("testsuite node = %T", suites.Item(0))
	}
	if dirs := suite.GetElementsByTagName("directory"); dirs.Length != 1 {
		t.Fatalf("directory nodes = %d", dirs.Length)
	}
}
