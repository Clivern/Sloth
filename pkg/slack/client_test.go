// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package slack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitSlackClient(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		client := New(Config{Token: "xoxb", Channel: "#general"})
		assert.Equal(t, "xoxb", client.config.Token)
		assert.Equal(t, "#general", client.config.Channel)
	})

	t.Run("Notify webhook", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				assert.Contains(t, string(body), `"text":"hello"`)
				w.Write([]byte("ok"))
			}),
		)
		defer server.Close()

		client := New(Config{WebhookURL: server.URL, Channel: "#alerts"})
		assert.NoError(t, client.Notify(context.Background(), "hello"))
	})

	t.Run("Notify chat.postMessage", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer xoxb-token", r.Header.Get("Authorization"))
				w.Write([]byte(`{"ok":true}`))
			}),
		)
		defer server.Close()

		client := New(Config{Token: "xoxb-token", Channel: "#general"})
		_, err := client.Post(
			context.Background(),
			server.URL,
			client.config.Token,
			Message{Text: "hi", Channel: "#general"},
			nil,
		)
		assert.NoError(t, err)
	})
}
