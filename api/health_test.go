// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntegrationHealthEndpoint(t *testing.T) {
	t.Run("HealthAction should return OK status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/_health", nil)
		w := httptest.NewRecorder()

		(&API{}).HealthAction(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, `{"status":"ok"}`, strings.TrimSpace(w.Body.String()))
	})
}
