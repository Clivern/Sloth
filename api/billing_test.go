// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/clivern/cognit/db"
	"github.com/clivern/cognit/module"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnitBillingActions(t *testing.T) {
	user := &db.User{Id: "user-1"}
	workspace := &db.Workspace{Id: "ws-1", Name: "Platform", Handle: "platform"}
	subscription := &db.Subscription{
		Id:              "sub-1",
		WorkspaceId:     "ws-1",
		AITokensBalance: 500,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	newAPI := func(workspace *db.Workspace, subs *fakeSubscriptionRecords) *API {
		return &API{
			Billing: module.NewBilling(
				&fakeWorkspaceRecords{workspace: workspace},
				subs,
				nil,
				module.Usage{},
			),
			WorkspaceUsers: &fakeWorkspaceMembers{count: 3},
			Usage:          &fakeUsageRecords{},
		}
	}
	newRouter := func(api *API) *chi.Mux {
		router := chi.NewRouter()
		router.Get("/api/v1/workspaces/{workspaceId}/billing/status", api.GetBillingStatusAction)
		router.Get("/api/v1/workspaces/{workspaceId}/billing/usage", api.GetBillingUsageAction)
		router.Post("/api/v1/workspaces/{workspaceId}/billing/checkout", api.CreateBillingCheckoutAction)
		router.Post("/api/v1/workspaces/{workspaceId}/billing/portal", api.CreateBillingPortalAction)
		router.Post("/api/v1/billing/webhook", api.StripeWebhookAction)
		return router
	}
	base := "/api/v1/workspaces/ws-1/billing"
	checkoutBody := `{"amountCents":1000}`

	t.Run("status returns the subscription balance", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/status", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"aiTokensBalance":500`)
	})

	t.Run("status returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/status", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_not_found")
	})

	t.Run("status returns 404 when there is no subscription", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/status", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "workspace_subscription_not_found")
	})

	t.Run("status returns 500 when the repository fails", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{err: errors.New("boom")})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/status", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_billing_status")
	})

	t.Run("usage returns used tokens and limits", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/usage", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"periodReset"`)
	})

	t.Run("usage returns 404 for unknown workspace", func(t *testing.T) {
		api := newAPI(nil, &fakeSubscriptionRecords{})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodGet, base+"/usage", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("checkout returns 401 without a user", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, base+"/checkout", strings.NewReader(checkoutBody)))

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("checkout returns 400 for an amount below the minimum", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, base+"/checkout", strings.NewReader(`{"amountCents":50}`)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("checkout returns 503 when stripe billing is not configured", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		req := requestWithUser(
			httptest.NewRequest(http.MethodPost, base+"/checkout", strings.NewReader(checkoutBody)),
			user,
		)
		newRouter(api).ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "stripe_billing_not_configured")
	})

	t.Run("portal returns 503 when stripe billing is not configured", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{subscription: subscription})
		w := httptest.NewRecorder()
		newRouter(api).ServeHTTP(w, httptest.NewRequest(http.MethodPost, base+"/portal", nil))

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("webhook returns 503 when stripe billing is not configured", func(t *testing.T) {
		api := newAPI(workspace, &fakeSubscriptionRecords{})
		w := httptest.NewRecorder()
		newRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", strings.NewReader(`{}`)))

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
