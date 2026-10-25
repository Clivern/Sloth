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
	"strings"
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
		assert.Equal(t, "https://events.eu.pagerduty.com", NewPagerDuty("events.eu.pagerduty.com", "rk", nil).endpoint)
	})

	t.Run("pagerduty records an info event as a change, not an incident", func(t *testing.T) {
		var body []byte
		var path string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			path = r.URL.Path
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		client := NewPagerDuty("events.pagerduty.com", "rk", map[string]string{"agent.instance.changed": "info"})
		client.endpoint = server.URL

		require.NoError(t, client.Send(context.Background(), Event{
			Id:         "task-1",
			Event:      "agent.instance.changed",
			OccurredAt: "2026-10-21T10:00:00Z",
			Data:       map[string]string{"action": "deregistered", "agentName": "users", "instanceId": "users-9203"},
		}))
		assert.Equal(t, "/v2/change/enqueue", path)
		assert.Contains(t, string(body), `"routing_key":"rk"`)
		assert.Contains(t, string(body), `"summary":"users-9203 (users) deregistered"`)
		assert.Contains(t, string(body), `"timestamp":"2026-10-21T10:00:00Z"`)
		assert.NotContains(t, string(body), `event_action`)
	})

	t.Run("pagerduty triggers an incident with the event severity", func(t *testing.T) {
		var body []byte
		var path string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			path = r.URL.Path
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		client := NewPagerDuty("events.pagerduty.com", "rk", map[string]string{"catalog.changed": "warning"})
		client.endpoint = server.URL

		require.NoError(t, client.Send(context.Background(), event))
		assert.Equal(t, "/v2/enqueue", path)
		assert.Contains(t, string(body), `"event_action":"trigger"`)
		assert.Contains(t, string(body), `"dedup_key":"task-1"`)
		assert.Contains(t, string(body), `"severity":"warning"`)
		assert.Contains(t, string(body), `"agent":"weather"`)
	})

	t.Run("pagerduty shares one incident per resource for status changes", func(t *testing.T) {
		var bodies []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			bodies = append(bodies, string(body))
			w.WriteHeader(http.StatusAccepted)
		}))
		defer server.Close()

		client := NewPagerDuty("events.pagerduty.com", "rk", map[string]string{"agent.health.changed": "critical"})
		client.endpoint = server.URL

		failing := Event{
			Id:           "task-1",
			Event:        "agent.health.changed",
			WorkspaceId:  "ws-1",
			ResourceType: "health_check",
			ResourceId:   "check-1",
			Data: map[string]string{
				"agentName":  "users",
				"instanceId": "users-9203",
				"checkId":    "http-health",
				"from":       "passing",
				"to":         "warning",
			},
		}
		require.NoError(t, client.Send(context.Background(), failing))

		recovered := failing
		recovered.Id = "task-2"
		recovered.Data = map[string]string{
			"agentName":  "users",
			"instanceId": "users-9203",
			"checkId":    "http-health",
			"from":       "warning",
			"to":         "passing",
		}
		require.NoError(t, client.Send(context.Background(), recovered))

		require.Len(t, bodies, 2)
		assert.Contains(t, bodies[0], `"event_action":"trigger"`)
		assert.Contains(t, bodies[0], `"dedup_key":"sloth:ws-1:health_check:check-1"`)
		assert.Contains(t, bodies[0], `"severity":"warning"`)
		assert.Contains(t, bodies[0], `"summary":"users degraded: http-health warning on users-9203"`)
		assert.Contains(t, bodies[1], `"event_action":"resolve"`)
		assert.Contains(t, bodies[1], `"dedup_key":"sloth:ws-1:health_check:check-1"`)
	})
}

func TestUnitTitle(t *testing.T) {
	health := func(to, output string) Event {
		return Event{
			Event: "agent.health.changed",
			Data: map[string]string{
				"agentName":  "users",
				"instanceId": "users-9203",
				"checkId":    "http-health",
				"from":       "passing",
				"to":         to,
				"output":     output,
			},
		}
	}

	t.Run("names a failing check with its output", func(t *testing.T) {
		assert.Equal(
			t,
			"users unhealthy: http-health critical on users-9203 (connection refused)",
			Title(health("critical", "connection refused")),
		)
	})

	t.Run("cuts long output", func(t *testing.T) {
		title := Title(health("warning", strings.Repeat("x", 200)))
		assert.True(t, strings.HasSuffix(title, "…)"))
		assert.Len(t, []rune(strings.TrimPrefix(title, "users degraded: http-health warning on users-9203 (")), 81)
	})

	t.Run("leaves the output off a recovery", func(t *testing.T) {
		assert.Equal(t, "users recovered: http-health passing on users-9203", Title(health("passing", "200 OK")))
	})

	t.Run("names an unknown status", func(t *testing.T) {
		assert.Equal(t, "users status unknown: http-health unknown on users-9203", Title(health("unknown", "")))
	})

	t.Run("names instance changes with their endpoint", func(t *testing.T) {
		data := map[string]string{
			"agentName":  "users",
			"instanceId": "users-9203",
			"scheme":     "http",
			"address":    "127.0.0.1",
			"port":       "9203",
			"path":       "/a2a/v1",
		}

		data["action"] = "registered"
		assert.Equal(t, "users-9203 (users) registered at http://127.0.0.1:9203/a2a/v1", Title(Event{Data: data}))

		data["action"] = "updated"
		assert.Equal(t, "users-9203 (users) moved to http://127.0.0.1:9203/a2a/v1", Title(Event{Data: data}))

		data["action"] = "deregistered"
		assert.Equal(t, "users-9203 (users) deregistered", Title(Event{Data: data}))
	})

	t.Run("falls back to the event name", func(t *testing.T) {
		assert.Equal(t, "catalog.changed", Title(Event{Event: "catalog.changed", Data: map[string]string{"agent": "weather"}}))
	})
}
