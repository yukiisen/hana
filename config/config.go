package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultLeaderboardWaitTime = 2 * time.Second
	defaultScoreCacheTTL       = 5 * time.Minute
)

type Config struct {
	API      APICredentials `json:"api"`
	Settings Settings       `json:"settings"`

	mu           sync.RWMutex
	path         string
	leaderboards map[string][]string
}

type APICredentials struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// Settings holds tunable runtime options.
type Settings struct {
	// LeaderboardWaitTime is how long to wait for freshly fetched scores
	// before starting to poll the API when a leaderboard has cache misses.
	LeaderboardWaitTime Duration `json:"leaderboard_wait_time"`
	// ScoreCacheTTL is how long a fetched score is served from the cache
	// before it is fetched again.
	ScoreCacheTTL Duration `json:"score_cache_ttl"`
}

// Duration is a time.Duration that marshals to and from a human-readable
// string such as "2s" or "5m".
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf(`duration must be a string like "2s": %w`, err)
	}
	if s == "" {
		return nil
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(parsed)
	return nil
}

// Std returns the value as a time.Duration.
func (d Duration) Std() time.Duration { return time.Duration(d) }

type fileConfig struct {
	API          APICredentials      `json:"api"`
	Settings     Settings            `json:"settings"`
	Leaderboards map[string][]string `json:"leaderboards"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var disk fileConfig
	if err := json.Unmarshal(data, &disk); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if disk.API.ClientID <= 0 {
		return nil, fmt.Errorf("invalid API client_id")
	}
	if disk.API.ClientSecret == "" {
		return nil, fmt.Errorf("missing API client_secret")
	}
	if disk.Leaderboards == nil {
		disk.Leaderboards = make(map[string][]string)
	}
	if disk.Settings.LeaderboardWaitTime <= 0 {
		disk.Settings.LeaderboardWaitTime = Duration(defaultLeaderboardWaitTime)
	}
	if disk.Settings.ScoreCacheTTL <= 0 {
		disk.Settings.ScoreCacheTTL = Duration(defaultScoreCacheTTL)
	}

	return &Config{
		API:          disk.API,
		Settings:     disk.Settings,
		path:         path,
		leaderboards: disk.Leaderboards,
	}, nil
}

// LeaderboardWaitTime returns the configured leaderboard wait time, falling
// back to the default when unset.
func (c *Config) LeaderboardWaitTime() time.Duration {
	if c.Settings.LeaderboardWaitTime <= 0 {
		return defaultLeaderboardWaitTime
	}
	return c.Settings.LeaderboardWaitTime.Std()
}

// ScoreCacheTTL returns the configured score cache lifetime, falling back to
// the default when unset.
func (c *Config) ScoreCacheTTL() time.Duration {
	if c.Settings.ScoreCacheTTL <= 0 {
		return defaultScoreCacheTTL
	}
	return c.Settings.ScoreCacheTTL.Std()
}

func (c *Config) Leaderboards() map[string][]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return cloneLeaderboards(c.leaderboards)
}

func (c *Config) SetLeaderboards(boards map[string][]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.leaderboards = cloneLeaderboards(boards)
}

func cloneLeaderboards(src map[string][]string) map[string][]string {
	dst := make(map[string][]string, len(src))
	for name, players := range src {
		dst[name] = append([]string(nil), players...)
	}
	return dst
}

// Atomically writes the current configuration to its original path.
func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(fileConfig{
		API:          c.API,
		Settings:     c.Settings,
		Leaderboards: c.leaderboards,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, ".config-*")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return fmt.Errorf("set config permissions: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close config: %w", err)
	}

	if err := os.Rename(tmpPath, c.path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}

	return nil
}
