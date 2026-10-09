// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationTokenPurchaseRepository(t *testing.T) {
	database := openTestDB(t)
	workspace := createTestWorkspace(t, database)
	repo := NewTokenPurchaseRepository(database)
	session := "cs_" + uuid.NewString()

	t.Run("create", func(t *testing.T) {
		created, err := repo.Create(&TokenPurchase{
			WorkspaceId:     workspace.Id,
			StripeSessionId: session,
			AmountCents:     500,
			Tokens:          1000,
		})
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("duplicate session is ignored", func(t *testing.T) {
		created, err := repo.Create(&TokenPurchase{
			WorkspaceId:     workspace.Id,
			StripeSessionId: session,
			AmountCents:     500,
			Tokens:          1000,
		})
		require.NoError(t, err)
		assert.False(t, created)
	})
}
