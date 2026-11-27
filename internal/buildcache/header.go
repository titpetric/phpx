package buildcache

import (
	"crypto/sha256"
	"encoding/binary"
	"runtime/debug"
)

// header holds everything a reader checks before it trusts a payload.
type header struct {
	Build   string
	Name    string
	Options string
	Content [sha256.Size]byte
}

// header writes the entry header and answers the bytes that open an entry.
func (h header) header() []byte {
	buf := make([]byte, 0, 64+len(h.Build)+len(h.Name)+len(h.Options))
	buf = append(buf, magic...)
	buf = binary.AppendUvarint(buf, Format)
	buf = appendString(buf, h.Build)
	buf = appendString(buf, h.Name)
	buf = appendString(buf, h.Options)
	buf = append(buf, h.Content[:]...)
	return buf
}

func appendString(buf []byte, s string) []byte {
	buf = binary.AppendUvarint(buf, uint64(len(s)))
	return append(buf, s...)
}

// split reads an entry, answering its header and the payload behind it.
func split(data []byte) (header, []byte, error) {
	var h header
	if len(data) < len(magic) || string(data[:len(magic)]) != magic {
		return h, nil, ErrMiss
	}
	data = data[len(magic):]
	format, n := binary.Uvarint(data)
	if n <= 0 || format != Format {
		return h, nil, ErrMiss
	}
	data = data[n:]
	var ok bool
	if h.Build, data, ok = readString(data); !ok {
		return h, nil, ErrMiss
	}
	if h.Name, data, ok = readString(data); !ok {
		return h, nil, ErrMiss
	}
	if h.Options, data, ok = readString(data); !ok {
		return h, nil, ErrMiss
	}
	if len(data) < sha256.Size {
		return h, nil, ErrMiss
	}
	copy(h.Content[:], data[:sha256.Size])
	return h, data[sha256.Size:], nil
}

func readString(data []byte) (string, []byte, bool) {
	n, size := binary.Uvarint(data)
	if size <= 0 || n > uint64(len(data)-size) {
		return "", nil, false
	}
	return string(data[size : size+int(n)]), data[size+int(n):], true
}

// buildInfo answers the revision and dirty marker a go build stamped in.
func buildInfo() ([2]string, bool) {
	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return [2]string{}, false
	}
	var out [2]string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			out[0] = s.Value
		case "vcs.modified":
			out[1] = s.Value
		}
	}
	return out, true
}
