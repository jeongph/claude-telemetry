package gitinfo

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GitInfo holds gathered git repository state.
type GitInfo struct {
	IsRepo    bool   `json:"is_repo"`
	Branch    string `json:"branch"`
	Ahead     int    `json:"ahead"`
	Behind    int    `json:"behind"`
	Added     int    `json:"added"`
	Deleted   int    `json:"deleted"`
	Untracked int    `json:"untracked"`
	Stash     int    `json:"stash"`
	Worktrees int    `json:"worktrees"`
}

// cacheEntry wraps GitInfo with a timestamp for TTL checks.
type cacheEntry struct {
	Info *GitInfo  `json:"info"`
	Ts   time.Time `json:"ts"`
}

const defaultCacheTTL = 5 * time.Second

// Gather is the main entry point. It checks the file-based cache first,
// and if there's a miss it runs git commands in parallel.
// cacheDir defaults to ~/.claude/statusline/cache/ if empty.
func Gather(cwd, cacheDir string) *GitInfo {
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cacheDir = filepath.Join(home, ".claude", "statusline", "cache")
		}
	}

	hash := repoHash(cwd)

	// cache hit?
	if cached, ok := readCache(cacheDir, hash, defaultCacheTTL); ok {
		return cached
	}

	info := gather(cwd)

	if cacheDir != "" {
		writeCache(cacheDir, hash, info)
	}

	return info
}

// gather runs all git commands and populates a GitInfo.
func gather(cwd string) *GitInfo {
	// First check whether this is a git repo at all.
	if err := gitCmd(cwd, "rev-parse", "--git-dir"); err != nil {
		return &GitInfo{IsRepo: false}
	}

	info := &GitInfo{IsRepo: true}
	var mu sync.Mutex
	var wg sync.WaitGroup

	run := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	// Branch
	run(func() {
		branch := gitOutput(cwd, "rev-parse", "--abbrev-ref", "HEAD")
		mu.Lock()
		info.Branch = strings.TrimSpace(branch)
		mu.Unlock()
	})

	// Ahead / Behind (fails silently if no upstream)
	run(func() {
		out := gitOutput(cwd, "rev-list", "--left-right", "--count", "HEAD...@{u}")
		parts := strings.Fields(strings.TrimSpace(out))
		if len(parts) == 2 {
			ahead, _ := strconv.Atoi(parts[0])
			behind, _ := strconv.Atoi(parts[1])
			mu.Lock()
			info.Ahead = ahead
			info.Behind = behind
			mu.Unlock()
		}
	})

	// Added / Deleted (staged + unstaged diff)
	run(func() {
		unstaged := gitOutput(cwd, "diff", "--numstat")
		staged := gitOutput(cwd, "diff", "--cached", "--numstat")
		a1, d1 := parseDiffNumstat(unstaged)
		a2, d2 := parseDiffNumstat(staged)
		mu.Lock()
		info.Added = a1 + a2
		info.Deleted = d1 + d2
		mu.Unlock()
	})

	// Untracked
	run(func() {
		out := gitOutput(cwd, "ls-files", "--others", "--exclude-standard")
		count := 0
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if line != "" {
				count++
			}
		}
		mu.Lock()
		info.Untracked = count
		mu.Unlock()
	})

	// Stash
	run(func() {
		out := gitOutput(cwd, "stash", "list")
		count := 0
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if line != "" {
				count++
			}
		}
		mu.Lock()
		info.Stash = count
		mu.Unlock()
	})

	// Worktrees (minus 1 for main)
	run(func() {
		out := gitOutput(cwd, "worktree", "list")
		count := 0
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if line != "" {
				count++
			}
		}
		if count > 0 {
			count-- // subtract main worktree
		}
		mu.Lock()
		info.Worktrees = count
		mu.Unlock()
	})

	wg.Wait()
	return info
}

// gitCmd runs a git command and returns only the error status.
func gitCmd(cwd string, args ...string) error {
	fullArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// gitOutput runs a git command and returns its stdout as a string.
// Returns empty string on error.
func gitOutput(cwd string, args ...string) string {
	fullArgs := append([]string{"--no-optional-locks", "-C", cwd}, args...)
	cmd := exec.Command("git", fullArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return ""
	}
	return out.String()
}

// repoHash returns the first 12 hex characters of the SHA-256 hash of the path.
func repoHash(cwd string) string {
	h := sha256.Sum256([]byte(cwd))
	return fmt.Sprintf("%x", h[:])[:12]
}

// parseDiffNumstat parses the output of `git diff --numstat` and returns
// the total added and deleted line counts.
func parseDiffNumstat(s string) (added, deleted int) {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		a, _ := strconv.Atoi(parts[0])
		d, _ := strconv.Atoi(parts[1])
		added += a
		deleted += d
	}
	return
}

// readCache attempts to read a cached GitInfo from disk.
// Returns (nil, false) on any failure or if the entry is expired.
func readCache(cacheDir, hash string, ttl time.Duration) (*GitInfo, bool) {
	if cacheDir == "" {
		return nil, false
	}
	path := filepath.Join(cacheDir, hash+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if time.Since(entry.Ts) > ttl {
		return nil, false
	}
	return entry.Info, true
}

// writeCache writes a GitInfo to disk atomically (write to .tmp, then rename).
func writeCache(cacheDir, hash string, info *GitInfo) {
	if cacheDir == "" {
		return
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return
	}
	entry := cacheEntry{Info: info, Ts: time.Now()}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	finalPath := filepath.Join(cacheDir, hash+".json")
	tmpPath := finalPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return
	}
	os.Rename(tmpPath, finalPath) //nolint:errcheck
}
