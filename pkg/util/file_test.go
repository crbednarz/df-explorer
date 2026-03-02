package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheDirIsDirectory(t *testing.T) {
	dir, err := CacheDir()
	if err != nil {
		assert.NoError(t, err)
	}

	stat, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.True(t, stat.IsDir())
}
