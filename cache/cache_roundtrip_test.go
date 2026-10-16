package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")

	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.UserID("yuki"); ok {
		t.Fatal("expected empty cache")
	}

	c.SetUserID("yuki", 42)
	c.SetScore("1:42", json.RawMessage(`{"id":7}`), time.Minute)
	if err := c.Save(); err != nil {
		t.Fatal(err)
	}

	c2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := c2.UserID("yuki"); !ok || id != 42 {
		t.Fatalf("user id = %d, %v", id, ok)
	}
	if raw, ok := c2.Score("1:42"); !ok {
		t.Fatalf("score missing: %v", ok)
	} else {
		var got struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw, &got); err != nil || got.ID != 7 {
			t.Fatalf("score = %s, %v", raw, err)
		}
	}

	c2.SetScore("2:42", json.RawMessage(`{"id":8}`), -time.Second)
	if _, ok := c2.Score("2:42"); ok {
		t.Fatal("expected expired score to be gone")
	}
	c2.SetUserID("expired", 1)
	c2.SetScore("3:1", json.RawMessage(`{}`), -time.Minute)
	if err := c2.Save(); err != nil {
		t.Fatal(err)
	}

	c3, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c3.Score("3:1"); ok {
		t.Fatal("expected expired score pruned on load")
	}
	if _, ok := c3.UserID("expired"); !ok {
		t.Fatal("expected user id present")
	}

	if fi, err := os.Stat(path); err != nil {
		t.Fatal(err)
	} else if fi.Mode().Perm() != 0600 {
		t.Fatalf("perm = %v", fi.Mode().Perm())
	}
}
