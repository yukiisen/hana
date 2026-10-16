// Package cache persists runtime data (user IDs and cached scores) to a
// JSON file separate from config.json.
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ScoreEntry is a single cached score. The score payload is stored as raw
// JSON so this package does not depend on the osu! API types.
type ScoreEntry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type fileCache struct {
	UserIDs map[string]int        `json:"user_ids"`
	Scores  map[string]ScoreEntry `json:"scores"`
}

// Cache is a mutex-guarded, file-backed store.
type Cache struct {
	mu      sync.Mutex
	path    string
	userIDs map[string]int
	scores  map[string]ScoreEntry
	dirty   bool
}

// New returns an empty cache that persists to path.
func New(path string) *Cache {
	return &Cache{
		path:    path,
		userIDs: make(map[string]int),
		scores:  make(map[string]ScoreEntry),
	}
}

// Load reads the cache file at path. A missing file yields an empty cache;
// a malformed file is an error.
func Load(path string) (*Cache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return New(path), nil
		}
		return nil, fmt.Errorf("read cache: %w", err)
	}

	var disk fileCache
	if err := json.Unmarshal(data, &disk); err != nil {
		return nil, fmt.Errorf("parse cache: %w", err)
	}

	c := New(path)
	for username, id := range disk.UserIDs {
		c.userIDs[username] = id
	}
	for key, entry := range disk.Scores {
		if time.Now().Before(entry.ExpiresAt) {
			c.scores[key] = entry
		}
	}

	return c, nil
}

func (c *Cache) UserID(username string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id, ok := c.userIDs[username]
	return id, ok
}

func (c *Cache) SetUserID(username string, id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.userIDs[username] = id
	c.dirty = true
}

func (c *Cache) UserIDCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.userIDs)
}

// Score returns the raw score stored under key, honoring its expiry.
func (c *Cache) Score(key string) (json.RawMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.scores[key]
	if !ok {
		return nil, false
	}

	if !time.Now().Before(entry.ExpiresAt) {
		delete(c.scores, key)
		c.dirty = true
		return nil, false
	}

	return entry.Data, true
}

func (c *Cache) SetScore(key string, data json.RawMessage, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.scores[key] = ScoreEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.dirty = true
}

// Save atomically writes the cache to its path if anything changed since
// the last save.
func (c *Cache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil
	}

	data, err := json.MarshalIndent(fileCache{
		UserIDs: c.userIDs,
		Scores:  c.scores,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, ".cache-*")
	if err != nil {
		return fmt.Errorf("create temp cache: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("set cache permissions: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write cache: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync cache: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close cache: %w", err)
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		return fmt.Errorf("replace cache: %w", err)
	}

	c.dirty = false
	return nil
}
