package files

import (
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/phpx/runner"
)

// registerReads installs the read side. Each of these looks in the runtime's
// source filesystem first, the tree an embedded or in-memory application
// ships, and falls back to the host filesystem for what only exists there, such
// as a file an earlier write produced.
func registerReads(rt *runner.Runtime, r root) {
	// glob returns the paths matching $pattern, searched in the source
	// filesystem when the runtime has one, otherwise on the host; a pattern
	// naming anything outside the root matches nothing, and a malformed one
	// matches nothing and not failing, as PHP's does.
	globOne := func(pattern string) []string {
		source := r.sourceFS()
		if source == nil {
			matches, err := filepath.Glob(r.resolve(pattern))
			if err != nil {
				return nil
			}
			return r.globSpelling(pattern, r.unresolve(matches))
		}
		name, ok := r.fsPath(pattern)
		if !ok {
			return nil
		}
		matches, err := iofs.Glob(source, name)
		if err != nil {
			return nil
		}
		return r.globSpelling(pattern, matches)
	}
	// glob takes php's flags: GLOB_BRACE expands {a,b}, GLOB_ONLYDIR keeps the
	// directories, GLOB_MARK ends them with a slash, GLOB_NOCHECK answers the
	// pattern itself when nothing matches.
	rt.RegisterFunc("glob", func(pattern string, flags ...int64) []string {
		var mode int64
		if len(flags) > 0 {
			mode = flags[0]
		}
		patterns := []string{pattern}
		if mode&1024 != 0 {
			patterns = globBraces(pattern)
		}
		var out []string
		for _, one := range patterns {
			for _, match := range globOne(one) {
				if mode&(8192|2) != 0 {
					info, err := os.Stat(r.resolve(match))
					isDir := err == nil && info.IsDir()
					if mode&8192 != 0 && !isDir {
						continue
					}
					if mode&2 != 0 && isDir {
						match += "/"
					}
				}
				out = append(out, match)
			}
		}
		if len(out) == 0 && mode&16 != 0 {
			return []string{pattern}
		}
		return out
	})

	// file_get_contents returns the contents of $filename as a string, or
	// false on failure; php://input is the raw request body (stdin under the
	// cli SAPI), and it looks for a relative path in the source filesystem
	// first, then on the host.
	rt.RegisterFunc("readlink", func(path string) any {
		target, err := os.Readlink(r.resolve(path))
		if err != nil {
			return false
		}
		return target
	})
	rt.RegisterFunc("file_get_contents", func(filename string) any {
		if body, ok := phpxDataURI(filename); ok {
			return body
		}
		if scheme, ok := strings.CutPrefix(filename, "php://"); ok {
			if scheme == "input" {
				b, err := io.ReadAll(inputSource(rt))
				if err != nil {
					return false
				}
				return string(b)
			}
			return false
		}
		if name, ok := r.uploadPath(filename); ok {
			b, err := os.ReadFile(name)
			if err != nil {
				return false
			}
			return string(b)
		}
		if source := r.sourceFS(); source != nil {
			if name, ok := r.fsPath(filename); ok {
				if b, err := iofs.ReadFile(source, name); err == nil {
					return string(b)
				}
			}
		}
		b, err := os.ReadFile(r.resolve(filename))
		if err != nil {
			return false
		}
		return string(b)
	})
	// file_exists reports whether $filename exists, in the source filesystem or on the host.
	rt.RegisterFunc("file_exists", func(filename string) bool {
		if name, ok := r.uploadPath(filename); ok {
			_, err := os.Stat(name)
			return err == nil
		}
		if _, err := r.statSource(filename); err == nil {
			return true
		}
		_, err := os.Stat(r.resolve(filename))
		return err == nil
	})
	// filemtime returns the modification time of $filename as a Unix
	// timestamp, or 0 when nothing is there, where PHP returns false.
	rt.RegisterFunc("filemtime", func(filename string) int64 {
		if name, ok := r.uploadPath(filename); ok {
			if st, err := os.Stat(name); err == nil {
				return st.ModTime().Unix()
			}
			return 0
		}
		if st, err := r.statSource(filename); err == nil {
			return st.ModTime().Unix()
		}
		if st, err := os.Stat(r.resolve(filename)); err == nil {
			return st.ModTime().Unix()
		}
		return 0
	})
}

// sourceFS is the read-only filesystem the script itself loads from. A
// host that runs a script straight off disk leaves it unset and every read
// through it has to tolerate its absence.
func (r root) sourceFS() iofs.FS {
	return r.rt.FS()
}

// statSource stats a script-supplied path in the source filesystem, and reports
// the absence of that filesystem, or of a way to name the path inside it, as a
// miss. A miss is not an answer: every caller falls through to the host.
func (r root) statSource(p string) (iofs.FileInfo, error) {
	source := r.sourceFS()
	if source == nil {
		return nil, iofs.ErrNotExist
	}
	name, ok := r.fsPath(p)
	if !ok {
		return nil, iofs.ErrNotExist
	}
	return iofs.Stat(source, name)
}

// globBraces expands the first {a,b} group of a pattern, and the groups the
// results still carry, in the order php's GLOB_BRACE yields them.
func globBraces(pattern string) []string {
	open := strings.IndexByte(pattern, '{')
	if open < 0 {
		return []string{pattern}
	}
	depth := 0
	for i := open; i < len(pattern); i++ {
		switch pattern[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				var out []string
				for _, alt := range splitBraceAlternatives(pattern[open+1 : i]) {
					out = append(out, globBraces(pattern[:open]+alt+pattern[i+1:])...)
				}
				return out
			}
		}
	}
	return []string{pattern}
}

// splitBraceAlternatives splits a brace body on the commas outside a nested group.
func splitBraceAlternatives(body string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, body[start:i])
				start = i + 1
			}
		}
	}
	return append(parts, body[start:])
}
