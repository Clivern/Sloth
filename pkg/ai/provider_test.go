// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package ai

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitProvider(t *testing.T) {
	t.Run("openrouter", func(t *testing.T) {
		viper.Set("app.ai.provider", ProviderOpenRouter)

		assert.IsType(t, &OpenRouterChat{}, NewChatClient())
	})

	t.Run("muse", func(t *testing.T) {
		viper.Set("app.ai.provider", ProviderMuse)

		assert.IsType(t, &MuseChat{}, NewChatClient())
	})
}
