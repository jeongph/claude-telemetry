package gitinfo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGatherInGitRepo(t *testing.T) {
	// The test runs inside the claude-telemetry repo itself
	cwd, _ := os.Getwd()
	root := filepath.Join(cwd, "..", "..")
	info := Gather(root, "")
	if !info.IsRepo {
		t.Error("should be a git repo")
	}
	if info.Branch == "" {
		t.Error("branch should not be empty")
	}
}

func TestGatherInNonGitDir(t *testing.T) {
	info := Gather(t.TempDir(), "")
	if info.IsRepo {
		t.Error("temp dir should not be a git repo")
	}
	if info.Branch != "" {
		t.Error("branch should be empty for non-git")
	}
}

func TestCacheWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	info := &GitInfo{Branch: "main", IsRepo: true}
	writeCache(dir, "testhash", info)
	cached, ok := readCache(dir, "testhash", 5*time.Second)
	if !ok {
		t.Fatal("cache should hit")
	}
	if cached.Branch != "main" {
		t.Errorf("cached branch = %q", cached.Branch)
	}
}

func TestCacheExpired(t *testing.T) {
	dir := t.TempDir()
	info := &GitInfo{Branch: "main", IsRepo: true}
	writeCache(dir, "testhash", info)
	_, ok := readCache(dir, "testhash", 0) // TTL=0 → expired
	if ok {
		t.Error("cache should be expired")
	}
}

func TestCacheCorrupted(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "badhash.json"), []byte("not json"), 0644) //nolint:errcheck
	_, ok := readCache(dir, "badhash", 5*time.Second)
	if ok {
		t.Error("corrupted cache should not hit")
	}
}

func TestRepoHash(t *testing.T) {
	h := repoHash("/some/path")
	if len(h) != 12 {
		t.Errorf("hash length = %d, want 12", len(h))
	}
}

func TestParseDiffNumstat(t *testing.T) {
	input := "10\t5\tfile1.go\n3\t1\tfile2.go\n"
	a, d := parseDiffNumstat(input)
	if a != 13 {
		t.Errorf("added = %d, want 13", a)
	}
	if d != 6 {
		t.Errorf("deleted = %d, want 6", d)
	}
}
