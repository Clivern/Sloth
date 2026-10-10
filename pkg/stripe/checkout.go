// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package stripe

import (
	"context"
	"fmt"
	"strconv"

	"github.com/samber/lo"
	stripesdk "github.com/stripe/stripe-go/v87"
)

// CheckoutOptions is input for creating a token purchase Checkout Session.
type CheckoutOptions struct {
	WorkspaceId   string
	AmountCents   int64
	Tokens        int64
	CustomerId    string
	CustomerEmail string
	SuccessURL    string
	CancelURL     string
}

// CreateCheckoutSession starts a Stripe Checkout payment for AI tokens.
func (c *Client) CreateCheckoutSession(ctx context.Context, opts CheckoutOptions) (*stripesdk.CheckoutSession, error) {
	if opts.AmountCents < MinTokenPurchaseCents || opts.Tokens <= 0 {
		return nil, ErrInvalidAmount
	}

	params := &stripesdk.CheckoutSessionCreateParams{
		Mode:       stripesdk.String(string(stripesdk.CheckoutSessionModePayment)),
		SuccessURL: stripesdk.String(opts.SuccessURL),
		CancelURL:  stripesdk.String(opts.CancelURL),
		LineItems: []*stripesdk.CheckoutSessionCreateLineItemParams{
			{
				Quantity: stripesdk.Int64(1),
				PriceData: &stripesdk.CheckoutSessionCreateLineItemPriceDataParams{
					Currency:   stripesdk.String("usd"),
					UnitAmount: stripesdk.Int64(opts.AmountCents),
					ProductData: &stripesdk.CheckoutSessionCreateLineItemPriceDataProductDataParams{
						Name: stripesdk.String("AI tokens"),
					},
				},
			},
		},
		Metadata: map[string]string{
			"workspaceId": opts.WorkspaceId,
			"tokens":      strconv.FormatInt(opts.Tokens, 10),
			"amountCents": strconv.FormatInt(opts.AmountCents, 10),
		},
		PaymentIntentData: &stripesdk.CheckoutSessionCreatePaymentIntentDataParams{
			Metadata: map[string]string{
				"workspaceId": opts.WorkspaceId,
				"tokens":      strconv.FormatInt(opts.Tokens, 10),
				"amountCents": strconv.FormatInt(opts.AmountCents, 10),
			},
		},
	}

	if lo.IsNotEmpty(opts.CustomerId) {
		params.Customer = stripesdk.String(opts.CustomerId)
	} else if lo.IsNotEmpty(opts.CustomerEmail) {
		params.CustomerEmail = stripesdk.String(opts.CustomerEmail)
		params.CustomerCreation = stripesdk.String(string(stripesdk.CheckoutSessionCustomerCreationAlways))
	}

	session, err := c.api.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create checkout session: %w", err)
	}

	return session, nil
}
