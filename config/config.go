package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	API          APICredentials   `json:"api"`
	Leaderboards map[string][]string `json:"leaderboards"`
}

type APICredentials struct {
	ClientID     int    `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.API.ClientID <= 0 {
		return nil, fmt.Errorf("invalid API client_id")
	}
	if cfg.API.ClientSecret == "" {
		return nil, fmt.Errorf("missing API client_secret")
	}
	if cfg.Leaderboards == nil {
		cfg.Leaderboards = make(map[string][]string)
	}

	return &cfg, nil
}
