package runner

import (
	"crypto/sha256"
	"strconv"

	"github.com/titpetric/phpx/internal/buildcache"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/model/astcodec"
)

// A runtime reads a file's parsed program out of the build cache when a
// previous process parsed the same bytes under the same options and build. The
// entry holds the whole AST, so the lexer, the parser and their allocations do
// not run at all; flatstack compiles the decoded program as it compiles a
// parsed one.
//
// The cache never decides correctness. Every failure - an absent entry, a
// different build, a content hash that moved, bytes that do not decode - falls
// back to the parse, and a parse whose payload cannot be encoded writes no
// entry.

// openBuildCache returns the store for dir, nil when the host named none.
func openBuildCache(dir string) *buildcache.Store {
	if dir == "" {
		return nil
	}
	store, err := buildcache.Open(dir, buildcache.BuildID())
	if err != nil {
		return nil
	}
	return store
}

// cacheOptions names the parse inputs the source does not hold. Two runtimes
// parsing one file under different options write to separate keys.
func (rt *Runtime) cacheOptions() string {
	return "go:" + strconv.FormatBool(rt.opts.GoRuntime())
}

// cacheKey names the entry of a source file under this runtime's options.
func (rt *Runtime) cacheKey(name string) string {
	return rt.buildCache.Key(name + "\x00" + rt.cacheOptions())
}

// loadCached answers the program a previous process parsed from src, nil on
// any miss.
func (rt *Runtime) loadCached(name string, src []byte) *model.Program {
	if rt.buildCache == nil {
		return nil
	}
	ent := buildcache.Entry{Name: name, Options: rt.cacheOptions(), Content: sha256.Sum256(src)}
	payload, err := rt.buildCache.Load(rt.cacheKey(name), ent)
	if err != nil {
		return nil
	}
	prog, err := astcodec.Decode(payload)
	if err != nil {
		return nil
	}
	return prog
}

// storeCached writes prog under the key of its source, for a later process.
func (rt *Runtime) storeCached(name string, src []byte, prog *model.Program) {
	if rt.buildCache == nil {
		return
	}
	payload, err := astcodec.Encode(prog)
	if err != nil {
		return
	}
	ent := buildcache.Entry{Name: name, Options: rt.cacheOptions(), Content: sha256.Sum256(src)}
	_ = rt.buildCache.Save(rt.cacheKey(name), ent, payload)
}

// BuildCacheStats reports the reads and writes this runtime made: hits, misses
// and entries written. A runtime with no cache reports zero for each.
func (rt *Runtime) BuildCacheStats() (hits, misses, writes int64) {
	if rt.buildCache == nil {
		return 0, 0, 0
	}
	return rt.buildCache.Stats()
}
