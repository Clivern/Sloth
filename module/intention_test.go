// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/clivern/cognit/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitIntention(t *testing.T) {
	workspace := &fakeWorkspaces{workspace: &db.Workspace{Id: "ws-1"}}

	t.Run("list returns not found for a missing workspace", func(t *testing.T) {
		intentions := NewIntention(&fakeIntentions{}, &fakeWorkspaces{})

		_, err := intentions.ListIntentions("ws-1", 20, 0)
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("list returns the rules and the total", func(t *testing.T) {
		skill := "invoice.extract"
		repo := &fakeIntentions{list: []*db.Intention{{Id: "i-1", SourceAgent: "planner", DestinationAgent: "invoice", Skill: &skill, Action: "allow"}}, total: 2}

		result, err := NewIntention(repo, workspace).ListIntentions("ws-1", 20, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, "invoice.extract", result.Intentions[0].Skill)
	})

	t.Run("get hides rules from another workspace", func(t *testing.T) {
		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-2"}}

		_, err := NewIntention(repo, workspace).GetIntention("ws-1", "i-1")
		assert.ErrorIs(t, err, ErrIntentionNotFound)
	})

	t.Run("get returns an empty skill when none is set", func(t *testing.T) {
		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-1", Action: "deny"}}

		response, err := NewIntention(repo, workspace).GetIntention("ws-1", "i-1")
		require.NoError(t, err)
		assert.Equal(t, "", response.Skill)
		assert.Equal(t, "deny", response.Action)
	})

	t.Run("create rejects actions other than allow or deny", func(t *testing.T) {
		_, err := NewIntention(&fakeIntentions{}, workspace).CreateIntention("ws-1", saveIntentionRequest("maybe"))
		assert.ErrorIs(t, err, ErrInvalidIntentionAction)
	})

	t.Run("create returns not found for a missing workspace", func(t *testing.T) {
		_, err := NewIntention(&fakeIntentions{}, &fakeWorkspaces{}).CreateIntention("ws-1", saveIntentionRequest("allow"))
		assert.ErrorIs(t, err, ErrWorkspaceNotFound)
	})

	t.Run("create stores the rule with its skill", func(t *testing.T) {
		repo := &fakeIntentions{}

		response, err := NewIntention(repo, workspace).CreateIntention("ws-1", saveIntentionRequest("allow"))
		require.NoError(t, err)
		require.Len(t, repo.created, 1)
		assert.Equal(t, db.Id("ws-1"), repo.created[0].WorkspaceId)
		assert.Equal(t, "invoice.extract", response.Skill)
		assert.Equal(t, "allow", response.Action)
	})

	t.Run("create wraps the repository error", func(t *testing.T) {
		_, err := NewIntention(&fakeIntentions{err: errors.New("boom")}, workspace).CreateIntention("ws-1", saveIntentionRequest("allow"))
		assert.ErrorIs(t, err, ErrFailedCreateIntention)
	})

	t.Run("update rejects invalid actions and rules from another workspace", func(t *testing.T) {
		_, err := NewIntention(&fakeIntentions{}, workspace).UpdateIntention("ws-1", "i-1", saveIntentionRequest("maybe"))
		assert.ErrorIs(t, err, ErrInvalidIntentionAction)

		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-2"}}
		_, err = NewIntention(repo, workspace).UpdateIntention("ws-1", "i-1", saveIntentionRequest("allow"))
		assert.ErrorIs(t, err, ErrIntentionNotFound)
	})

	t.Run("update changes the rule in place", func(t *testing.T) {
		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-1", Action: "allow"}}

		response, err := NewIntention(repo, workspace).UpdateIntention("ws-1", "i-1", saveIntentionRequest("deny"))
		require.NoError(t, err)
		require.Len(t, repo.updated, 1)
		assert.Equal(t, "deny", repo.updated[0].Action)
		assert.Equal(t, "planner", repo.updated[0].SourceAgent)
		assert.Equal(t, "deny", response.Action)
	})

	t.Run("delete returns not found for a rule from another workspace", func(t *testing.T) {
		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-2"}}

		err := NewIntention(repo, workspace).DeleteIntention("ws-1", "i-1")
		assert.ErrorIs(t, err, ErrIntentionNotFound)
	})

	t.Run("delete removes the rule", func(t *testing.T) {
		repo := &fakeIntentions{item: &db.Intention{Id: "i-1", WorkspaceId: "ws-1"}}

		require.NoError(t, NewIntention(repo, workspace).DeleteIntention("ws-1", "i-1"))
		assert.Equal(t, []db.Id{"i-1"}, repo.deleted)
	})

	t.Run("check denies when no rule matches", func(t *testing.T) {
		response, err := NewIntention(&fakeIntentions{}, workspace).CheckIntention("ws-1", &CheckIntentionRequest{
			Source:      IntentionSource{Agent: "planner"},
			Destination: IntentionDestination{Agent: "invoice"},
		})
		require.NoError(t, err)
		assert.Equal(t, db.IntentionActionDeny, response.Action)
		assert.Nil(t, response.Intention)
	})

	t.Run("check returns the matching rule", func(t *testing.T) {
		repo := &fakeIntentions{match: &db.Intention{Id: "i-1", WorkspaceId: "ws-1", SourceAgent: "planner", DestinationAgent: "invoice", Action: "allow"}}

		response, err := NewIntention(repo, workspace).CheckIntention("ws-1", &CheckIntentionRequest{
			Source:      IntentionSource{Agent: "planner"},
			Destination: IntentionDestination{Agent: "invoice"},
		})
		require.NoError(t, err)
		assert.Equal(t, "allow", response.Action)
		require.NotNil(t, response.Intention)
		assert.Equal(t, db.Id("i-1"), response.Intention.Id)
	})

	t.Run("check wraps the match error", func(t *testing.T) {
		_, err := NewIntention(&fakeIntentions{err: errors.New("boom")}, workspace).CheckIntention("ws-1", &CheckIntentionRequest{})
		assert.ErrorIs(t, err, ErrFailedCheckIntention)
	})
}
