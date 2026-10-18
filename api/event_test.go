// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/stretchr/testify/assert"
)

func TestUnitEventActions(t *testing.T) {
	stored := &db.Event{
		Id:          "event-1",
		WorkspaceId: "ws-1",
		Type:        "catalog.changed",
		Data:        new(`{"agent":"weather"}`),
		CreatedAt:   time.Now(),
	}

	t.Run("list returns events with pagination meta", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{events: []*db.Event{stored}, total: 1})
		w := httptest.NewRecorder()
		newEventRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/events?limit=5&offset=0", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"type":"catalog.changed"`)
		assert.Contains(t, w.Body.String(), `"data":{"agent":"weather"}`)
		assert.Contains(t, w.Body.String(), `"total":1`)
	})

	t.Run("list returns 500 when the repository fails", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{listErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newEventRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/events", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_list_events")
	})

	t.Run("list returns 400 when the workspace id is missing", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{})
		w := httptest.NewRecorder()
		api.ListEventsAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("get returns one event", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{event: stored})
		w := httptest.NewRecorder()
		newEventRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/events/event-1", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"event-1"`)
	})

	t.Run("get returns 404 for an event in another workspace", func(t *testing.T) {
		other := *stored
		other.WorkspaceId = "ws-2"
		api := newEventAPI(&fakeEvents{event: &other})
		w := httptest.NewRecorder()
		newEventRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/events/event-1", nil))

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "event_not_found")
	})

	t.Run("get returns 500 when the repository fails", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{getErr: errors.New("boom")})
		w := httptest.NewRecorder()
		newEventRouter(api).
			ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/ws-1/events/event-1", nil))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed_get_event")
	})

	t.Run("get returns 400 when ids are missing", func(t *testing.T) {
		api := newEventAPI(&fakeEvents{})
		w := httptest.NewRecorder()
		api.GetEventAction(w, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
