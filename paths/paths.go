// Package paths resolves the OS-specific directories where ohana stores its
// configuration and cache files.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "ohana"

// ConfigFile returns the OS-specific path to the named configuration file,
// creating its parent directory if needed.
//
// On Linux this is $XDG_CONFIG_HOME/ohana (default ~/.config/ohana); on
// Windows it is %AppData%\ohana.
func ConfigFile(name string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate config directory: %w", err)
	}

	return ensure(filepath.Join(base, appName, name))
}

// CacheFile returns the OS-specific path to the named cache file, creating
// its parent directory if needed.
//
// On Linux this is $XDG_CACHE_HOME/ohana (default ~/.cache/ohana); on
// Windows it is %LocalAppData%\ohana.
func CacheFile(name string) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locate cache directory: %w", err)
	}

	return ensure(filepath.Join(base, appName, name))
}

func ensure(file string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}

	return file, nil
}
