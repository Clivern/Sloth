// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitEvent(t *testing.T) {
	t.Run("get hides events from another workspace", func(t *testing.T) {
		events := NewEvent(&fakeEvents{item: &db.Event{Id: "evt-1", WorkspaceId: "ws-2"}})

		_, err := events.GetEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrEventNotFound)
	})

	t.Run("get returns not found for a missing event", func(t *testing.T) {
		events := NewEvent(&fakeEvents{})

		_, err := events.GetEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrEventNotFound)
	})

	t.Run("get returns the event", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		events := NewEvent(&fakeEvents{
			item: &db.Event{
				Id:          "evt-1",
				WorkspaceId: "ws-1",
				Type:        EventCatalogChanged,
				ActorName:   new("Ada"),
				Data:        new(`{"agent":"weather"}`),
				CreatedAt:   createdAt,
			},
		})

		response, err := events.GetEvent("ws-1", "evt-1")
		require.NoError(t, err)
		assert.Equal(t, EventCatalogChanged, response.Type)
		assert.Equal(t, "Ada", *response.ActorName)
		assert.JSONEq(t, `{"agent":"weather"}`, string(response.Data))
		assert.Equal(t, "2026-01-02T03:04:05Z", response.CreatedAt)
	})

	t.Run("get wraps the repository error", func(t *testing.T) {
		events := NewEvent(&fakeEvents{err: errors.New("boom")})

		_, err := events.GetEvent("ws-1", "evt-1")
		assert.ErrorIs(t, err, ErrFailedGetEvent)
	})

	t.Run("list returns the events and the total", func(t *testing.T) {
		events := NewEvent(&fakeEvents{
			list:  []*db.Event{{Id: "evt-1", WorkspaceId: "ws-1", Type: EventAuditLog}},
			total: 7,
		})

		result, err := events.ListEvents("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(7), result.Total)
		require.Len(t, result.Events, 1)
		assert.Equal(t, EventAuditLog, result.Events[0].Type)
	})

	t.Run("list wraps the repository error", func(t *testing.T) {
		events := NewEvent(&fakeEvents{err: errors.New("boom")})

		_, err := events.ListEvents("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrFailedListEvents)
	})
}
