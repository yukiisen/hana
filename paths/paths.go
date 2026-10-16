// Package paths resolves the OS-specific directories where hana stores its
// configuration and cache files.
package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "hana"

// ConfigFile returns the OS-specific path to the named configuration file,
// creating its parent directory if needed.
//
// On Linux this is $XDG_CONFIG_HOME/hana (default ~/.config/hana); on
// Windows it is %AppData%\hana.
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
// On Linux this is $XDG_CACHE_HOME/hana (default ~/.cache/hana); on
// Windows it is %LocalAppData%\hana.
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
