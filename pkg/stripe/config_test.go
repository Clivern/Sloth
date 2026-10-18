// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitTokens(t *testing.T) {
	t.Run("TokensForCents", func(t *testing.T) {
		cfg := Config{TokensPerUSD: 20_000}
		assert.Equal(t, int64(20_000), cfg.TokensForCents(100))
		assert.Equal(t, int64(10_000), cfg.TokensForCents(50))

		fallback := Config{}
		assert.Equal(t, int64(DefaultTokensPerUSD), fallback.TokensForCents(100))
	})

	t.Run("TokensPerUSD", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		assert.Equal(t, int64(DefaultTokensPerUSD), TokensPerUSD())

		viper.Set("app.billing.tokens_per_usd", int64(50_000))
		assert.Equal(t, int64(50_000), TokensPerUSD())
	})
}
