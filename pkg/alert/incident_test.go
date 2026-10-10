// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitIncident(t *testing.T) {
	t.Run("Endpoint includes the alert source config ID", func(t *testing.T) {
		client := NewIncident(IncidentConfig{Token: "secret", AlertSourceConfigID: "01ABC"})
		assert.Equal(t, "https://api.incident.io/v2/alert_events/http/01ABC", client.endpoint)
	})

	t.Run("Send posts a firing alert with bearer auth", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))

			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), `"status":"firing"`)
			assert.Contains(t, string(body), `"title":"Database down"`)
			assert.Contains(t, string(body), `"deduplication_key":"db-1"`)
			assert.Contains(t, string(body), `"severity":"critical"`)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewIncident(IncidentConfig{Token: "secret", AlertSourceConfigID: "01ABC"})
		client.endpoint = server.URL

		err := client.Send(context.Background(), Alert{
			Key:      "db-1",
			Status:   StatusFiring,
			Severity: SeverityCritical,
			Title:    "Database down",
		})
		assert.NoError(t, err)
	})

	t.Run("Send returns an error on a rejected event", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{}`)
		}))
		defer server.Close()

		client := NewIncident(IncidentConfig{Token: "bad", AlertSourceConfigID: "01ABC"})
		client.endpoint = server.URL

		err := client.Send(context.Background(), Alert{Key: "db-1", Status: StatusFiring, Title: "Database down"})
		assert.ErrorContains(t, err, "status 401")
	})
}
