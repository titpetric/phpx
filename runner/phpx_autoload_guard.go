//go:build phpx

package runner

import (
	"strings"
	"sync"
)

// php does not autoload a class that is already being autoloaded: the nested
// request answers "not found". Without that a loader registered as the string
// 'Class::method' for a class the loader itself would have to load asks for
// that class again, forever.
var phpxLoading struct {
	sync.Mutex
	active map[*Runtime]map[string]bool
}

// phpxEnterAutoload marks class as loading on rt and reports false when it
// already is. The release function clears the mark.
func phpxEnterAutoload(rt *Runtime, class string) (release func(), ok bool) {
	key := strings.ToLower(strings.TrimPrefix(class, "\\"))
	phpxLoading.Lock()
	defer phpxLoading.Unlock()
	if phpxLoading.active == nil {
		phpxLoading.active = map[*Runtime]map[string]bool{}
	}
	set := phpxLoading.active[rt]
	if set == nil {
		set = map[string]bool{}
		phpxLoading.active[rt] = set
	}
	if set[key] {
		return func() {}, false
	}
	set[key] = true
	return func() {
		phpxLoading.Lock()
		defer phpxLoading.Unlock()
		delete(set, key)
		if len(set) == 0 {
			delete(phpxLoading.active, rt)
		}
	}, true
}
