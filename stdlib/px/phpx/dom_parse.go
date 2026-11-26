//go:build phpx

package phpx

import (
	"sort"
	"strconv"
	"strings"
)

// The parse is the one libxml performs for the HTML the console components
// build: elements, attributes, text and comments, with void elements and the
// closing rules the generated markup never needs beyond matching an end tag.
// Source no caller here writes (a script block, a CDATA section, a namespace)
// parses as text instead of failing: the renderer reads its own output.

// voidElements close at their start tag, as the HTML parser closes them.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// parseDOM builds the tree for html, owned by doc.
func parseDOM(html string, doc *DOMDocument) *domNode {
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

	for i := 0; i < len(html); {
		if html[i] != '<' {
			end := strings.IndexByte(html[i:], '<')
			if end < 0 {
				end = len(html) - i
			}
			text := decodeDOMEntities(html[i : i+end])
			if strings.TrimSpace(text) != "" {
				add(newDOMText(text))
			}
			i += end
			continue
		}
		switch {
		case strings.HasPrefix(html[i:], "<!--"):
			end := strings.Index(html[i+4:], "-->")
			if end < 0 {
				end = len(html) - i - 4
			}
			add(newDOMComment(html[i+4 : i+4+end]))
			i += 4 + end
			if strings.HasPrefix(html[i:], "-->") {
				i += 3
			}
		case strings.HasPrefix(html[i:], "<!"), strings.HasPrefix(html[i:], "<?"):
			if strings.HasPrefix(html[i:], "<?") {
				if end := strings.Index(html[i:], "?>"); end >= 0 {
					i += end + 2
					continue
				}
			}
			if end := strings.IndexByte(html[i:], '>'); end >= 0 {
				i += end + 1
				continue
			}
			i = len(html)
		case strings.HasPrefix(html[i:], "</"):
			if end := strings.IndexByte(html[i:], '>'); end >= 0 {
				name := strings.ToLower(strings.TrimSpace(html[i+2 : i+end]))
				stack = closeDOMElement(stack, name)
				i += end + 1
				continue
			}
			i = len(html)
		default:
			el, next := parseDOMStartTag(html, i)
			if el == nil {
				// A '<' that opens no tag is text.
				add(newDOMText("<"))
				i++
				continue
			}
			n := add(el)
			if !voidElements[n.NodeName] {
				stack = append(stack, n)
			}
			i = next
		}
	}

	reflowDOM(root)
	return root
}

// parseDOMStartTag reads one start tag at offset i.
func parseDOMStartTag(html string, i int) (*DOMElement, int) {
	j := i + 1
	if j >= len(html) {
		return nil, i
	}
	start := j
	for j < len(html) && !isDOMSpace(html[j]) && html[j] != '>' && html[j] != '/' {
		j++
	}
	if start == j {
		return nil, i
	}
	el := newDOMElement(html[start:j], "")
	for j < len(html) {
		for j < len(html) && isDOMSpace(html[j]) {
			j++
		}
		if j >= len(html) {
			el.syncAttributes()
			return el, j
		}
		if html[j] == '>' {
			el.syncAttributes()
			return el, j + 1
		}
		if html[j] == '/' {
			j++
			if j < len(html) && html[j] == '>' {
				j++
			}
			el.syncAttributes()
			return el, j
		}
		attrStart := j
		for j < len(html) && !isDOMSpace(html[j]) && html[j] != '=' && html[j] != '>' && html[j] != '/' {
			j++
		}
		attr := strings.ToLower(html[attrStart:j])
		for j < len(html) && isDOMSpace(html[j]) {
			j++
		}
		if j < len(html) && html[j] == '=' {
			j++
			for j < len(html) && isDOMSpace(html[j]) {
				j++
			}
			if j < len(html) && (html[j] == '"' || html[j] == '\'') {
				quote := html[j]
				j++
				valueStart := j
				for j < len(html) && html[j] != quote {
					j++
				}
				el.setAttr(attr, decodeDOMEntities(html[valueStart:j]))
				if j < len(html) {
					j++
				}
				continue
			}
			valueStart := j
			for j < len(html) && !isDOMSpace(html[j]) && html[j] != '>' {
				j++
			}
			el.setAttr(attr, decodeDOMEntities(html[valueStart:j]))
			continue
		}
		el.setAttr(attr, "")
	}
	el.syncAttributes()
	return el, j
}

// closeDOMElement pops the open elements down to the one named.
func closeDOMElement(stack []*domNode, name string) []*domNode {
	for k := len(stack) - 1; k >= 1; k-- {
		if stack[k].NodeName == name {
			return stack[:k]
		}
	}
	return stack
}

// reflowDOM fills first/last child and the sibling links after the parse.
func reflowDOM(n *domNode) {
	for _, child := range n.ChildNodes {
		c := nodeOf(child)
		if c == nil {
			continue
		}
		reflowDOM(c)
		c.NodeValue = domTextContent(c)
		c.TextContent = c.NodeValue
	}
	n.reflow()
}

// domTextContent is the text php answers for a node's value: a text or a
// comment carries its own, an element the concatenation of its descendants'.
func domTextContent(n *domNode) string {
	switch n.NodeType {
	case 3, 8:
		return n.NodeValue
	}
	var b strings.Builder
	for _, child := range n.ChildNodes {
		if c := nodeOf(child); c != nil {
			b.WriteString(domTextContent(c))
		}
	}
	return b.String()
}

// saveDOMNode serialises one node as saveXML does.
func saveDOMNode(n *domNode) string {
	if n == nil {
		return ""
	}
	switch n.NodeType {
	case 3:
		return escapeDOMText(n.NodeValue)
	case 8:
		return "<!--" + n.NodeValue + "-->"
	}
	var b strings.Builder
	b.WriteString("<")
	b.WriteString(n.NodeName)
	names := make([]string, 0, len(n.Attrs))
	for name := range n.Attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		b.WriteString(" ")
		b.WriteString(name)
		b.WriteString(`="`)
		b.WriteString(escapeDOMAttr(n.Attrs[name]))
		b.WriteString(`"`)
	}
	if voidElements[n.NodeName] && len(n.ChildNodes) == 0 {
		b.WriteString("/>")
		return b.String()
	}
	b.WriteString(">")
	for _, child := range n.ChildNodes {
		b.WriteString(saveDOMNode(nodeOf(child)))
	}
	b.WriteString("</")
	b.WriteString(n.NodeName)
	b.WriteString(">")
	return b.String()
}

func isDOMSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

// escapeDOMText escapes a text node the way saveXML does.
func escapeDOMText(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

// escapeDOMAttr escapes an attribute value the way saveXML does.
func escapeDOMAttr(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

// domEntities maps the named references the components emit.
var domEntities = map[string]string{
	"amp": "&", "lt": "<", "gt": ">", "quot": "\"", "apos": "'",
	"nbsp": " ", "copy": "©", "reg": "®", "trade": "™",
	"hellip": "…", "mdash": "—", "ndash": "–",
	"lsquo": "‘", "rsquo": "’", "ldquo": "“", "rdquo": "”",
	"laquo": "«", "raquo": "»", "times": "×", "middot": "·",
	"euro": "€", "pound": "£", "yen": "¥",
}

// decodeDOMEntities resolves character references in text and attributes.
func decodeDOMEntities(s string) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		end := strings.IndexByte(s[i:], ';')
		if end < 0 || end > 12 {
			b.WriteByte('&')
			i++
			continue
		}
		ref := s[i+1 : i+end]
		if strings.HasPrefix(ref, "#") {
			value := ref[1:]
			base := 10
			if strings.HasPrefix(value, "x") || strings.HasPrefix(value, "X") {
				value = value[1:]
				base = 16
			}
			if n, err := strconv.ParseInt(value, base, 32); err == nil && n >= 0 {
				b.WriteRune(rune(n))
				i += end + 1
				continue
			}
		} else if text, ok := domEntities[ref]; ok {
			b.WriteString(text)
			i += end + 1
			continue
		}
		b.WriteByte('&')
		i++
	}
	return b.String()
}
