// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/pkg/stripe"
	"github.com/clivern/cognit/pkg/util"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	stripesdk "github.com/stripe/stripe-go/v87"
)

var (
	ErrBillingPortalUnavailable    = errors.New("stripe billing portal is unavailable for this workspace")
	ErrBillingSubscriptionNotFound = errors.New("workspace subscription not found")
)

// Billing coordinates prepaid AI tokens with Stripe Checkout and Portal.
type Billing struct {
	WorkspaceRepository    db.WorkspaceRepository
	SubscriptionRepository db.SubscriptionRepository
	PurchaseRepository     db.TokenPurchaseRepository
	Usage                  Usage
}

// BillingStatusResponse is the billing state exposed to the UI.
type BillingStatusResponse struct {
	Id                 db.Id  `json:"id"`
	WorkspaceId        db.Id  `json:"workspaceId"`
	ProviderCustomerId string `json:"providerCustomerId"`
	AITokensBalance    int64  `json:"aiTokensBalance"`
	TokensPerUsd       int64  `json:"tokensPerUsd"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// BillingCheckoutRequest is the body for creating a Stripe Checkout session.
type BillingCheckoutRequest struct {
	AmountCents int64 `json:"amountCents" validate:"required,gte=100" label:"Amount"`
}

// BillingSessionResponse contains a Stripe redirect URL.
type BillingSessionResponse struct {
	URL string `json:"url"`
}

// BillingWebhookResponse is returned after a Stripe webhook is handled.
type BillingWebhookResponse struct {
	Received bool `json:"received"`
}

// WorkspaceUsageLimits are the consumption caps shown on the billing page.
type WorkspaceUsageLimits struct {
	WorkspaceMembers int64 `json:"workspaceMembers"`
	AITokens         int64 `json:"aiTokens"`
}

// BillingUsageResponse is workspace usage and limits for the billing page.
type BillingUsageResponse struct {
	PeriodReset string                `json:"periodReset"`
	Used        WorkspaceUsageMetrics `json:"used"`
	Limits      WorkspaceUsageLimits  `json:"limits"`
}

var workspaceUsageLimits = WorkspaceUsageLimits{
	WorkspaceMembers: 3,
}

// NewBilling creates a billing module.
func NewBilling(workspaces db.WorkspaceRepository, subscriptions db.SubscriptionRepository, purchases db.TokenPurchaseRepository, usage Usage) *Billing {
	return &Billing{
		WorkspaceRepository:    workspaces,
		SubscriptionRepository: subscriptions,
		PurchaseRepository:     purchases,
		Usage:                  usage,
	}
}

// GetBillingStatus returns billing status for a workspace.
func (b *Billing) GetBillingStatus(workspaceId db.Id) (*BillingStatusResponse, error) {
	workspace, err := b.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrBillingSubscriptionNotFound
	}

	return &BillingStatusResponse{
		Id:                 subscription.Id,
		WorkspaceId:        subscription.WorkspaceId,
		ProviderCustomerId: lo.FromPtr(subscription.ProviderCustomerId),
		AITokensBalance:    subscription.AITokensBalance,
		TokensPerUsd:       stripe.TokensPerUSD(),
		CreatedAt:          subscription.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:          subscription.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// GetBillingUsage returns usage and limits for a workspace.
func (b *Billing) GetBillingUsage(workspaceId db.Id, deps UsageSnapshotDeps) (*BillingUsageResponse, error) {
	workspace, err := b.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrBillingSubscriptionNotFound
	}

	used, err := b.Usage.GetWorkspaceUsage(deps, workspaceId)
	if err != nil {
		return nil, err
	}

	limits := workspaceUsageLimits
	limits.AITokens = subscription.AITokensBalance + used.AITokens

	_, periodEnd := util.CurrentMonthPeriod()

	return &BillingUsageResponse{
		PeriodReset: periodEnd.UTC().Format(time.RFC3339),
		Used:        *used,
		Limits:      limits,
	}, nil
}

// CreateCheckoutSession starts Stripe Checkout for an AI token purchase.
func (b *Billing) CreateCheckoutSession(ctx context.Context, workspaceId db.Id, user *db.User, req *BillingCheckoutRequest, successURL string, cancelURL string) (*BillingSessionResponse, error) {
	workspace, err := b.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	client, err := stripe.New()
	if err != nil {
		return nil, err
	}

	tokens := client.Config().TokensForCents(req.AmountCents)
	if req.AmountCents < stripe.MinTokenPurchaseCents || tokens <= 0 {
		return nil, stripe.ErrInvalidAmount
	}

	subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrBillingSubscriptionNotFound
	}

	session, err := client.CreateCheckoutSession(ctx, stripe.CheckoutOptions{
		WorkspaceId:   workspaceId.String(),
		AmountCents:   req.AmountCents,
		Tokens:        tokens,
		CustomerId:    lo.FromPtr(subscription.ProviderCustomerId),
		CustomerEmail: user.Email,
		SuccessURL:    successURL,
		CancelURL:     cancelURL,
	})
	if err != nil {
		return nil, err
	}

	return &BillingSessionResponse{URL: session.URL}, nil
}

// CreatePortalSession opens Stripe's customer portal.
func (b *Billing) CreatePortalSession(ctx context.Context, workspaceId db.Id, returnURL string) (*BillingSessionResponse, error) {
	workspace, err := b.WorkspaceRepository.GetById(workspaceId)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, ErrWorkspaceNotFound
	}

	client, err := stripe.New()
	if err != nil {
		return nil, err
	}

	subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrBillingSubscriptionNotFound
	}

	customerId := lo.FromPtr(subscription.ProviderCustomerId)
	if lo.IsEmpty(customerId) {
		return nil, ErrBillingPortalUnavailable
	}

	session, err := client.CreatePortalSession(ctx, stripe.PortalOptions{
		CustomerId: customerId,
		ReturnURL:  returnURL,
	})
	if err != nil {
		return nil, err
	}

	return &BillingSessionResponse{URL: session.URL}, nil
}

// HandleWebhook updates local billing state from Stripe webhook events.
func (b *Billing) HandleWebhook(payload []byte, signature string) error {
	client, err := stripe.New()
	if err != nil {
		return err
	}

	event, err := client.ConstructWebhookEvent(payload, signature)
	if err != nil {
		return err
	}

	if event.Type != "checkout.session.completed" {
		return nil
	}

	var session stripesdk.CheckoutSession
	err = json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		return fmt.Errorf("decode checkout session: %w", err)
	}

	return b.CreditTokenPurchase(client, &session)
}

// CreditTokenPurchase credits AI tokens after a successful Stripe payment.
func (b *Billing) CreditTokenPurchase(client *stripe.Client, session *stripesdk.CheckoutSession) error {
	workspaceId := db.Id(session.Metadata["workspaceId"])
	if lo.IsEmpty(workspaceId) {
		return nil
	}

	amountCents := session.AmountTotal
	tokens := client.Config().TokensForCents(amountCents)
	if tokens <= 0 {
		return nil
	}

	inserted, err := b.PurchaseRepository.Create(&db.TokenPurchase{
		WorkspaceId:     workspaceId,
		StripeSessionId: session.ID,
		AmountCents:     amountCents,
		Tokens:          tokens,
	})
	if err != nil {
		return fmt.Errorf("record token purchase: %w", err)
	}
	if !inserted {
		return nil
	}

	err = b.SubscriptionRepository.AddTokens(workspaceId, tokens)
	if err != nil {
		return fmt.Errorf("credit ai tokens: %w", err)
	}

	if session.Customer != nil && lo.IsNotEmpty(session.Customer.ID) {
		subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
		if err != nil {
			return err
		}
		if subscription != nil {
			subscription.ProviderCustomerId = lo.ToPtr(session.Customer.ID)
			err = b.SubscriptionRepository.Update(subscription)
			if err != nil {
				return err
			}
		}
	}

	log.Info().
		Str("workspaceId", workspaceId.String()).
		Str("sessionId", session.ID).
		Int64("amountCents", amountCents).
		Int64("tokens", tokens).
		Msg("AI tokens credited")

	return nil
}

// GetWorkspaceSubscription gets a workspace subscription from the database.
func (b *Billing) GetWorkspaceSubscription(workspaceId db.Id) (*db.Subscription, error) {
	subscription, err := b.SubscriptionRepository.GetByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrBillingSubscriptionNotFound
	}

	return subscription, nil
}
