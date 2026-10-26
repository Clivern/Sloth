// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package ai

import (
	"github.com/clivern/sloth/pkg/ai/muse"
	"github.com/clivern/sloth/pkg/ai/openrouter"
	"github.com/spf13/viper"
)

const (
	ProviderOpenRouter = "openrouter"
	ProviderMuse       = "muse"
)

// IsConfigured reports whether an AI provider API key is set.
func IsConfigured() bool {
	return viper.GetString("app.ai.api_key") != ""
}

// NewChatClient returns the chat client for the app.ai.provider config.
func NewChatClient() ChatClient {
	apiKey := viper.GetString("app.ai.api_key")
	model := viper.GetString("app.ai.chat_model")

	if viper.GetString("app.ai.provider") == ProviderMuse {
		return NewMuseChat(muse.New(muse.Config{
			APIKey: apiKey,
			Model:  model,
			Effort: viper.GetString("app.ai.reasoning_effort"),
		}))
	}

	return NewOpenRouterChat(openrouter.New(openrouter.Config{APIKey: apiKey, Model: model}))
}
