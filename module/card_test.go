// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitCard(t *testing.T) {
	serve := func(t *testing.T, handler http.HandlerFunc) (string, int) {
		server := httptest.NewServer(handler)
		t.Cleanup(server.Close)

		host, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
		require.NoError(t, err)
		number, err := strconv.Atoi(port)
		require.NoError(t, err)

		return host, number
	}

	t.Run("falls back to http and returns the card", func(t *testing.T) {
		host, port := serve(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, DefaultCardPath, r.URL.Path)
			w.Write([]byte(`{"name":"Support","version":"1.0.0"}`))
		})

		card, err := NewCard().Fetch(context.Background(), host, port, DefaultCardPath)
		require.NoError(t, err)
		assert.Equal(t, "Support", card["name"])
	})

	t.Run("fails on a non 2xx response", func(t *testing.T) {
		host, port := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		_, err := NewCard().Fetch(context.Background(), host, port, DefaultCardPath)
		assert.Error(t, err)
	})

	t.Run("rejects a body that is not a JSON object", func(t *testing.T) {
		host, port := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`[1,2]`))
		})

		_, err := NewCard().Fetch(context.Background(), host, port, DefaultCardPath)
		assert.ErrorIs(t, err, ErrInvalidAgentCard)
	})

	t.Run("rejects a card larger than the limit", func(t *testing.T) {
		host, port := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`{"name":"` + strings.Repeat("a", MaxCardSize) + `"}`))
		})

		_, err := NewCard().Fetch(context.Background(), host, port, DefaultCardPath)
		assert.ErrorIs(t, err, ErrCardTooLarge)
	})
}
