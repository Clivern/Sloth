// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitGravatarURL(t *testing.T) {
	t.Run("Complete gravatar workflow", func(t *testing.T) {
		url := GravatarURL("user@example.com", 0)
		assert.True(t, strings.HasPrefix(url, "https://www.gravatar.com/avatar/"))
		assert.Contains(t, url, "?d=identicon")
		assert.NotContains(t, url, "s=")

		urlSized := GravatarURL("u@e.com", 128)
		assert.Contains(t, urlSized, "s=128")

		assert.Equal(t, GravatarURL("test@example.com", 80), GravatarURL("test@example.com", 80))
		assert.Equal(t, GravatarURL("User@Example.COM", 0), GravatarURL("user@example.com", 0))

		assert.Empty(t, GravatarURL("", 0))
		assert.Empty(t, GravatarURL("   ", 0))
		assert.NotContains(t, GravatarURL("u@e.com", 3000), "s=")
	})
}
