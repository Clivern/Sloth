// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitBus(t *testing.T) {
	t.Run("get bus returns nil before the bus starts", func(t *testing.T) {
		assert.Nil(t, GetBus())
	})
}

func TestUnitBusConnect(t *testing.T) {
	t.Run("start bus returns the connect error for an unreachable server", func(t *testing.T) {
		viper.Set("app.broker.nats.url", "nats://127.0.0.1:1")
		t.Cleanup(func() { viper.Set("app.broker.nats.url", "") })

		err := StartBus()
		assert.ErrorContains(t, err, "connect nats bus")
		assert.Nil(t, GetBus())
	})
}
