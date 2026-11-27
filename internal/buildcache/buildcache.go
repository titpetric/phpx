// Package buildcache stores what a phpscript build made of a source file,
// under the user's cache directory, for a later process to read back.
//
// An entry is named by the build and by the source name. The header carries
// the content hash, the parse options and the name again. A read that finds
// one of them different answers a miss. A cache is an optimisation, and every
// failure here is a miss: the run parses the file and reports nothing.
package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// magic opens every entry and names the format a reader must know.
const magic = "phpbast1"

// Format is the entry layout. A reader meeting another number answers a miss.
const Format = 1

// maxEntry bounds a read: a corrupt file cannot cost a tree's worth of bytes.
const maxEntry = 64 << 20

// ErrMiss reports an entry that is absent or does not answer this read.
var ErrMiss = errors.New("buildcache: miss")

// DefaultDir answers the directory entries live in: $XDG_CACHE_HOME/phpx, or
// $HOME/.cache/phpx when XDG_CACHE_HOME is unset. PHPScript_CACHE overrides it.
func DefaultDir() string {
	if dir := os.Getenv("PHPScript_CACHE"); dir != "" {
		return dir
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "phpx")
}

// BuildID names the binary an entry belongs to. A dirty tree and a build
// without version information take the executable's identity instead.
func BuildID() string {
	id := "dev"
	if info, ok := buildInfo(); ok {
		if rev := info[0]; rev != "" {
			if len(rev) > 12 {
				rev = rev[:12]
			}
			id = rev
			if info[1] == "true" {
				id += dirtySuffix()
			}
			return id
		}
	}
	return id + dirtySuffix()
}

// dirtySuffix names the executable, which changes with every rebuild.
func dirtySuffix() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	st, err := os.Stat(exe)
	if err != nil {
		return ""
	}
	return "-" + strconv.FormatInt(st.ModTime().UnixNano(), 36) + "-" + strconv.FormatInt(st.Size(), 36)
}

// Store reads and writes entries in one directory. It holds no state beyond
// the counters, and two processes may use one directory at once: a write
// lands through a rename.
type Store struct {
	dir   string
	build string

	hits   atomic.Int64
	misses atomic.Int64
	writes atomic.Int64
}

// Open returns a store rooted at dir, creating it on first write.
func Open(dir, build string) (*Store, error) {
	if dir == "" {
		return nil, ErrMiss
	}
	return &Store{dir: dir, build: build}, nil
}

// Entry names what a read wants. A stored header answers only when every
// field matches.
type Entry struct {
	Name    string
	Options string
	Content [sha256.Size]byte
}

// Key names the entry of a source file by the path a process resolved.
func (s *Store) Key(name string) string {
	sum := sha256.Sum256([]byte(name))
	return s.build + "_" + hex.EncodeToString(sum[:]) + ".phpb"
}

// Load answers the payload of key when it matches want.
func (s *Store) Load(key string, want Entry) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, key))
	if err != nil || len(data) > maxEntry {
		s.misses.Add(1)
		return nil, ErrMiss
	}
	got, payload, err := split(data)
	if err != nil || got.Build != s.build || got.Name != want.Name ||
		got.Options != want.Options || got.Content != want.Content {
		s.misses.Add(1)
		return nil, ErrMiss
	}
	s.hits.Add(1)
	return payload, nil
}

// Save writes payload under key, replacing the entry there.
func (s *Store) Save(key string, ent Entry, payload []byte) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	head := header{Build: s.build, Name: ent.Name, Options: ent.Options, Content: ent.Content}
	data := append(head.header(), payload...)
	tmp, err := os.CreateTemp(s.dir, key+".tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, filepath.Join(s.dir, key)); err != nil {
		os.Remove(name)
		return err
	}
	s.writes.Add(1)
	return nil
}

// Stats reports the reads and writes this process made.
func (s *Store) Stats() (hits, misses, writes int64) {
	return s.hits.Load(), s.misses.Load(), s.writes.Load()
}

// Prune drops entries older than age, then the oldest until the directory is
// under limit bytes, and answers how many it removed.
func (s *Store) Prune(limit int64, age time.Duration) int {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0
	}
	type item struct {
		path string
		size int64
		mod  time.Time
	}
	items := make([]item, 0, len(entries))
	var total int64
	removed := 0
	cutoff := time.Now().Add(-age)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".phpb") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		total += info.Size()
		if info.ModTime().Before(cutoff) {
			if os.Remove(filepath.Join(s.dir, e.Name())) == nil {
				total -= info.Size()
				removed++
			}
			continue
		}
		items = append(items, item{path: filepath.Join(s.dir, e.Name()), size: info.Size(), mod: info.ModTime()})
	}
	for limit > 0 && total > limit && len(items) > 0 {
		oldest := 0
		for i := range items {
			if items[i].mod.Before(items[oldest].mod) {
				oldest = i
			}
		}
		if os.Remove(items[oldest].path) == nil {
			total -= items[oldest].size
			removed++
		}
		items = append(items[:oldest], items[oldest+1:]...)
	}
	return removed
}
