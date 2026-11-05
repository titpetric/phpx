package files

import (
	"encoding/base64"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	RegisterWrapper("stdout", func(*runner.Runtime, string) (any, error) { return os.Stdout, nil })
	RegisterWrapper("stderr", func(*runner.Runtime, string) (any, error) { return os.Stderr, nil })
	scratch := func(*runner.Runtime, string) (any, error) {
		f, err := os.CreateTemp("", "phpx-stream-")
		if err != nil {
			return false, nil
		}
		_ = os.Remove(f.Name())
		return f, nil
	}
	RegisterWrapper("memory", scratch)
	RegisterWrapper("temp", scratch)
}

// phpxStreamContents is stream_get_contents with its $length and $offset, which
// upstream ignores: sebastian/diff writes php://memory and reads it back from 0.
func phpxStreamContents(stream io.Reader, rest []any) (string, error) {
	length, offset := int64(-1), int64(-1)
	if len(rest) > 0 {
		if n, ok := rest[0].(int64); ok {
			length = n
		}
	}
	if len(rest) > 1 {
		if n, ok := rest[1].(int64); ok {
			offset = n
		}
	}
	if seeker, ok := stream.(io.Seeker); ok && offset >= 0 {
		if _, err := seeker.Seek(offset, io.SeekStart); err != nil {
			return "", err
		}
	}
	if length >= 0 {
		stream = io.LimitReader(stream, length)
	}
	b, err := io.ReadAll(stream)
	return string(b), err
}

// phpxPathinfo is pathinfo() with its $flags, which upstream does not take.
func phpxPathinfo(path string, flags ...int64) any {
	parts := model.NewArray()
	// A trailing slash names the directory: pathinfo('dir/') is 'dir'.
	if trimmed := strings.TrimRight(path, "/"); trimmed != "" {
		path = trimmed
	}
	var dir, base, ext, name string
	hasDir := strings.Contains(path, "/")
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		dir, base = path[:i], path[i+1:]
		if dir == "" {
			dir = "/"
		}
	} else {
		base = path
	}
	hasExt := false
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		ext, name, hasExt = base[i+1:], base[:i], true
	} else {
		name = base
	}
	mask := int64(15)
	if len(flags) > 0 {
		mask = flags[0]
	}
	if mask == 15 {
		if hasDir {
			parts.Set("dirname", dir)
		} else {
			parts.Set("dirname", ".")
		}
		parts.Set("basename", base)
		if hasExt {
			parts.Set("extension", ext)
		}
		parts.Set("filename", name)
		return parts
	}
	switch mask {
	case 1:
		if hasDir {
			return dir
		}
		return "."
	case 2:
		return base
	case 4:
		return ext
	case 8:
		return name
	}
	return ""
}

// streamMeta is what stream_get_meta_data reports for a handle fopen made.
type streamMeta struct {
	uri, mode, wrapper, kind string
	seekable                 bool
}

// streamInfo maps a handle to its streamMeta. A *os.File is a comparable
// pointer, which is all fopen hands back.
var streamInfo sync.Map

// rememberStream records the facts fopen knew when it made h.
func rememberStream(h any, uri, mode, wrapper, kind string, seekable bool) {
	switch h.(type) {
	case *os.File:
		streamInfo.Store(h, streamMeta{uri: uri, mode: mode, wrapper: wrapper, kind: kind, seekable: seekable})
	}
}

// phpxStreamMetaData answers stream_get_meta_data() for a handle.
func phpxStreamMetaData(h any) any {
	meta, known := streamInfo.Load(h)
	m, _ := meta.(streamMeta)
	if !known {
		m = streamMeta{mode: "r+", wrapper: "plainfile", kind: "STDIO", seekable: true}
	}
	out := model.NewArray()
	out.Set("timed_out", false)
	out.Set("blocked", true)
	out.Set("eof", false)
	if m.wrapper == "PHP" || m.wrapper == "plainfile" {
		out.Set("wrapper_type", m.wrapper)
	}
	out.Set("stream_type", m.kind)
	out.Set("mode", m.mode)
	out.Set("unread_bytes", int64(0))
	out.Set("seekable", m.seekable)
	out.Set("uri", m.uri)
	return out
}

// phpxDataURI decodes a data: URI into its bytes: data://text/plain,hi and
// data:text/plain;base64,aGk=.
func phpxDataURI(filename string) (string, bool) {
	rest, ok := strings.CutPrefix(filename, "data:")
	if !ok {
		return "", false
	}
	rest = strings.TrimPrefix(rest, "//")
	head, body, found := strings.Cut(rest, ",")
	if !found {
		return "", false
	}
	if strings.HasSuffix(head, ";base64") {
		raw, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			raw, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(body, "="))
			if err != nil {
				return "", false
			}
		}
		return string(raw), true
	}
	decoded, err := url.PathUnescape(body)
	if err != nil {
		return "", false
	}
	return decoded, true
}

// phpxOpenData makes a readable handle over the bytes of a data: URI.
func phpxOpenData(filename string) (any, bool) {
	body, ok := phpxDataURI(filename)
	if !ok {
		return nil, false
	}
	f, err := os.CreateTemp("", "phpx-data-")
	if err != nil {
		return false, true
	}
	_ = os.Remove(f.Name())
	_, _ = f.WriteString(body)
	_, _ = f.Seek(0, io.SeekStart)
	rememberStream(f, filename, "r", "RFC2397", "RFC2397", true)
	return f, true
}
