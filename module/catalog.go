// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/clivern/sloth/db"

	"github.com/samber/lo"
)

// ListCatalogSkillsRequest filters and pages the skills a caller may call.
type ListCatalogSkillsRequest struct {
	Tags    []string
	Passing bool
	Limit   int
	Offset  int
}

// CatalogSkill is one skill the caller may call, with the number of agents that serve it.
type CatalogSkill struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	InputModes  []string `json:"inputModes"`
	OutputModes []string `json:"outputModes"`
	Agents      int      `json:"agents"`
}

// ListCatalogSkillsResponse is one page of callable skills and the total across all pages.
type ListCatalogSkillsResponse struct {
	Skills []*CatalogSkill
	Total  int64
}

// catalogCard is the part of an agent card the catalog reads.
type catalogCard struct {
	DefaultInputModes  []string           `json:"defaultInputModes"`
	DefaultOutputModes []string           `json:"defaultOutputModes"`
	Skills             []catalogCardSkill `json:"skills"`
}

// catalogCardSkill is one skill on an agent card.
type catalogCardSkill struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	InputModes  []string `json:"inputModes"`
	OutputModes []string `json:"outputModes"`
}

// Catalog is the module for the skills and agents a calling agent may call.
type Catalog struct {
	AgentRepository     db.AgentRepository
	InstanceRepository  db.AgentInstanceRepository
	IntentionRepository db.IntentionRepository
}

// NewCatalog creates a catalog module with the given repositories.
func NewCatalog(agents db.AgentRepository, instances db.AgentInstanceRepository, intentions db.IntentionRepository) *Catalog {
	return &Catalog{
		AgentRepository:     agents,
		InstanceRepository:  instances,
		IntentionRepository: intentions,
	}
}

// ListSkills returns the skills on workspace agent cards that the caller may call, grouped by skill id.
func (c *Catalog) ListSkills(workspaceId db.Id, caller *db.Agent, req *ListCatalogSkillsRequest) (*ListCatalogSkillsResponse, error) {
	agents, err := c.agents(workspaceId, caller, req.Passing)
	if err != nil {
		return nil, err
	}

	byId := map[string]*CatalogSkill{}
	for _, agent := range agents {
		var card catalogCard
		_ = json.Unmarshal([]byte(lo.FromPtr(agent.Card)), &card)

		for _, skill := range card.Skills {
			if len(req.Tags) > 0 && !lo.Some(skill.Tags, req.Tags) {
				continue
			}

			allowed, err := c.allowed(workspaceId, caller.Name, agent.Name, skill.Id)
			if err != nil {
				return nil, err
			}
			if !allowed {
				continue
			}

			// Agents are sorted by name, so the first card to list a skill describes it.
			entry, ok := byId[skill.Id]
			if !ok {
				entry = &CatalogSkill{
					Id:          skill.Id,
					Name:        skill.Name,
					Description: skill.Description,
					Tags:        lo.CoalesceSliceOrEmpty(skill.Tags, []string{}),
					InputModes:  lo.CoalesceSliceOrEmpty(skill.InputModes, card.DefaultInputModes, []string{}),
					OutputModes: lo.CoalesceSliceOrEmpty(skill.OutputModes, card.DefaultOutputModes, []string{}),
				}
				byId[skill.Id] = entry
			}

			entry.Agents++
		}
	}

	skills := lo.Values(byId)
	slices.SortFunc(skills, func(a, b *CatalogSkill) int { return strings.Compare(a.Id, b.Id) })

	start := min(req.Offset, len(skills))
	end := min(start+req.Limit, len(skills))

	return &ListCatalogSkillsResponse{
		Skills: skills[start:end],
		Total:  int64(len(skills)),
	}, nil
}

// agents returns the workspace agents with a card other than the caller, sorted by name, optionally only those with a live instance.
func (c *Catalog) agents(workspaceId db.Id, caller *db.Agent, passing bool) ([]*db.Agent, error) {
	total, err := c.AgentRepository.CountByWorkspaceId(workspaceId)
	if err != nil {
		return nil, err
	}

	agents, err := c.AgentRepository.ListByWorkspaceId(workspaceId, int(total), 0)
	if err != nil {
		return nil, err
	}

	agents = lo.Filter(agents, func(agent *db.Agent, _ int) bool {
		return agent.Id != caller.Id && agent.Card != nil
	})
	slices.SortFunc(agents, func(a, b *db.Agent) int { return strings.Compare(a.Name, b.Name) })

	if !passing {
		return agents, nil
	}

	now := time.Now().UTC()

	var live []*db.Agent
	for _, agent := range agents {
		count, err := c.InstanceRepository.CountLiveByAgentId(agent.Id, now)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			live = append(live, agent)
		}
	}

	return live, nil
}

// allowed reports whether an intention lets source call a skill on destination; no intention means deny.
func (c *Catalog) allowed(workspaceId db.Id, source, destination, skill string) (bool, error) {
	intention, err := c.IntentionRepository.Match(workspaceId, source, destination, skill)
	if err != nil {
		return false, err
	}

	return intention != nil && intention.Action == db.IntentionActionAllow, nil
}
