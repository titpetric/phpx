//go:build !phpx

package runner

// phpxSite is absent in the untagged build; the field only needs a type.
type phpxSite struct{}

// phpxNoteSite answers the untagged build, which records no site.
func (rt *Runtime) phpxNoteSite(_ error, _ *Scope) {}

// phpxSiteMethod answers the untagged build, which knows no site.
func (rt *Runtime) phpxSiteMethod(_ any, _ string) (any, bool) { return nil, false }
