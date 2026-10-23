// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"testing"

	"github.com/clivern/sloth/db"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitCatalog(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant", Card: lo.ToPtr(`{"skills":[{"id":"ticket.open"}]}`)}
	directory := &db.Agent{
		Id:   "agent-1",
		Name: "user-directory",
		Card: lo.ToPtr(`{
			"defaultInputModes": ["application/json"],
			"defaultOutputModes": ["application/json"],
			"skills": [
				{"id": "user.profile.get", "name": "Get user profile", "tags": ["profile", "users"]},
				{"id": "user.delete", "name": "Delete user", "tags": ["users"]}
			]
		}`),
	}
	mirror := &db.Agent{
		Id:   "agent-2",
		Name: "user-mirror",
		Card: lo.ToPtr(`{"skills":[{"id":"user.profile.get","name":"Profile (mirror)","inputModes":["text/plain"]}]}`),
	}
	pending := &db.Agent{Id: "agent-3", Name: "pending"}

	newCatalog := func(rules map[string]string, live map[db.Id]int64) *Catalog {
		agents := &fakeAgents{list: []*db.Agent{mirror, pending, caller, directory}, total: 4}
		return NewCatalog(agents, &fakeCatalogInstances{live: live}, &fakeCatalogIntentions{rules: rules})
	}

	t.Run("list skills keeps only skills an allow intention covers", func(t *testing.T) {
		catalog := newCatalog(map[string]string{
			"support-assistant/user-directory/":            "allow",
			"support-assistant/user-directory/user.delete": "deny",
		}, nil)

		result, err := catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "user.profile.get", result.Skills[0].Id)
		assert.Equal(t, 1, result.Skills[0].Agents)
		assert.Equal(t, []string{"application/json"}, result.Skills[0].InputModes)
	})

	t.Run("list skills groups agents and describes a skill from the first agent by name", func(t *testing.T) {
		catalog := newCatalog(map[string]string{
			"support-assistant/user-directory/": "allow",
			"support-assistant/user-mirror/":    "allow",
		}, nil)

		result, err := catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, []string{"user.delete", "user.profile.get"}, lo.Map(result.Skills, func(s *CatalogSkill, _ int) string { return s.Id }))
		assert.Equal(t, 2, result.Skills[1].Agents)
		assert.Equal(t, "Get user profile", result.Skills[1].Name)
	})

	t.Run("list skills never lists the caller and returns nothing without intentions", func(t *testing.T) {
		catalog := newCatalog(map[string]string{"support-assistant/support-assistant/": "allow"}, nil)

		result, err := catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20})
		require.NoError(t, err)
		assert.Empty(t, result.Skills)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("list skills filters by tag and by live instances", func(t *testing.T) {
		catalog := newCatalog(map[string]string{
			"support-assistant/user-directory/": "allow",
			"support-assistant/user-mirror/":    "allow",
		}, map[db.Id]int64{"agent-2": 1})

		result, err := catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Passing: true, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "Profile (mirror)", result.Skills[0].Name)

		result, err = catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Tags: []string{"profile"}, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, "user.profile.get", result.Skills[0].Id)
	})

	t.Run("list skills pages after sorting", func(t *testing.T) {
		catalog := newCatalog(map[string]string{"support-assistant/user-directory/": "allow"}, nil)

		result, err := catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 1, Offset: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, "user.profile.get", result.Skills[0].Id)

		result, err = catalog.ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20, Offset: 5})
		require.NoError(t, err)
		assert.Empty(t, result.Skills)
	})
}
