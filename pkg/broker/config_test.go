// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package broker

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitGetConfig(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("app.broker.provider", "nats")
	viper.Set("app.broker.nats.url", "nats://localhost:4222")
	viper.Set("app.broker.nats.name", "sloth")
	viper.Set("app.broker.nats.queue", "workers")

	cfg := GetConfig()
	assert.Equal(t, "nats", cfg.Provider)
	assert.Equal(t, "nats://localhost:4222", cfg.NATS.URL)
	assert.Equal(t, "sloth", cfg.NATS.Name)
	assert.Equal(t, "workers", cfg.NATS.Queue)
}
