// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/locale"
	"github.com/clivern/cognit/module"
	"github.com/clivern/cognit/pkg/stripe"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

// GetBillingStatusAction returns the current workspace billing state.
func (a *API) GetBillingStatusAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	log.Info().
		Str("workspaceId", workspaceId).
		Msg("Getting billing status")

	status, err := a.Billing.GetBillingStatus(db.Id(workspaceId))
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrBillingSubscriptionNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_subscription_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to get billing status")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_billing_status"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, status)
}

// GetBillingUsageAction returns workspace usage and plan limits for billing.
func (a *API) GetBillingUsageAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	log.Info().
		Str("workspaceId", workspaceId).
		Msg("Getting billing usage")

	usage, err := a.Billing.GetBillingUsage(db.Id(workspaceId), module.UsageSnapshotDeps{
		WorkspaceUserRepository: a.WorkspaceUsers,
		UsageRepository:         a.Usage,
	})
	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrBillingSubscriptionNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_subscription_not_found"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to get billing usage")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_get_billing_usage"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, usage)
}

// CreateBillingCheckoutAction creates a Stripe Checkout session for AI tokens.
func (a *API) CreateBillingCheckoutAction(w http.ResponseWriter, r *http.Request) {
	var req module.BillingCheckoutRequest
	err := a.DecodeAndValidate(r, &req)
	if err != nil {
		a.WriteValidationError(w, err)
		return
	}

	user, ok := a.GetUser(r)
	if !ok || user == nil {
		a.WriteJSON(w, http.StatusUnauthorized, map[string]any{
			"errorMessage": locale.TR(r, "not_authenticated"),
		})
		return
	}

	workspaceId := chi.URLParam(r, "workspaceId")
	log.Info().
		Str("workspaceId", workspaceId).
		Str("userId", user.Id.String()).
		Int64("amountCents", req.AmountCents).
		Msg("New billing checkout request")

	session, err := a.Billing.CreateCheckoutSession(
		r.Context(),
		db.Id(workspaceId),
		user,
		&req,
		fmt.Sprintf("%s/billing?checkout=success", viper.GetString("app.url")),
		fmt.Sprintf("%s/billing?checkout=cancel", viper.GetString("app.url")),
	)

	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrBillingSubscriptionNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_subscription_not_found"),
			})
		case errors.Is(err, stripe.ErrInvalidAmount):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_billing_amount"),
			})
		case errors.Is(err, stripe.ErrBillingDisabled):
			a.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
				"errorMessage": locale.TR(r, "stripe_billing_not_configured"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to create billing checkout session")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_create_billing_checkout_session"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, session)
}

// CreateBillingPortalAction creates a Stripe Billing Portal session.
func (a *API) CreateBillingPortalAction(w http.ResponseWriter, r *http.Request) {
	workspaceId := chi.URLParam(r, "workspaceId")
	log.Info().
		Str("workspaceId", workspaceId).
		Msg("New billing portal request")

	session, err := a.Billing.CreatePortalSession(
		r.Context(),
		db.Id(workspaceId),
		fmt.Sprintf("%s/billing", viper.GetString("app.url")),
	)

	if err != nil {
		switch {
		case errors.Is(err, module.ErrWorkspaceNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_not_found"),
			})
		case errors.Is(err, module.ErrBillingPortalUnavailable):
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "stripe_billing_available_after_purchase"),
			})
		case errors.Is(err, module.ErrBillingSubscriptionNotFound):
			a.WriteJSON(w, http.StatusNotFound, map[string]any{
				"errorMessage": locale.TR(r, "workspace_subscription_not_found"),
			})
		case errors.Is(err, stripe.ErrBillingDisabled):
			a.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
				"errorMessage": locale.TR(r, "stripe_billing_not_configured"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to create billing portal session")
			a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"errorMessage": locale.TR(r, "failed_create_billing_portal_session"),
			})
		}
		return
	}

	a.WriteJSON(w, http.StatusOK, session)
}

// StripeWebhookAction receives Stripe billing webhook events.
func (a *API) StripeWebhookAction(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Stripe webhook rejected: failed to read payload")
		a.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"errorMessage": locale.TR(r, "invalid_webhook_payload"),
		})
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	log.Info().
		Int("payloadBytes", len(payload)).
		Bool("hasSignature", lo.IsNotEmpty(signature)).
		Msg("Stripe webhook received")

	err = a.Billing.HandleWebhook(payload, signature)
	if err != nil {
		switch {
		case errors.Is(err, stripe.ErrBillingDisabled), errors.Is(err, stripe.ErrWebhookNotConfigured):
			log.Warn().
				Err(err).
				Msg("Stripe webhook rejected: billing not configured")
			a.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
				"errorMessage": locale.TR(r, "stripe_billing_not_configured"),
			})
		default:
			log.Error().
				Err(err).
				Msg("Failed to handle Stripe webhook")
			a.WriteJSON(w, http.StatusBadRequest, map[string]any{
				"errorMessage": locale.TR(r, "invalid_stripe_webhook"),
			})
		}
		return
	}

	log.Info().
		Msg("Stripe webhook handled")

	a.WriteJSON(w, http.StatusOK, module.BillingWebhookResponse{
		Received: true,
	})
}
