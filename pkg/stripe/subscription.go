// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"context"
	"fmt"

	stripesdk "github.com/stripe/stripe-go/v87"
)

// GetSubscription loads a Stripe subscription by Id.
func (c *Client) GetSubscription(ctx context.Context, subscriptionId string) (*stripesdk.Subscription, error) {
	subscription, err := c.api.V1Subscriptions.Retrieve(
		ctx,
		subscriptionId,
		&stripesdk.SubscriptionRetrieveParams{},
	)
	if err != nil {
		return nil, fmt.Errorf("retrieve subscription: %w", err)
	}

	return subscription, nil
}
