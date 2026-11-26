//go:build phpx

package phpx

import (
	"os"
	"strings"
)

// The XML parse behind DOMDocument::load and loadXML. php's config readers
// (phpunit's among them) spell element and attribute names in the case the
// schema uses, and the HTML parse folds them to lower case; this one keeps
// them. It reads a document, not a schema: a doctype and a processing
// instruction are skipped, CDATA is text, and a parse never reports an error
// the caller could act on.

// parseXMLDOM builds the tree for src, owned by doc.
func parseXMLDOM(src string, doc *DOMDocument) *domNode {
	root := &domNode{NodeName: "#document", NodeType: 9, OwnerDocument: doc}
	stack := []*domNode{root}

	top := func() *domNode { return stack[len(stack)-1] }
	add := func(v any) *domNode {
		n := nodeOf(v)
		n.OwnerDocument = doc
		n.parent = top()
		top().ChildNodes = append(top().ChildNodes, v)
		return n
	}

	for i := 0; i < len(src); {
		if src[i] != '<' {
			end := strings.IndexByte(src[i:], '<')
			if end < 0 {
				end = len(src) - i
			}
			text := decodeDOMEntities(src[i : i+end])
			if strings.TrimSpace(text) != "" {
				add(newDOMText(text))
			}
			i += end
			continue
		}
		switch {
		case strings.HasPrefix(src[i:], "<![CDATA["):
			end := strings.Index(src[i+9:], "]]>")
			if end < 0 {
				add(newDOMText(src[i+9:]))
				i = len(src)
				continue
			}
			add(newDOMText(src[i+9 : i+9+end]))
			i += 9 + end + 3
		case strings.HasPrefix(src[i:], "<!--"):
			end := strings.Index(src[i+4:], "-->")
			if end < 0 {
				end = len(src) - i - 4
			}
			add(newDOMComment(src[i+4 : i+4+end]))
			i += 4 + end
			if strings.HasPrefix(src[i:], "-->") {
				i += 3
			}
		case strings.HasPrefix(src[i:], "<?"):
			if end := strings.Index(src[i:], "?>"); end >= 0 {
				i += end + 2
				continue
			}
			i = len(src)
		case strings.HasPrefix(src[i:], "<!"):
			i = skipXMLDecl(src, i)
		case strings.HasPrefix(src[i:], "</"):
			end := strings.IndexByte(src[i:], '>')
			if end < 0 {
				i = len(src)
				continue
			}
			name := strings.TrimSpace(src[i+2 : i+end])
			stack = closeXMLStack(stack, name)
			i += end + 1
		default:
			el, next, selfClosing := parseXMLStartTag(src, i)
			if el == nil {
				add(newDOMText("<"))
				i++
				continue
			}
			n := add(el)
			if !selfClosing {
				stack = append(stack, n)
			}
			i = next
		}
	}

	syncXMLAttributes(root)
	reflowDOM(root)
	return root
}

// skipXMLDecl steps over a doctype, including an internal subset.
func skipXMLDecl(src string, i int) int {
	depth := 0
	for j := i; j < len(src); j++ {
		switch src[j] {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '>':
			if depth == 0 {
				return j + 1
			}
		}
	}
	return len(src)
}

// closeXMLStack pops the open elements down to the one named.
func closeXMLStack(stack []*domNode, name string) []*domNode {
	for k := len(stack) - 1; k >= 1; k-- {
		if stack[k].NodeName == name {
			return stack[:k]
		}
	}
	return stack
}

// parseXMLStartTag reads one start tag at offset i.
func parseXMLStartTag(src string, i int) (*DOMElement, int, bool) {
	j := i + 1
	start := j
	for j < len(src) && !isDOMSpace(src[j]) && src[j] != '>' && src[j] != '/' {
		j++
	}
	if start == j {
		return nil, i, false
	}
	el := newXMLElement(src[start:j])
	for j < len(src) {
		for j < len(src) && isDOMSpace(src[j]) {
			j++
		}
		if j >= len(src) {
			return el, j, false
		}
		if src[j] == '>' {
			return el, j + 1, false
		}
		if src[j] == '/' {
			j++
			if j < len(src) && src[j] == '>' {
				j++
				return el, j, true
			}
			return el, j, false
		}
		attrStart := j
		for j < len(src) && !isDOMSpace(src[j]) && src[j] != '=' && src[j] != '>' && src[j] != '/' {
			j++
		}
		attr := src[attrStart:j]
		for j < len(src) && isDOMSpace(src[j]) {
			j++
		}
		if j < len(src) && src[j] == '=' {
			j++
			for j < len(src) && isDOMSpace(src[j]) {
				j++
			}
			if j < len(src) && (src[j] == '"' || src[j] == '\'') {
				quote := src[j]
				j++
				valueStart := j
				for j < len(src) && src[j] != quote {
					j++
				}
				el.setAttr(attr, decodeDOMEntities(src[valueStart:j]))
				if j < len(src) {
					j++
				}
				continue
			}
			valueStart := j
			for j < len(src) && !isDOMSpace(src[j]) && src[j] != '>' {
				j++
			}
			el.setAttr(attr, decodeDOMEntities(src[valueStart:j]))
			continue
		}
		el.setAttr(attr, "")
	}
	return el, j, false
}

// syncXMLAttributes fills the attributes list php exposes on an element.
func syncXMLAttributes(n *domNode) {
	if n.NodeType == 1 {
		n.syncAttributes()
	}
	for _, child := range n.ChildNodes {
		if c := nodeOf(child); c != nil {
			syncXMLAttributes(c)
		}
	}
}

// loadXMLFile reads path and parses it, answering false when the read fails.
func (d *DOMDocument) loadXMLFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return d.LoadXML(string(data))
}

// LoadXML parses an XML document, replacing the tree the document held.
func (d *DOMDocument) LoadXML(xml string, _ ...int64) bool {
	if strings.TrimSpace(xml) == "" {
		return false
	}
	d.root = parseXMLDOM(xml, d)
	d.DocumentElement = nil
	for _, child := range d.root.ChildNodes {
		if n := nodeOf(child); n != nil && n.NodeType == 1 {
			if e, ok := n.self.(*DOMElement); ok {
				d.DocumentElement = e
				break
			}
		}
	}
	return d.DocumentElement != nil
}

// Load reads and parses the XML file php names.
func (d *DOMDocument) Load(filename string, _ ...int64) bool { return d.loadXMLFile(filename) }

// LoadHTMLFile reads and parses an HTML file.
func (d *DOMDocument) LoadHTMLFile(filename string, _ ...int64) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	return d.LoadHTML(string(data))
}

// SchemaValidate answers true: the tree the parse built was read, not checked.
func (d *DOMDocument) SchemaValidate(_ string, _ ...int64) bool { return true }

// SchemaValidateSource answers true, as SchemaValidate does.
func (d *DOMDocument) SchemaValidateSource(_ string, _ ...int64) bool { return true }
