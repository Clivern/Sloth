// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitSenders(t *testing.T) {
	event := Event{
		Id:          "task-1",
		EventId:     "event-1",
		Event:       "catalog.changed",
		WorkspaceId: "ws-1",
		OccurredAt:  "2026-10-21T10:00:00Z",
		Data:        map[string]string{"agent": "weather"},
	}

	t.Run("webhook posts a signed body", func(t *testing.T) {
		var body []byte
		var signature, name, delivery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			signature = r.Header.Get("X-Sloth-Signature")
			name = r.Header.Get("X-Sloth-Event")
			delivery = r.Header.Get("X-Sloth-Delivery")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		require.NoError(t, NewWebhook(server.URL, "whsec_test").Send(context.Background(), event))

		mac := hmac.New(sha256.New, []byte("whsec_test"))
		mac.Write(body)
		assert.Equal(t, fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil))), signature)
		assert.Equal(t, "catalog.changed", name)
		assert.Equal(t, "task-1", delivery)
		assert.JSONEq(t, `{"id":"task-1","eventId":"event-1","event":"catalog.changed","workspaceId":"ws-1","occurredAt":"2026-10-21T10:00:00Z","data":{"agent":"weather"}}`, string(body))
	})

	t.Run("webhook returns an error on a non 2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		err := NewWebhook(server.URL, "whsec_test").Send(context.Background(), event)
		assert.ErrorContains(t, err, "status 502")
	})

	t.Run("slack posts the event as text", func(t *testing.T) {
		var body map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}))
		defer server.Close()

		require.NoError(t, NewSlack(server.URL).Send(context.Background(), event))
		assert.Equal(t, "*catalog.changed*\n• *agent*: weather", body["text"])
	})

	t.Run("datadog builds the site intake url", func(t *testing.T) {
		assert.Equal(t, "https://http-intake.logs.datadoghq.eu/api/v2/logs", NewDatadog("datadoghq.eu", "key").endpoint)
	})

	t.Run("datadog posts a tagged log entry with the api key", func(t *testing.T) {
		var apiKey string
		var body []map[string]string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey = r.Header.Get("DD-API-KEY")
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		client := NewDatadog("datadoghq.com", "dd-key")
		client.endpoint = server.URL

		require.NoError(t, client.Send(context.Background(), event))
		assert.Equal(t, "dd-key", apiKey)
		require.Len(t, body, 1)
		assert.Equal(t, "event:catalog.changed,workspace:ws-1", body[0]["ddtags"])
		assert.Contains(t, body[0]["message"], `"eventId":"event-1"`)
	})

	t.Run("splunk posts to the collector with the token and index", func(t *testing.T) {
		var path, auth string
		var body map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			auth = r.Header.Get("Authorization")
			_ = json.NewDecoder(r.Body).Decode(&body)
		}))
		defer server.Close()

		require.NoError(t, NewSplunk(server.URL+"/", "hec-token", "audit").Send(context.Background(), event))
		assert.Equal(t, "/services/collector/event", path)
		assert.Equal(t, "Splunk hec-token", auth)
		assert.Equal(t, "audit", body["index"])
	})

	t.Run("pagerduty builds the region url", func(t *testing.T) {
		assert.Equal(t, "https://events.eu.pagerduty.com/v2/enqueue", NewPagerDuty("events.eu.pagerduty.com", "rk", nil).endpoint)
	})

	t.Run("pagerduty triggers an incident with the event severity", func(t *testing.T) {
		var body []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		client := NewPagerDuty("events.pagerduty.com", "rk", map[string]string{"catalog.changed": "info"})
		client.endpoint = server.URL

		require.NoError(t, client.Send(context.Background(), event))
		assert.Contains(t, string(body), `"routing_key":"rk"`)
		assert.Contains(t, string(body), `"event_action":"trigger"`)
		assert.Contains(t, string(body), `"dedup_key":"task-1"`)
		assert.Contains(t, string(body), `"severity":"info"`)
		assert.Contains(t, string(body), `"agent":"weather"`)
	})
}
