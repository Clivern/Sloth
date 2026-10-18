// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package worker

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/pkg/broker"
	"github.com/clivern/sloth/pkg/encrypt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// deliveryTasks captures the status calls the delivery handler makes.
type deliveryTasks struct {
	db.AsyncTaskRepository
	completed []db.Id
	failed    []string
}

func (f *deliveryTasks) MarkRunning(id db.Id) error {
	return nil
}

func (f *deliveryTasks) Complete(id db.Id, result string) error {
	f.completed = append(f.completed, id)
	return nil
}

func (f *deliveryTasks) Fail(id db.Id, message string) error {
	f.failed = append(f.failed, message)
	return nil
}

// deliveryIntegrations returns one stored integration.
type deliveryIntegrations struct {
	db.IntegrationRepository
	item *db.Integration
}

func (f *deliveryIntegrations) GetById(id db.Id) (*db.Integration, error) {
	return f.item, nil
}

func deliveryMsg(t *testing.T) *broker.Msg {
	t.Helper()
	raw, err := json.Marshal(map[string]string{
		"taskId":        "task-1",
		"integrationId": "int-1",
		"eventId":       "event-1",
		"event":         "catalog.changed",
		"actorId":       "user-1",
		"actorName":     "Ada",
		"actorType":     "user",
		"data":          `{"agent":"weather"}`,
		"occurredAt":    "2026-10-21T10:00:00Z",
	})
	require.NoError(t, err)
	return &broker.Msg{Data: raw}
}

func TestUnitHandleIntegrationDelivery(t *testing.T) {
	sealed, err := encrypt.Encrypt("whsec_test")
	require.NoError(t, err)

	t.Run("webhook posts a signed body and completes the task", func(t *testing.T) {
		var body []byte
		var signature, event string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			signature = r.Header.Get("X-Sloth-Signature")
			event = r.Header.Get("X-Sloth-Event")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		tasks := &deliveryTasks{}
		h := &handlers{tasks: tasks, integrations: &deliveryIntegrations{item: &db.Integration{
			Id:          "int-1",
			WorkspaceId: "ws-1",
			Type:        db.IntegrationTypeWebhook,
			Config:      new(`{"url":"` + server.URL + `","events":["catalog.changed"],"secretEncrypted":"` + sealed + `"}`),
			IsActive:    true,
		}}}

		require.NoError(t, h.HandleIntegrationDelivery(context.Background(), deliveryMsg(t)))

		mac := hmac.New(sha256.New, []byte("whsec_test"))
		mac.Write(body)
		assert.Equal(t, fmt.Sprintf("sha256=%s", hex.EncodeToString(mac.Sum(nil))), signature)
		assert.Equal(t, "catalog.changed", event)
		assert.JSONEq(t, `{"id":"task-1","eventId":"event-1","event":"catalog.changed","workspaceId":"ws-1","actorId":"user-1","actorName":"Ada","actorType":"user","occurredAt":"2026-10-21T10:00:00Z","data":{"agent":"weather"}}`, string(body))
		assert.Equal(t, []db.Id{"task-1"}, tasks.completed)
	})

	t.Run("non 2xx response fails the task", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		tasks := &deliveryTasks{}
		h := &handlers{tasks: tasks, integrations: &deliveryIntegrations{item: &db.Integration{
			Type:     db.IntegrationTypeSlack,
			Config:   new(`{"url":"` + server.URL + `","events":["catalog.changed"]}`),
			IsActive: true,
		}}}

		require.NoError(t, h.HandleIntegrationDelivery(context.Background(), deliveryMsg(t)))
		require.Len(t, tasks.failed, 1)
		assert.Contains(t, tasks.failed[0], "status 502")
		assert.Empty(t, tasks.completed)
	})

	t.Run("disabled integration completes without sending", func(t *testing.T) {
		tasks := &deliveryTasks{}
		h := &handlers{tasks: tasks, integrations: &deliveryIntegrations{item: &db.Integration{
			Type:     db.IntegrationTypeWebhook,
			Config:   new(`{"url":"http://127.0.0.1:1"}`),
			IsActive: false,
		}}}

		require.NoError(t, h.HandleIntegrationDelivery(context.Background(), deliveryMsg(t)))
		assert.Equal(t, []db.Id{"task-1"}, tasks.completed)
		assert.Empty(t, tasks.failed)
	})

	t.Run("invalid payload returns ErrInvalidPayload", func(t *testing.T) {
		h := &handlers{tasks: &deliveryTasks{}}

		err := h.HandleIntegrationDelivery(context.Background(), &broker.Msg{Data: []byte("not json")})
		assert.ErrorIs(t, err, ErrInvalidPayload)
	})
}
