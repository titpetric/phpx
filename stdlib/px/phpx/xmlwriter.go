//go:build phpx

package phpx

import (
	"bytes"
	"os"
	"strings"

	"github.com/titpetric/phpx/runner"
)

// XMLWriter is php's streaming XML writer, the surface phpunit and
// php-code-coverage write their reports through. The binding builds the
// document as text: startElement opens a tag, writeAttribute fills it, and the
// first content or endElement closes it, which is the order the callers use.
type XMLWriter struct {
	buf          bytes.Buffer
	stack        []xmlFrame
	open         bool
	indent       bool
	indentString string
	// atLineStart is set while the buffer ends at the start of a line, which
	// decides whether an indented element takes a newline before it.
	atLineStart bool
	// inAttribute is set by startAttribute and ended by endAttribute.
	inAttribute bool
	// uri is the file openUri named; empty writes to memory.
	uri string
	started bool
}

func init() {
	runner.RegisterArea(`Runtime\PX\XMLWriter`, registerPhpxXMLWriter)
}

// registerPhpxXMLWriter installs the writer class.
func registerPhpxXMLWriter(rt *runner.Runtime) {
	rt.RegisterConstructor("XMLWriter", func() *XMLWriter { return newXMLWriter() })
}

func newXMLWriter() *XMLWriter {
	return &XMLWriter{indentString: " ", atLineStart: true}
}

// OpenMemory directs the document to the in-memory buffer.
func (w *XMLWriter) OpenMemory() bool {
	w.uri = ""
	return true
}

// OpenUri directs the document to the file named by uri.
func (w *XMLWriter) OpenUri(uri string) bool {
	f, err := os.Create(uri)
	if err != nil {
		return false
	}
	_ = f.Close()
	w.uri = uri
	return true
}

// StartDocument writes the XML declaration.
func (w *XMLWriter) StartDocument(args ...any) bool {
	if w.started {
		return false
	}
	w.started = true
	version := "1.0"
	encoding := ""
	if len(args) > 0 && args[0] != nil {
		version, _ = args[0].(string)
	}
	if len(args) > 1 && args[1] != nil {
		encoding, _ = args[1].(string)
	}
	w.buf.WriteString(`<?xml version="` + version + `"`)
	if encoding != "" {
		w.buf.WriteString(` encoding="` + encoding + `"`)
	}
	w.buf.WriteString("?>\n")
	w.atLineStart = true
	return true
}

// EndDocument ends the document and flushes a file destination.
func (w *XMLWriter) EndDocument() bool {
	if !w.atLineStart {
		w.buf.WriteString("\n")
		w.atLineStart = true
	}
	w.started = false
	return w.flushFile()
}

// StartElement opens the element named name.
func (w *XMLWriter) StartElement(name string) bool {
	w.closeStart()
	w.lineBreak()
	w.buf.WriteString("<" + name)
	w.stack = append(w.stack, xmlFrame{name: name})
	w.open = true
	w.atLineStart = false
	return true
}

// WriteAttribute writes one attribute onto the open start tag.
func (w *XMLWriter) WriteAttribute(name, value string) bool {
	if !w.open {
		return false
	}
	w.buf.WriteString(" " + name + `="` + escapeXMLAttr(value) + `"`)
	return true
}

// StartAttribute opens an attribute whose value text() writes.
func (w *XMLWriter) StartAttribute(name string) bool {
	if !w.open {
		return false
	}
	w.buf.WriteString(" " + name + `="`)
	w.inAttribute = true
	return true
}

// EndAttribute closes an attribute startAttribute opened.
func (w *XMLWriter) EndAttribute() bool {
	if !w.inAttribute {
		return false
	}
	w.buf.WriteString(`"`)
	w.inAttribute = false
	return true
}

// EndElement closes the innermost element, self-closing it when it holds nothing.
func (w *XMLWriter) EndElement() bool {
	if len(w.stack) == 0 {
		return false
	}
	frame := w.pop()
	if w.open {
		w.buf.WriteString("/>")
		w.open = false
	} else {
		if frame.lastWasElement {
			w.lineBreak()
		}
		w.buf.WriteString("</" + frame.name + ">")
	}
	w.atLineStart = false
	return true
}

// FullEndElement closes the innermost element with an explicit end tag.
func (w *XMLWriter) FullEndElement() bool {
	if len(w.stack) == 0 {
		return false
	}
	frame := w.pop()
	if w.open {
		w.buf.WriteString(">")
		w.open = false
	} else if frame.lastWasElement {
		w.lineBreak()
	}
	w.buf.WriteString("</" + frame.name + ">")
	w.atLineStart = false
	return true
}

// WriteElement writes an element that holds content, or an empty one.
func (w *XMLWriter) WriteElement(name string, content ...string) bool {
	w.StartElement(name)
	if len(content) > 0 && content[0] != "" {
		w.Text(content[0])
	}
	return w.EndElement()
}

// Text writes escaped character data.
func (w *XMLWriter) Text(content string) bool {
	if w.inAttribute {
		w.buf.WriteString(escapeXMLAttr(content))
		return true
	}
	w.closeStart()
	w.buf.WriteString(escapeXMLText(content))
	w.atLineStart = false
	return true
}

// WriteRaw writes content with no escaping.
func (w *XMLWriter) WriteRaw(content string) bool {
	w.closeStart()
	w.buf.WriteString(content)
	w.atLineStart = false
	return true
}

// WriteComment writes a comment.
func (w *XMLWriter) WriteComment(content string) bool {
	w.closeStart()
	w.lineBreak()
	w.buf.WriteString("<!--" + content + "-->")
	w.atLineStart = false
	return true
}

// StartComment opens a comment whose text() writes.
func (w *XMLWriter) StartComment() bool {
	w.closeStart()
	w.lineBreak()
	w.buf.WriteString("<!--")
	w.atLineStart = false
	return true
}

// EndComment closes a comment startComment opened.
func (w *XMLWriter) EndComment() bool {
	w.buf.WriteString("-->")
	w.atLineStart = false
	return true
}

// StartCdata opens a CDATA section.
func (w *XMLWriter) StartCdata() bool {
	w.closeStart()
	w.buf.WriteString("<![CDATA[")
	w.atLineStart = false
	return true
}

// EndCdata closes a CDATA section.
func (w *XMLWriter) EndCdata() bool {
	w.buf.WriteString("]]>")
	w.atLineStart = false
	return true
}

// WriteCdata writes a complete CDATA section.
func (w *XMLWriter) WriteCdata(content string) bool {
	w.closeStart()
	w.buf.WriteString("<![CDATA[" + content + "]]>")
	w.atLineStart = false
	return true
}

// SetIndent turns pretty printing on or off.
func (w *XMLWriter) SetIndent(on bool) bool {
	w.indent = on
	return true
}

// SetIndentString sets the string one indent level writes.
func (w *XMLWriter) SetIndentString(s string) bool {
	w.indentString = s
	return true
}

// OutputMemory answers the document, and empties the buffer when flush is set.
func (w *XMLWriter) OutputMemory(flush ...bool) string {
	out := w.buf.String()
	if len(flush) == 0 || flush[0] {
		w.buf.Reset()
	}
	return out
}

// Flush writes the buffer to the file destination and answers what it wrote.
func (w *XMLWriter) Flush(empty ...bool) string {
	w.flushFile()
	out := w.buf.String()
	if len(empty) == 0 || empty[0] {
		w.buf.Reset()
	}
	return out
}

// xmlFrame is one open element: its name and whether its last child was an
// element, which decides where an indented closing tag goes.
type xmlFrame struct {
	name           string
	lastWasElement bool
}

// pop removes the innermost frame, noting on its parent that an element ended.
func (w *XMLWriter) pop() xmlFrame {
	frame := w.stack[len(w.stack)-1]
	w.stack = w.stack[:len(w.stack)-1]
	if len(w.stack) > 0 {
		w.stack[len(w.stack)-1].lastWasElement = true
	}
	return frame
}

// lineBreak starts an indented element on its own line at its depth.
func (w *XMLWriter) lineBreak() {
	if !w.indent {
		return
	}
	if !w.atLineStart {
		w.buf.WriteString("\n")
	}
	w.buf.WriteString(strings.Repeat(w.indentString, len(w.stack)))
}

// closeStart ends the open start tag so content can follow it.
func (w *XMLWriter) closeStart() {
	if !w.open {
		return
	}
	w.buf.WriteString(">")
	w.open = false
}

// flushFile appends the buffer to the file openUri named.
func (w *XMLWriter) flushFile() bool {
	if w.uri == "" {
		return true
	}
	f, err := os.OpenFile(w.uri, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return false
	}
	defer f.Close()
	if _, err := f.WriteString(w.buf.String()); err != nil {
		return false
	}
	w.buf.Reset()
	return true
}

// escapeXMLText escapes the characters that cannot stand in character data.
func escapeXMLText(s string) string { return xmlTextEscaper.Replace(s) }

// escapeXMLAttr escapes a value that stands in a double-quoted attribute.
func escapeXMLAttr(s string) string { return xmlAttrEscaper.Replace(s) }

var xmlTextEscaper = strings.NewReplacer(
	"&", "&amp;", "<", "&lt;", ">", "&gt;", "\r", "&#13;",
)

var xmlAttrEscaper = strings.NewReplacer(
	"&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "\r", "&#13;", "\n", "&#10;", "\t", "&#9;",
)
