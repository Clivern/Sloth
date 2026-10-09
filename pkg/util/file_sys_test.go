// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testBaseDir(t *testing.T) string {
	dir := t.TempDir()
	_ = os.WriteFile(fmt.Sprintf("%s/.gitignore", dir), nil, 0644)

	return dir
}

func TestUnitFileSystem(t *testing.T) {
	base := testBaseDir(t)

	t.Run("FileExists", func(t *testing.T) {
		assert.True(t, FileExists(fmt.Sprintf("%s/.gitignore", base)))
		assert.False(t, FileExists(fmt.Sprintf("%s/not_found.md", base)))
		assert.False(t, FileExists(base))
	})

	t.Run("DirExists", func(t *testing.T) {
		assert.True(t, DirExists(base))
		assert.False(t, DirExists(fmt.Sprintf("%s/not_found", base)))
		assert.False(t, DirExists(fmt.Sprintf("%s/.gitignore", base)))
	})

	t.Run("EnsureDir and DeleteDir", func(t *testing.T) {
		newDir := fmt.Sprintf("%s/test_new_dir", base)
		_ = DeleteDir(newDir)
		assert.NoError(t, EnsureDir(newDir, 0755))
		assert.True(t, DirExists(newDir))
		assert.NoError(t, EnsureDir(base, 0755))
		assert.NoError(t, DeleteDir(newDir))
		assert.False(t, DirExists(newDir))
		assert.NoError(t, DeleteDir(fmt.Sprintf("%s/non_existing_dir", base)))
	})
}
