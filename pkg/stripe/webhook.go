// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"fmt"

	stripesdk "github.com/stripe/stripe-go/v87"
	"github.com/stripe/stripe-go/v87/webhook"
)

// ConstructWebhookEvent validates and parses a Stripe webhook payload.
func (c *Client) ConstructWebhookEvent(payload []byte, signature string) (stripesdk.Event, error) {
	event, err := webhook.ConstructEventWithOptions(
		payload,
		signature,
		c.config.WebhookSecret,
		webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		return stripesdk.Event{}, fmt.Errorf("construct webhook event: %w", err)
	}

	return event, nil
}
