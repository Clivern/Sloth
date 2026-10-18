// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"github.com/clivern/sloth/conf"
	"github.com/spf13/viper"
	stripesdk "github.com/stripe/stripe-go/v87"
)

// Client wraps the Stripe SDK using sloth billing configuration.
type Client struct {
	api    *stripesdk.Client
	config Config
}

// New returns a Stripe client loaded from app.billing config.
func New() (*Client, error) {
	if !conf.IsSaaS() || viper.GetString("app.billing.provider") != "stripe" {
		return nil, ErrBillingDisabled
	}

	return &Client{
		api: stripesdk.NewClient(viper.GetString("app.billing.secret_key")),
		config: Config{
			SecretKey:     viper.GetString("app.billing.secret_key"),
			WebhookSecret: viper.GetString("app.billing.webhook_secret"),
			TokensPerUSD:  viper.GetInt64("app.billing.tokens_per_usd"),
			URLs: URLs{
				CheckoutSuccess: viper.GetString("app.billing.urls.checkout_success"),
				CheckoutCancel:  viper.GetString("app.billing.urls.checkout_cancel"),
				PortalReturn:    viper.GetString("app.billing.urls.portal_return"),
			},
		},
	}, nil
}

// Config returns the billing configuration used by this client.
func (c *Client) Config() Config {
	return c.config
}
