//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

// The DOM subset the console renderer reads. Laravel builds every console
// component as an HTML string and Termwind parses it with DOMDocument to walk
// the nodes; no script here parses XML, runs XPath or edits a document, so the
// binding holds the parse, the traversal and the serialisation that renderer
// calls and nothing else. The types answer `instanceof` by name, which is how
// the renderer tells a text node from an element.

func init() {
	runner.RegisterArea(`Runtime\PX\DOM`, registerPhpxDOM)
}

// registerPhpxDOM installs the node classes and the constants the parse reads.
func registerPhpxDOM(rt *runner.Runtime) {
	for name, value := range map[string]int64{
		"LIBXML_NOERROR": 32, "LIBXML_COMPACT": 65536, "LIBXML_HTML_NODEFDTD": 4,
		"LIBXML_NOBLANKS": 256, "LIBXML_NOXMLDECL": 2, "LIBXML_HTML_NOIMPLIED": 8192,
		"XML_ELEMENT_NODE": 1, "XML_TEXT_NODE": 3, "XML_COMMENT_NODE": 8,
		"XML_DOCUMENT_NODE": 9, "XML_HTML_DOCUMENT_NODE": 13,
		"XML_ATTRIBUTE_NODE": 2, "XML_CDATA_SECTION_NODE": 4, "XML_PI_NODE": 7,
		"LIBXML_NONET": 2048, "LIBXML_NOCDATA": 16384, "LIBXML_NOENT": 2,
		"LIBXML_NOWARNING": 64, "LIBXML_NOEMPTYTAG": 4,
	} {
		rt.SetConst(name, value)
	}
	// php's constructor takes the version and encoding of the declaration and
	// the HTML parse ignores both; two parameters cut the call at the arity
	// php cuts it at, and the runtime zero-pads a call that passes none.
	rt.RegisterConstructor("DOMDocument", func(version, encoding any) *DOMDocument { return newDOMDocument() })
	rt.RegisterConstructor("DOMElement", func(name string, value ...string) *DOMElement {
		return newDOMElement(name, firstString(value))
	})
	rt.RegisterConstructor("DOMText", func(value string) *DOMText { return newDOMText(value) })
	rt.RegisterConstructor("DOMComment", func(value string) *DOMComment { return newDOMComment(value) })
	// The class queries the document the caller parsed.
	rt.RegisterConstructor("DOMXPath", func(document *DOMDocument) *DOMXPath {
		return &DOMXPath{Document: document}
	})
}

// DOMDocument is `new DOMDocument` and the tree the parse builds.
type DOMDocument struct {
	DocumentElement *DOMElement
	// The flags a caller writes before or after a parse. The binding keeps
	// them as fields so php's `$document->preserveWhiteSpace = false` is a
	// write and not a refused property; the parse already drops whitespace,
	// which is what the flag asks for.
	PreserveWhiteSpace bool
	FormatOutput       bool
	ValidateOnParse    bool
	SubstituteEntities bool
	ResolveExternals   bool
	root               *domNode
}

// DOMElement is one element node.
type DOMElement struct{ domNode }

// DOMText is one text node.
type DOMText struct{ domNode }

// DOMComment is one comment node.
type DOMComment struct{ domNode }

// DOMNodeList is the list getElementsByTagName answers.
type DOMNodeList struct {
	Length int64
	items  []any
	// pos is the foreach cursor, which Rewind resets.
	pos int
}

// Item answers the node at position i, or nil past the end.
func (l *DOMNodeList) Item(i int64) any {
	if l == nil || i < 0 || int(i) >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// domNode carries the fields every node class shares. The exported names are
// the php properties, matched case-insensitively by the runtime's reader.
type domNode struct {
	NodeName        string
	NodeType        int64
	NodeValue       string
	ChildNodes      []any
	ParentNode      any
	PreviousSibling any
	NextSibling     any
	OwnerDocument   *DOMDocument
	TagName         string
	TextContent     string
	FirstChild      any
	LastChild       any
	Attrs           map[string]string
	// Attributes is the list php exposes on a node, length only: a caller
	// asks how many attributes an element carries.
	Attributes      *DOMNodeList
	parent          *domNode
	self            any
}

func newDOMDocument() *DOMDocument {
	return &DOMDocument{}
}

func newDOMElement(name, value string) *DOMElement {
	e := &DOMElement{}
	e.NodeName = strings.ToLower(name)
	e.TagName = e.NodeName
	e.NodeType = 1
	e.NodeValue = value
	e.TextContent = value
	e.Attrs = map[string]string{}
	e.self = e
	return e
}

// newXMLElement builds an element whose name keeps the case the document
// spelled, which is what the XML parse answers.
func newXMLElement(name string) *DOMElement {
	e := &DOMElement{}
	e.NodeName = name
	e.TagName = name
	e.NodeType = 1
	e.Attrs = map[string]string{}
	e.self = e
	return e
}

func newDOMText(value string) *DOMText {
	t := &DOMText{}
	t.NodeName = "#text"
	t.NodeType = 3
	t.NodeValue = value
	t.TextContent = value
	t.self = t
	return t
}

func newDOMComment(value string) *DOMComment {
	c := &DOMComment{}
	c.NodeName = "#comment"
	c.NodeType = 8
	c.NodeValue = value
	c.TextContent = value
	c.self = c
	return c
}

// LoadHTML parses html into the document, replacing the tree it held.
func (d *DOMDocument) LoadHTML(html string, _ ...int64) bool {
	d.root = parseDOM(html, d)
	d.DocumentElement = nil
	for _, child := range d.root.ChildNodes {
		if n := nodeOf(child); n != nil && n.NodeType == 1 {
			if e, ok := n.self.(*DOMElement); ok {
				d.DocumentElement = e
				break
			}
		}
	}
	return true
}

// SaveXML serialises the named node, or the whole document without one.
func (d *DOMDocument) SaveXML(node ...any) string {
	if len(node) > 0 && node[0] != nil {
		if n := nodeOf(node[0]); n != nil {
			return saveDOMNode(n)
		}
	}
	if d.root == nil {
		return ""
	}
	var b strings.Builder
	for _, child := range d.root.ChildNodes {
		b.WriteString(saveDOMNode(nodeOf(child)))
	}
	return b.String()
}

// SaveHTML is SaveXML for the callers that spell it that way.
func (d *DOMDocument) SaveHTML(node ...any) string { return d.SaveXML(node...) }

// CreateElement builds an element node owned by the document.
func (d *DOMDocument) CreateElement(name string, value ...string) *DOMElement {
	e := newDOMElement(name, firstString(value))
	e.OwnerDocument = d
	return e
}

// CreateTextNode builds a text node owned by the document.
func (d *DOMDocument) CreateTextNode(value string) *DOMText {
	t := newDOMText(value)
	t.OwnerDocument = d
	return t
}

// CreateComment builds a comment node owned by the document.
func (d *DOMDocument) CreateComment(value string) *DOMComment {
	c := newDOMComment(value)
	c.OwnerDocument = d
	return c
}

// GetElementsByTagName answers every element of that name below the document.
func (d *DOMDocument) GetElementsByTagName(name string) *DOMNodeList {
	out := &DOMNodeList{}
	if d.root == nil {
		return out
	}
	want := strings.ToLower(name)
	var walk func(*domNode)
	walk = func(n *domNode) {
		if n.NodeType == 1 && (want == "*" || n.NodeName == want || strings.ToLower(n.NodeName) == want) {
			out.items = append(out.items, n.self)
		}
		for _, child := range n.ChildNodes {
			if c := nodeOf(child); c != nil {
				walk(c)
			}
		}
	}
	for _, child := range d.root.ChildNodes {
		if c := nodeOf(child); c != nil {
			walk(c)
		}
	}
	out.Length = int64(len(out.items))
	return out
}

// GetElementById is the lookup a page uses when it names an id.
func (d *DOMDocument) GetElementById(id string) any {
	list := d.GetElementsByTagName("*")
	for _, item := range list.items {
		if n := nodeOf(item); n != nil && n.Attrs["id"] == id {
			return item
		}
	}
	return nil
}

// GetElementsByTagName answers every element of that name below the node.
func (n *domNode) GetElementsByTagName(name string) *DOMNodeList {
	out := &DOMNodeList{}
	want := strings.ToLower(name)
	var walk func(*domNode)
	walk = func(n *domNode) {
		if n.NodeType == 1 && (want == "*" || n.NodeName == want || strings.ToLower(n.NodeName) == want) {
			out.items = append(out.items, n.self)
		}
		for _, child := range n.ChildNodes {
			if c := nodeOf(child); c != nil {
				walk(c)
			}
		}
	}
	for _, child := range n.ChildNodes {
		if c := nodeOf(child); c != nil {
			walk(c)
		}
	}
	out.Length = int64(len(out.items))
	return out
}

// GetAttribute answers the attribute value, or "" when the node has none.
func (n *domNode) GetAttribute(name string) string {
	if n.Attrs == nil {
		return ""
	}
	if value, ok := n.Attrs[name]; ok {
		return value
	}
	return n.Attrs[strings.ToLower(name)]
}

// GetAttributeNS answers the attribute with that local name, whatever prefix
// the document gave it, as php reads `xsi:noNamespaceSchemaLocation`.
func (n *domNode) GetAttributeNS(_, name string) string {
	if n.Attrs == nil {
		return ""
	}
	name = strings.TrimPrefix(name, "xsi:")
	if value, ok := n.Attrs[name]; ok {
		return value
	}
	for key, value := range n.Attrs {
		if local := key[strings.LastIndexByte(key, ':')+1:]; local == name {
			return value
		}
	}
	return ""
}

// SetAttribute writes an attribute and keeps the attribute list current.
func (n *domNode) SetAttribute(name, value string) any {
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	n.Attrs[name] = value
	n.Attributes = &DOMNodeList{Length: int64(len(n.Attrs))}
	return nil
}

// HasAttribute reports whether the node carries the attribute.
func (n *domNode) HasAttribute(name string) bool {
	if n.Attrs == nil {
		return false
	}
	if _, ok := n.Attrs[name]; ok {
		return true
	}
	_, ok := n.Attrs[strings.ToLower(name)]
	return ok
}

// AppendChild attaches child and answers it.
func (n *domNode) AppendChild(child any) any { return n.appendChild(child) }

// RemoveChild detaches child and answers it.
func (n *domNode) RemoveChild(child any) any {
	c := nodeOf(child)
	if c == nil {
		return nil
	}
	for i, have := range n.ChildNodes {
		if nodeOf(have) == c {
			n.ChildNodes = append(n.ChildNodes[:i], n.ChildNodes[i+1:]...)
			c.parent = nil
			n.reflow()
			return child
		}
	}
	return nil
}

// appendChild attaches child below n and keeps the sibling links.
func (n *domNode) appendChild(child any) any {
	c := nodeOf(child)
	if c == nil {
		return nil
	}
	c.parent = n
	c.OwnerDocument = n.OwnerDocument
	n.ChildNodes = append(n.ChildNodes, child)
	n.reflow()
	return child
}

// reflow recomputes first/last child and the sibling links php exposes.
func (n *domNode) reflow() {
	n.FirstChild, n.LastChild = nil, nil
	for i, child := range n.ChildNodes {
		c := nodeOf(child)
		if c == nil {
			continue
		}
		c.ParentNode = n.self
		c.PreviousSibling, c.NextSibling = nil, nil
		if i > 0 {
			if prev := nodeOf(n.ChildNodes[i-1]); prev != nil {
				c.PreviousSibling = prev.self
			}
		}
		if i+1 < len(n.ChildNodes) {
			if next := nodeOf(n.ChildNodes[i+1]); next != nil {
				c.NextSibling = next.self
			}
		}
		if i == 0 {
			n.FirstChild = child
		}
		n.LastChild = child
	}
}

// nodeOf answers the shared node behind a typed value.
func nodeOf(v any) *domNode {
	switch t := v.(type) {
	case *DOMElement:
		return &t.domNode
	case *DOMText:
		return &t.domNode
	case *DOMComment:
		return &t.domNode
	}
	return nil
}

func firstString(values []string) string {
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
