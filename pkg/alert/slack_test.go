// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package alert

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/cognit/pkg/slack"
	"github.com/stretchr/testify/assert"
)

func TestUnitSlack(t *testing.T) {
	t.Run("Send posts the alert text to the webhook", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), `[critical] [firing] Database down`)
				w.Write([]byte("ok"))
			}),
		)
		defer server.Close()

		client := NewSlack(slack.Config{WebhookURL: server.URL, Channel: "#alerts"})

		err := client.Send(context.Background(), Alert{
			Status:   StatusFiring,
			Severity: SeverityCritical,
			Title:    "Database down",
			Message:  "primary is unreachable",
		})
		assert.NoError(t, err)
	})
}
