package buildcache

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir(), "abc123def456")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func entry(name string, content string) Entry {
	return Entry{Name: name, Options: "go=false", Content: sha256.Sum256([]byte(content))}
}

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	ent := entry("/src/a.php", "<?php echo 1;")
	key := s.Key(ent.Name)
	if _, err := s.Load(key, ent); err == nil {
		t.Fatal("Load on an empty store answered an entry")
	}
	if err := s.Save(key, ent, []byte("payload")); err != nil {
		t.Fatal(err)
	}
	got, err := s.Load(key, ent)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload" {
		t.Fatalf("Load = %q, want payload", got)
	}
	hits, misses, writes := s.Stats()
	if hits != 1 || misses != 1 || writes != 1 {
		t.Fatalf("Stats = %d/%d/%d, want 1/1/1", hits, misses, writes)
	}
}

func TestStoreRefusesAnotherContent(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	written := entry("/src/a.php", "<?php echo 1;")
	key := s.Key(written.Name)
	if err := s.Save(key, written, []byte("payload")); err != nil {
		t.Fatal(err)
	}
	changed := written
	changed.Content = sha256.Sum256([]byte("<?php echo 2;"))
	if _, err := s.Load(key, changed); err == nil {
		t.Fatal("Load answered an entry written for other content")
	}
}

func TestStoreRefusesAnotherNameAndBuild(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	ent := entry("/src/a.php", "<?php echo 1;")
	if err := s.Save(s.Key(ent.Name), ent, []byte("payload")); err != nil {
		t.Fatal(err)
	}
	other := ent
	other.Name = "/src/b.php"
	if _, err := s.Load(s.Key(ent.Name), other); err == nil {
		t.Fatal("Load answered an entry written for another name")
	}
	next, err := Open(s.dir, "anotherbuild")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := next.Load(next.Key(ent.Name), ent); err == nil {
		t.Fatal("Load answered an entry written by another build")
	}
}

func TestStoreRefusesCorruptEntry(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	ent := entry("/src/a.php", "<?php echo 1;")
	key := s.Key(ent.Name)
	if err := s.Save(key, ent, []byte("payload")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.dir, key), []byte("not an entry"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(key, ent); err == nil {
		t.Fatal("Load answered a corrupt entry")
	}
}

func TestKeyNamesBuildAndPath(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	key := s.Key("/src/a.php")
	if !strings.HasPrefix(key, "abc123def456_") || !strings.HasSuffix(key, ".phpb") {
		t.Fatalf("Key = %q, want the build prefix and a .phpb suffix", key)
	}
	if len(key) != len("abc123def456_")+sha256.Size*2+len(".phpb") {
		t.Fatalf("Key = %q, want one sha256 of the name", key)
	}
	if s.Key("/src/b.php") == key {
		t.Fatal("two names answered one key")
	}
}

func TestPruneDropsOldEntries(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	old := entry("/src/old.php", "old")
	fresh := entry("/src/new.php", "new")
	if err := s.Save(s.Key(old.Name), old, make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(s.Key(fresh.Name), fresh, make([]byte, 64)); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(filepath.Join(s.dir, s.Key(old.Name)), past, past); err != nil {
		t.Fatal(err)
	}
	if removed := s.Prune(0, 24*time.Hour); removed != 1 {
		t.Fatalf("Prune removed %d, want 1", removed)
	}
	if _, err := s.Load(s.Key(fresh.Name), fresh); err != nil {
		t.Fatalf("Prune removed the fresh entry: %v", err)
	}
}

func TestPruneHoldsTheSizeLimit(t *testing.T) {
	t.Parallel()
	s := testStore(t)
	for _, name := range []string{"/a.php", "/b.php", "/c.php"} {
		ent := entry(name, name)
		if err := s.Save(s.Key(ent.Name), ent, make([]byte, 100)); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatal(err)
	}
	one := int64(0)
	for _, e := range entries {
		info, _ := e.Info()
		one = max(one, info.Size())
	}
	if removed := s.Prune(one, time.Hour); removed == 0 {
		t.Fatal("Prune removed nothing over the limit")
	}
	total := int64(0)
	names, _ := os.ReadDir(s.dir)
	for _, e := range names {
		info, _ := e.Info()
		total += info.Size()
	}
	if total > one {
		t.Fatalf("Prune left %d bytes over a %d byte limit", total, one)
	}
}

func TestBuildIDNamesTheBuild(t *testing.T) {
	t.Parallel()
	id := BuildID()
	if id == "" {
		t.Fatal("BuildID = \"\"")
	}
	if strings.ContainsAny(id, "/ ") {
		t.Fatalf("BuildID = %q, want a name a file can hold", id)
	}
}
