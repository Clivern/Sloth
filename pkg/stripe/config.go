// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import "github.com/spf13/viper"

const (
	DefaultTokensPerUSD   = 20_000
	MinTokenPurchaseCents = 100
)

// URLs holds Stripe redirect URLs (relative or absolute).
type URLs struct {
	CheckoutSuccess string
	CheckoutCancel  string
	PortalReturn    string
}

// Config holds Stripe client configuration.
type Config struct {
	SecretKey     string
	WebhookSecret string
	TokensPerUSD  int64
	URLs          URLs
}

// TokensForCents converts a USD cent amount into AI tokens.
func (c Config) TokensForCents(amountCents int64) int64 {
	rate := c.TokensPerUSD
	if rate <= 0 {
		rate = DefaultTokensPerUSD
	}

	return amountCents * rate / 100
}

// TokensPerUSD is the AI token grant for each US dollar.
func TokensPerUSD() int64 {
	rate := viper.GetInt64("app.billing.tokens_per_usd")
	if rate > 0 {
		return rate
	}

	return DefaultTokensPerUSD
}
