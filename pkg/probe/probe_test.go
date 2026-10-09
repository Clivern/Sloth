// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package probe

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnitProbe(t *testing.T) {
	t.Run("HTTP passing on 2xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/health", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		result := HTTP(context.Background(), HTTPTarget{URL: server.URL + "/health"})
		assert.Equal(t, Passing, result.Status)
	})

	t.Run("HTTP critical on 5xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		result := HTTP(context.Background(), HTTPTarget{URL: server.URL})
		assert.Equal(t, Critical, result.Status)
		assert.Contains(t, result.Output, "500")
	})

	t.Run("HTTP critical when unreachable", func(t *testing.T) {
		host, port := closedPort(t)
		result := HTTP(context.Background(), HTTPTarget{URL: "http://" + net.JoinHostPort(host, strconv.Itoa(port)), Timeout: time.Second})
		assert.Equal(t, Critical, result.Status)
	})

	t.Run("HTTP critical on empty url", func(t *testing.T) {
		result := HTTP(context.Background(), HTTPTarget{})
		assert.Equal(t, Critical, result.Status)
		assert.Contains(t, result.Output, "invalid target")
	})

	t.Run("TCP passing when port open", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		assert.NoError(t, err)
		defer listener.Close()

		port := listener.Addr().(*net.TCPAddr).Port
		result := TCP(context.Background(), TCPTarget{Address: "127.0.0.1", Port: port})
		assert.Equal(t, Passing, result.Status)
	})

	t.Run("TCP critical when port closed", func(t *testing.T) {
		host, port := closedPort(t)
		result := TCP(context.Background(), TCPTarget{Address: host, Port: port, Timeout: time.Second})
		assert.Equal(t, Critical, result.Status)
	})

	t.Run("Output is capped", func(t *testing.T) {
		result := HTTP(context.Background(), HTTPTarget{URL: "http://" + strings.Repeat("a", MaxOutput+10) + ".invalid/", Timeout: time.Second})
		assert.Equal(t, Critical, result.Status)
		assert.LessOrEqual(t, len(result.Output), MaxOutput)
	})
}

func closedPort(t *testing.T) (string, int) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	return "127.0.0.1", port
}
