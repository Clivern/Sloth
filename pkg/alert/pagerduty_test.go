// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitPagerDuty(t *testing.T) {
	t.Run("Send triggers an incident", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"routing_key":"rk"`)
			assert.Contains(t, string(body), `"event_action":"trigger"`)
			assert.Contains(t, string(body), `"dedup_key":"db-1"`)
			assert.Contains(t, string(body), `"severity":"critical"`)
			assert.Contains(t, string(body), `"message":"down"`)
			w.Write([]byte(`{"status":"success"}`))
		}))
		defer server.Close()

		client := NewPagerDuty(PagerDutyConfig{RoutingKey: "rk"})
		client.endpoint = server.URL

		err := client.Send(context.Background(), Alert{
			Key:      "db-1",
			Status:   StatusFiring,
			Severity: SeverityCritical,
			Title:    "Database down",
			Message:  "down",
		})
		assert.NoError(t, err)
	})

	t.Run("Send resolves an incident", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"event_action":"resolve"`)
			w.Write([]byte(`{"status":"success"}`))
		}))
		defer server.Close()

		client := NewPagerDuty(PagerDutyConfig{RoutingKey: "rk"})
		client.endpoint = server.URL

		err := client.Send(context.Background(), Alert{Key: "db-1", Status: StatusResolved, Title: "Recovered"})
		assert.NoError(t, err)
	})

	t.Run("Send returns an error on a rejected event", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Invalid routing key"}`))
		}))
		defer server.Close()

		client := NewPagerDuty(PagerDutyConfig{RoutingKey: "bad"})
		client.endpoint = server.URL

		err := client.Send(context.Background(), Alert{Key: "db-1", Status: StatusFiring, Title: "Database down"})
		assert.ErrorContains(t, err, "status 400")
	})
}
