// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"context"
	"fmt"

	stripesdk "github.com/stripe/stripe-go/v87"
)

// PortalOptions is input for creating a billing portal session.
type PortalOptions struct {
	CustomerId string
	ReturnURL  string
}

// CreatePortalSession opens the Stripe customer billing portal.
func (c *Client) CreatePortalSession(ctx context.Context, opts PortalOptions) (*stripesdk.BillingPortalSession, error) {
	session, err := c.api.V1BillingPortalSessions.Create(ctx, &stripesdk.BillingPortalSessionCreateParams{
		Customer:  stripesdk.String(opts.CustomerId),
		ReturnURL: stripesdk.String(opts.ReturnURL),
	})
	if err != nil {
		return nil, fmt.Errorf("create portal session: %w", err)
	}

	return session, nil
}
