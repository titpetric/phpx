//go:build phpx

package runner

import (
	"reflect"
	"strings"
)

// phpxSite is where a Go error first surfaced in a script: the file of the
// frame and the line of the statement that failed.
type phpxSite struct {
	file string
	line int
}

// phpxSiteLimit bounds the table: an error nobody asks about stays in it
// until the table is dropped whole.
const phpxSiteLimit = 1024

// phpxNoteSite records the site of an error a statement returned, once. The
// innermost statement sees it first, which is the line php reports. Only
// pointer errors are keyed; a value error could be unhashable.
func (rt *Runtime) phpxNoteSite(err error, scope *Scope) {
	if reflect.TypeOf(err).Kind() != reflect.Pointer {
		return
	}
	if _, seen := rt.phpxSites[err]; seen {
		return
	}
	if rt.phpxSites == nil || len(rt.phpxSites) >= phpxSiteLimit {
		rt.phpxSites = map[error]phpxSite{}
	}
	site := phpxSite{file: rt.entrypoint, line: rt.currentSourceLine()}
	if scope != nil {
		if file, ok := scope.Get("__FILE__"); ok {
			site.file, _ = file.(string)
		}
	}
	rt.phpxSites[err] = site
}

// phpxSiteMethod answers getFile and getLine for a Go error that surfaced in a
// script, ahead of throwableMethod's empty answer.
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
