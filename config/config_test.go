package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaults(t *testing.T) {
	path := writeConfig(t, `{"api":{"client_id":1,"client_secret":"secret"}}`)

	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := c.LeaderboardWaitTime(); got != 2*time.Second {
		t.Fatalf("LeaderboardWaitTime = %v, want 2s", got)
	}
	if got := c.ScoreCacheTTL(); got != 5*time.Minute {
		t.Fatalf("ScoreCacheTTL = %v, want 5m", got)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	path := writeConfig(t, `{
		"api": {"client_id": 1, "client_secret": "secret"},
		"settings": {
			"leaderboard_wait_time": "500ms",
			"score_cache_ttl": "90s"
		}
	}`)

	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if got := c.LeaderboardWaitTime(); got != 500*time.Millisecond {
		t.Fatalf("LeaderboardWaitTime = %v, want 500ms", got)
	}
	if got := c.ScoreCacheTTL(); got != 90*time.Second {
		t.Fatalf("ScoreCacheTTL = %v, want 90s", got)
	}

	c.SetLeaderboards(map[string][]string{"friends": {"yuki"}})
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	c2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := c2.LeaderboardWaitTime(); got != 500*time.Millisecond {
		t.Fatalf("reloaded LeaderboardWaitTime = %v, want 500ms", got)
	}
	if got := c2.ScoreCacheTTL(); got != 90*time.Second {
		t.Fatalf("reloaded ScoreCacheTTL = %v, want 90s", got)
	}
}

func TestDurationRejectsInvalid(t *testing.T) {
	path := writeConfig(t, `{
		"api": {"client_id": 1, "client_secret": "secret"},
		"settings": {"score_cache_ttl": "banana"}
	}`)

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}
