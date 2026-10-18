// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitIntegration(t *testing.T) {
	t.Run("list returns integrations without their secrets", func(t *testing.T) {
		repo := &fakeIntegrations{
			list: []*db.Integration{
				{
					Id:       "int-1",
					Type:     db.IntegrationTypeSplunk,
					Name:     "splunk",
					Config:   new(`{"url":"https://splunk.test:8088","index":"audit","events":["audit.log"],"secretEncrypted":"sealed"}`),
					IsActive: true,
				},
			},
			total: 3,
		}

		result, err := NewIntegration(repo).ListIntegrations("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Equal(t, "https://splunk.test:8088", result.Integrations[0].URL)
		assert.Equal(t, "audit", result.Integrations[0].Index)
		assert.Equal(t, []string{"audit.log"}, result.Integrations[0].Events)
		assert.True(t, result.Integrations[0].HasSecret)
		assert.Empty(t, result.Integrations[0].Secret)
	})

	t.Run("list wraps the repository error", func(t *testing.T) {
		_, err := NewIntegration(&fakeIntegrations{err: errors.New("boom")}).ListIntegrations("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrFailedListIntegrations)
	})

	t.Run("get hides integrations from another workspace", func(t *testing.T) {
		repo := &fakeIntegrations{item: &db.Integration{Id: "int-1", WorkspaceId: "ws-2"}}

		_, err := NewIntegration(repo).GetIntegration("ws-1", "int-1")
		assert.ErrorIs(t, err, ErrIntegrationNotFound)
	})

	t.Run("create rejects settings missing for the type", func(t *testing.T) {
		requests := []*SaveIntegrationRequest{
			{Type: db.IntegrationTypeWebhook, Name: "hook"},
			{Type: db.IntegrationTypeSlack, Name: "slack"},
			{Type: db.IntegrationTypeDatadog, Name: "dd", Site: "datadoghq.com"},
			{Type: db.IntegrationTypeDatadog, Name: "dd", Site: "example.com", Secret: "api-key-123"},
			{Type: db.IntegrationTypeSplunk, Name: "splunk", URL: "https://splunk.test"},
			{Type: db.IntegrationTypePagerDuty, Name: "pd", Site: "events.pagerduty.com"},
			{Type: db.IntegrationTypePagerDuty, Name: "pd", Site: "example.com", Secret: "routing-key-1"},
		}

		for _, req := range requests {
			_, err := NewIntegration(&fakeIntegrations{}).CreateIntegration("ws-1", req)
			assert.ErrorIs(t, err, ErrInvalidIntegrationConfig, req.Name)
		}
	})

	t.Run("create generates a webhook signing secret and returns it once", func(t *testing.T) {
		repo := &fakeIntegrations{}

		response, err := NewIntegration(repo).CreateIntegration(
			"ws-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypeWebhook, Name: "hook", URL: "https://example.test/hook", Events: []string{"audit.log", "gateway.call.denied"}, IsActive: true},
		)
		require.NoError(t, err)
		require.Len(t, repo.created, 1)
		assert.True(t, strings.HasPrefix(response.Secret, "whsec_"))
		assert.True(t, response.HasSecret)
		assert.True(t, repo.created[0].IsActive)
		assert.Equal(t, []string{"audit.log", "gateway.call.denied"}, response.Events)
		assert.NotContains(t, *repo.created[0].Config, response.Secret)

		var config IntegrationConfig
		require.NoError(t, json.Unmarshal([]byte(*repo.created[0].Config), &config))
		assert.Equal(t, []string{"audit.log", "gateway.call.denied"}, config.Events)

		secret, err := encrypt.Decrypt(config.SecretEncrypted)
		require.NoError(t, err)
		assert.Equal(t, response.Secret, secret)
	})

	t.Run("create stores a given secret without echoing it", func(t *testing.T) {
		repo := &fakeIntegrations{}

		response, err := NewIntegration(repo).CreateIntegration(
			"ws-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypeDatadog, Name: "dd", Site: "datadoghq.eu", Secret: "dd-api-key-1"},
		)
		require.NoError(t, err)
		assert.Empty(t, response.Secret)
		assert.True(t, response.HasSecret)
		assert.Equal(t, "datadoghq.eu", response.Site)
	})

	t.Run("create stores a pagerduty routing key with its region", func(t *testing.T) {
		repo := &fakeIntegrations{}

		response, err := NewIntegration(repo).CreateIntegration(
			"ws-1",
			&SaveIntegrationRequest{
				Type:   db.IntegrationTypePagerDuty,
				Name:   "on-call",
				Site:   "events.eu.pagerduty.com",
				Secret: "routing-key-1",
				Events: []string{"agent.health.changed"},
			},
		)
		require.NoError(t, err)
		assert.Empty(t, response.Secret)
		assert.True(t, response.HasSecret)
		assert.Equal(t, "events.eu.pagerduty.com", response.Site)
		assert.NotContains(t, *repo.created[0].Config, "routing-key-1")
	})

	t.Run("update rejects a pagerduty region outside the list", func(t *testing.T) {
		repo := &fakeIntegrations{item: &db.Integration{Id: "int-1", WorkspaceId: "ws-1", Type: db.IntegrationTypePagerDuty, Config: new(`{}`)}}

		_, err := NewIntegration(repo).UpdateIntegration(
			"ws-1",
			"int-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypePagerDuty, Name: "pd", Site: "example.com"},
		)
		assert.ErrorIs(t, err, ErrInvalidIntegrationConfig)
	})

	t.Run("create leaves slack without a secret", func(t *testing.T) {
		response, err := NewIntegration(&fakeIntegrations{}).CreateIntegration(
			"ws-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypeSlack, Name: "slack", URL: "https://hooks.slack.com/services/T/B/X"},
		)
		require.NoError(t, err)
		assert.False(t, response.HasSecret)
		assert.Empty(t, response.Secret)
	})

	t.Run("update rejects a type change", func(t *testing.T) {
		repo := &fakeIntegrations{item: &db.Integration{Id: "int-1", WorkspaceId: "ws-1", Type: db.IntegrationTypeWebhook, Config: new(`{}`)}}

		_, err := NewIntegration(repo).UpdateIntegration(
			"ws-1",
			"int-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypeSlack, Name: "slack", URL: "https://hooks.slack.com/x"},
		)
		assert.ErrorIs(t, err, ErrInvalidIntegrationConfig)
	})

	t.Run("update keeps the stored secret when none is given", func(t *testing.T) {
		repo := &fakeIntegrations{
			item: &db.Integration{
				Id:          "int-1",
				WorkspaceId: "ws-1",
				Type:        db.IntegrationTypeSplunk,
				Config:      new(`{"url":"https://old.test","secretEncrypted":"sealed"}`),
				IsActive:    true,
			},
		}

		response, err := NewIntegration(repo).UpdateIntegration(
			"ws-1",
			"int-1",
			&SaveIntegrationRequest{Type: db.IntegrationTypeSplunk, Name: "splunk", URL: "https://new.test", Events: []string{"system.health.degraded"}},
		)
		require.NoError(t, err)
		require.Len(t, repo.updated, 1)
		assert.False(t, repo.updated[0].IsActive)
		assert.Contains(t, *repo.updated[0].Config, `"secretEncrypted":"sealed"`)
		assert.Equal(t, "https://new.test", response.URL)
		assert.Equal(t, []string{"system.health.degraded"}, response.Events)
		assert.True(t, response.HasSecret)
	})

	t.Run("delete returns not found for another workspace", func(t *testing.T) {
		repo := &fakeIntegrations{item: &db.Integration{Id: "int-1", WorkspaceId: "ws-2"}}

		err := NewIntegration(repo).DeleteIntegration("ws-1", "int-1")
		assert.ErrorIs(t, err, ErrIntegrationNotFound)
	})

	t.Run("delete removes the integration", func(t *testing.T) {
		repo := &fakeIntegrations{item: &db.Integration{Id: "int-1", WorkspaceId: "ws-1"}}

		require.NoError(t, NewIntegration(repo).DeleteIntegration("ws-1", "int-1"))
		assert.Equal(t, []db.Id{"int-1"}, repo.deleted)
	})
}
