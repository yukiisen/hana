package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("AppData", dir)

	got, err := ConfigFile("config.json")
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "hana", "config.json")
	if got != want {
		t.Fatalf("ConfigFile = %q, want %q", got, want)
	}

	assertDir(t, filepath.Dir(got))
}

func TestCacheFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LocalAppData", dir)

	got, err := CacheFile("cache.json")
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(dir, "hana", "cache.json")
	if got != want {
		t.Fatalf("CacheFile = %q, want %q", got, want)
	}

	assertDir(t, filepath.Dir(got))
}

func assertDir(t *testing.T, dir string) {
	t.Helper()

	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat %q: %v", dir, err)
	}
	if !fi.IsDir() {
		t.Fatalf("%q is not a directory", dir)
	}
}
