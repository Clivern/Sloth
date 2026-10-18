// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestUnitHelpers(t *testing.T) {
	t.Run("AppURL", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		viper.Set("app.url", "https://example.com/")

		assert.Equal(t, "https://example.com/login", AppURL("/login"))
	})

	t.Run("CurrentMonthPeriod", func(t *testing.T) {
		start, end := CurrentMonthPeriod()
		now := time.Now().UTC()

		assert.Equal(t, time.UTC, start.Location())
		assert.Equal(t, 1, start.Day())
		assert.Equal(t, now.Month(), start.Month())
		assert.Equal(t, now.Year(), start.Year())
		assert.True(t, end.Equal(start.AddDate(0, 1, 0)))
	})

	t.Run("WriteJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
	})

	t.Run("GenerateUUID", func(t *testing.T) {
		id, err := GenerateUUID()
		assert.NoError(t, err)
		assert.Len(t, id, 36)
	})

	t.Run("GenerateSecureToken", func(t *testing.T) {
		token, err := GenerateSecureToken(16)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("RandInt", func(t *testing.T) {
		n, err := RandInt(1, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, n, 1)
		assert.LessOrEqual(t, n, 10)
	})

	t.Run("ParseQueryLabels", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/?id=1&limit=10&offset=0&env=prod&team=ops", nil)
		labels := ParseQueryLabels(r)
		assert.Equal(t, map[string]string{"env": "prod", "team": "ops"}, labels)
	})

	t.Run("ParsePagination", func(t *testing.T) {
		defaults := httptest.NewRequest(http.MethodGet, "/", nil)
		limit, offset := ParsePagination(defaults)
		assert.Equal(t, 50, limit)
		assert.Equal(t, 0, offset)

		custom := httptest.NewRequest(http.MethodGet, "/?limit=25&offset=5", nil)
		limit, offset = ParsePagination(custom)
		assert.Equal(t, 25, limit)
		assert.Equal(t, 5, offset)

		invalid := httptest.NewRequest(http.MethodGet, "/?limit=999&offset=-1", nil)
		limit, offset = ParsePagination(invalid)
		assert.Equal(t, 50, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("RandomHandle", func(t *testing.T) {
		handle, err := RandomHandle(3, 4)
		assert.NoError(t, err)
		assert.Contains(t, handle, "-")
	})

	t.Run("RemoveLabelFromJSON", func(t *testing.T) {
		raw := `["a","b","c"]`
		next, ok := RemoveLabelFromJSON(&raw, "b")
		assert.True(t, ok)
		assert.JSONEq(t, `["a","c"]`, *next)

		empty, ok := RemoveLabelFromJSON(&raw, "missing")
		assert.False(t, ok)
		assert.Equal(t, &raw, empty)

		last := `["only"]`
		cleared, ok := RemoveLabelFromJSON(&last, "only")
		assert.True(t, ok)
		assert.Nil(t, cleared)

		var nilLabels *string
		unchanged, ok := RemoveLabelFromJSON(nilLabels, "a")
		assert.False(t, ok)
		assert.Nil(t, unchanged)
	})

	t.Run("JSONRawFromString", func(t *testing.T) {
		raw := `{"a":1}`
		assert.Equal(t, json.RawMessage(raw), JSONRawFromString(&raw))
		assert.Nil(t, JSONRawFromString(nil))

		empty := ""
		assert.Nil(t, JSONRawFromString(&empty))
	})

	t.Run("JSONSliceFromString", func(t *testing.T) {
		raw := `["a","b"]`
		assert.Equal(t, []string{"a", "b"}, JSONSliceFromString[string](&raw))
		assert.Nil(t, JSONSliceFromString[string](nil))

		invalid := `{not-json`
		assert.Nil(t, JSONSliceFromString[string](&invalid))
	})
}
