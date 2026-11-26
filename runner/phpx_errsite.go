//go:build phpx

package runner

import (
	"errors"
	"reflect"
	"strings"
)

// phpxSite is where a Go error first surfaced: the file and the line.
type phpxSite struct {
	file string
	line int
}

const phpxSiteLimit = 1024

// phpxNoteSite notes the site on err and on every cause it unwraps to.
func (rt *Runtime) phpxNoteSite(err error, scope *Scope) {
	if err == nil {
		return
	}
	site := phpxSite{file: rt.entrypoint, line: rt.currentSourceLine()}
	if scope != nil {
		if file, ok := scope.Get("__FILE__"); ok {
			site.file, _ = file.(string)
		}
	}
	for cause, depth := err, 0; cause != nil && depth < 32; cause, depth = errors.Unwrap(cause), depth+1 {
		rt.phpxNoteOne(cause, site)
	}
}

// phpxNoteOne keys one comparable error in the site table.
func (rt *Runtime) phpxNoteOne(err error, site phpxSite) {
	t := reflect.TypeOf(err)
	if t == nil || !t.Comparable() {
		return
	}
	if _, seen := rt.phpxSites[err]; seen {
		return
	}
	if rt.phpxSites == nil || len(rt.phpxSites) >= phpxSiteLimit {
		rt.phpxSites = map[error]phpxSite{}
	}
	rt.phpxSites[err] = site
}

// phpxSiteMethod answers getFile and getLine for a Go error that surfaced in a script.
func (rt *Runtime) phpxSiteMethod(base any, method string) (any, bool) {
	err, ok := base.(error)
	if !ok {
		return nil, false
	}
	site, ok := rt.phpxSites[err]
	if !ok {
		return nil, false
	}
	switch strings.ToLower(method) {
	case "getfile":
		return site.file, true
	case "getline":
		return int64(site.line), true
	}
	return nil, false
}
