// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"testing"

	"github.com/clivern/sloth/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnitCatalog(t *testing.T) {
	caller := &db.Agent{Id: "agent-0", Name: "support-assistant"}

	t.Run("list skills passes the caller and filters to the repository", func(t *testing.T) {
		skills := &fakeAgentSkills{}

		_, err := NewCatalog(skills).ListSkills("ws-1", caller, &ListCatalogSkillsRequest{
			Tags:    []string{"profile"},
			Passing: true,
			Limit:   20,
			Offset:  40,
		})
		require.NoError(t, err)
		assert.Equal(t, "support-assistant", skills.filter.Caller)
		assert.Equal(t, []string{"profile"}, skills.filter.Tags)
		assert.True(t, skills.filter.Passing)
		assert.Equal(t, 20, skills.filter.Limit)
		assert.Equal(t, 40, skills.filter.Offset)
	})

	t.Run("list skills sends an empty tag list when no tags are given", func(t *testing.T) {
		skills := &fakeAgentSkills{}

		_, err := NewCatalog(skills).ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, []string{}, skills.filter.Tags)
	})

	t.Run("list skills maps rows and keeps the total", func(t *testing.T) {
		skills := &fakeAgentSkills{
			skills: []*db.CallableSkill{
				{SkillId: "user.profile.get", Name: "Get user profile", Tags: []string{"profile"}, InputModes: []string{"application/json"}, Agents: 2},
			},
			total: 7,
		}

		result, err := NewCatalog(skills).ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(7), result.Total)
		require.Len(t, result.Skills, 1)
		assert.Equal(t, "user.profile.get", result.Skills[0].Id)
		assert.Equal(t, 2, result.Skills[0].Agents)
		assert.Equal(t, []string{"application/json"}, result.Skills[0].InputModes)
		assert.Equal(t, []string{}, result.Skills[0].OutputModes)
	})

	t.Run("list skills returns repository errors", func(t *testing.T) {
		_, err := NewCatalog(&fakeAgentSkills{err: errors.New("db down")}).ListSkills("ws-1", caller, &ListCatalogSkillsRequest{Limit: 20})
		assert.Error(t, err)
	})
}
