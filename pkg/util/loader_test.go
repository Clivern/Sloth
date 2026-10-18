// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitLoader(t *testing.T) {
	t.Run("GetBaseDir", func(t *testing.T) {
		root := t.TempDir()
		marker := filepath.Join(root, "cache")
		assert.NoError(t, os.Mkdir(marker, 0755))

		subDir := filepath.Join(root, "nested", "deep")
		assert.NoError(t, os.MkdirAll(subDir, 0755))

		origWD, err := os.Getwd()
		assert.NoError(t, err)
		t.Cleanup(func() { _ = os.Chdir(origWD) })

		assert.NoError(t, os.Chdir(subDir))
		base := GetBaseDir("cache")
		assert.True(t, DirExists(filepath.Join(base, "cache")))
	})

	t.Run("LoadConfigs", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		configPath := filepath.Join(t.TempDir(), "config.yaml")
		content := []byte("app:\n  name: sloth\n  port: 8080\n")
		assert.NoError(t, os.WriteFile(configPath, content, 0644))

		assert.NoError(t, LoadConfigs(configPath))
		assert.Equal(t, "sloth", viper.GetString("app.name"))
		assert.Equal(t, 8080, viper.GetInt("app.port"))

		assert.Error(t, LoadConfigs(filepath.Join(t.TempDir(), "missing.yaml")))
	})
}
