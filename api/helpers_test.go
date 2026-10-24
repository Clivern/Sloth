// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http/httptest"
	"testing"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/module"

	"github.com/stretchr/testify/assert"
)

func TestUnitEventActor(t *testing.T) {
	t.Run("names the access key behind the request", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/", nil)
		r.Header.Set("User-Agent", "python-httpx/0.28.1")
		r = r.WithContext(module.WithPrincipal(r.Context(), &module.Principal{
			AccessKey: &db.AccessKey{Id: "key-1", Name: "demo"},
		}))

		actor := (&API{}).EventActor(r)
		assert.Equal(t, db.EventActorTypeAccessKey, actor.ActorType)
		assert.Equal(t, db.Id("key-1"), actor.ActorId)
		assert.Equal(t, "demo", actor.ActorName)
		assert.Equal(t, "python-httpx/0.28.1", actor.UserAgent)
		assert.NotEmpty(t, actor.IPAddress)
	})

	t.Run("names the user behind the request", func(t *testing.T) {
		r := httptest.NewRequest("PUT", "/", nil)
		r = r.WithContext(module.WithPrincipal(r.Context(), &module.Principal{
			User: &db.User{Id: "user-1", Name: "Ahmed"},
		}))

		actor := (&API{}).EventActor(r)
		assert.Equal(t, db.EventActorTypeUser, actor.ActorType)
		assert.Equal(t, db.Id("user-1"), actor.ActorId)
		assert.Equal(t, "Ahmed", actor.ActorName)
	})
}
