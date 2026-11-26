//go:build phpx

package phpx

import (
	"strings"
	"testing"
)

// TestDOMParse pins the tree the console renderer walks: the body element, its
// attributes and classes, the text and element children, the sibling links and
// the serialisation saveXML answers.
func TestDOMParse(t *testing.T) {
	doc := newDOMDocument()
	html := `<?xml encoding="UTF-8"><!DOCTYPE html><html><body><div class="mt-1"><span>hi</span> &amp; <b>there</b></div><!--note--><br></body></html>`
	if !doc.LoadHTML(html) {
		t.Fatal("loadHTML answered false")
	}
	body := doc.GetElementsByTagName("body").Item(0)
	if body == nil {
		t.Fatal("no body element")
	}
	bn := nodeOf(body)
	if bn.NodeName != "body" {
		t.Fatalf("body name = %q", bn.NodeName)
	}
	div := doc.GetElementsByTagName("div").Item(0)
	if got := nodeOf(div).GetAttribute("class"); got != "mt-1" {
		t.Errorf("div class = %q, want mt-1", got)
	}
	span := doc.GetElementsByTagName("span").Item(0)
	if got := nodeOf(span).NodeValue; got != "hi" {
		t.Errorf("span nodeValue = %q, want hi", got)
	}
	if nodeOf(span).PreviousSibling != nil {
		t.Error("first child has a previous sibling")
	}
	if sib := nodeOf(nodeOf(span).NextSibling); sib == nil || sib.NodeType != 3 {
		t.Errorf("span next sibling = %v, want a text node", sib)
	}
	if got := saveDOMNode(nodeOf(body)); !strings.Contains(got, `<div class="mt-1">`) || !strings.Contains(got, `<br/>`) {
		t.Errorf("saveXML(body) = %q", got)
	}
	text := doc.CreateTextNode(`a & b`)
	if got := saveDOMNode(nodeOf(text)); got != "a &amp; b" {
		t.Errorf("saveXML(text) = %q", got)
	}
	if nodeOf(text).OwnerDocument != doc {
		t.Error("created text is not owned by the document")
	}
}

// TestDOMTextContent pins the value an element answers: the text of its
// descendants, which is what the renderer compares against ”.
func TestDOMTextContent(t *testing.T) {
	doc := newDOMDocument()
	doc.LoadHTML(`<div>a<span>b</span>c</div>`)
	div := doc.GetElementsByTagName("div").Item(0)
	if got := nodeOf(div).NodeValue; got != "abc" {
		t.Errorf("div nodeValue = %q, want abc", got)
	}
	if doc.GetElementsByTagName("table").Item(0) != nil {
		t.Error("missing element answered a node")
	}
}
