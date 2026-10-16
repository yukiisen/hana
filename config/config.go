
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	API APICredentials `json:"api"`

	mu           sync.RWMutex
	path         string
	leaderboards map[string][]string
}

type APICredentials struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type fileConfig struct {
	API          APICredentials     `json:"api"`
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

	return &Config{
		API:          disk.API,
		path:         path,
		leaderboards: disk.Leaderboards,
	}, nil
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
