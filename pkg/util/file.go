package util

import (
	"fmt"
	"os"
	"path"
)

// CachDir returns a shared cache directory for this utility.
// Example: ~/.cache/df-explorer/
// If the directory does not exist, it will be created.
func CacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("unable to get user cache directory: %w", err)
	}

	cacheDir := path.Join(userCacheDir, "df-explorer")
	err = os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return "", fmt.Errorf("unable to create cache directory: %w", err)
	}

	return cacheDir, nil
}
