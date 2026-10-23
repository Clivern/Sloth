// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"errors"
	"fmt"

	"github.com/clivern/sloth/db"

	"github.com/samber/lo"
	"github.com/spf13/viper"
)

var ErrSkillNotFound = errors.New("skill not found")

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

// ListCatalogSkillAgentsRequest filters and pages the agents a caller may call for one skill.
type ListCatalogSkillAgentsRequest struct {
	Passing bool
	Limit   int
	Offset  int
}

// CatalogSkillInfo describes the skill a list of agents is for.
type CatalogSkillInfo struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// CatalogAgentInstances counts an agent's instances and how many of them are live.
type CatalogAgentInstances struct {
	Total   int `json:"total"`
	Passing int `json:"passing"`
}

// CatalogAgent is one agent the caller may call for a skill, with gateway urls.
type CatalogAgent struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Version     string                `json:"version"`
	Status      string                `json:"status"`
	Instances   CatalogAgentInstances `json:"instances"`
	CardUrl     string                `json:"cardUrl"`
	SendUrl     string                `json:"sendUrl"`
}

// ListCatalogSkillAgentsResponse is the skill, one page of callable agents and the total across all pages.
type ListCatalogSkillAgentsResponse struct {
	Skill  *CatalogSkillInfo
	Agents []*CatalogAgent
	Total  int64
}

// Catalog is the module for the skills and agents a calling agent may call.
type Catalog struct {
	SkillRepository db.AgentSkillRepository
}

// NewCatalog creates a catalog module with the given repositories.
func NewCatalog(skills db.AgentSkillRepository) *Catalog {
	return &Catalog{
		SkillRepository: skills,
	}
}

// ListSkills returns one page of the workspace skills the caller may call, grouped by skill id.
func (c *Catalog) ListSkills(workspaceId db.Id, caller *db.Agent, req *ListCatalogSkillsRequest) (*ListCatalogSkillsResponse, error) {
	skills, total, err := c.SkillRepository.ListCallable(workspaceId, &db.ListCallableSkillsFilter{
		Caller:  caller.Name,
		Tags:    lo.CoalesceSliceOrEmpty(req.Tags, []string{}),
		Passing: req.Passing,
		Limit:   req.Limit,
		Offset:  req.Offset,
	})
	if err != nil {
		return nil, err
	}

	return &ListCatalogSkillsResponse{
		Skills: lo.Map(skills, func(skill *db.CallableSkill, _ int) *CatalogSkill {
			return &CatalogSkill{
				Id:          skill.SkillId,
				Name:        skill.Name,
				Description: skill.Description,
				Tags:        lo.CoalesceSliceOrEmpty(skill.Tags, []string{}),
				InputModes:  lo.CoalesceSliceOrEmpty(skill.InputModes, []string{}),
				OutputModes: lo.CoalesceSliceOrEmpty(skill.OutputModes, []string{}),
				Agents:      skill.Agents,
			}
		}),
		Total: total,
	}, nil
}

// ListSkillAgents returns the skill and one page of the agents the caller may call for it.
func (c *Catalog) ListSkillAgents(workspaceId db.Id, caller *db.Agent, skillId string, req *ListCatalogSkillAgentsRequest) (*ListCatalogSkillAgentsResponse, error) {
	skill, err := c.SkillRepository.GetSkill(workspaceId, skillId)
	if err != nil {
		return nil, err
	}
	if skill == nil {
		return nil, ErrSkillNotFound
	}

	agents, total, err := c.SkillRepository.ListCallableAgents(workspaceId, skillId, &db.ListCallableAgentsFilter{
		Caller:  caller.Name,
		Passing: req.Passing,
		Limit:   req.Limit,
		Offset:  req.Offset,
	})
	if err != nil {
		return nil, err
	}

	base := fmt.Sprintf("%s/api/v1/workspaces/%s/agents", viper.GetString("app.url"), workspaceId)

	return &ListCatalogSkillAgentsResponse{
		Skill: &CatalogSkillInfo{
			Id:          skill.SkillId,
			Name:        skill.Name,
			Description: skill.Description,
			Tags:        lo.CoalesceSliceOrEmpty(skill.Tags, []string{}),
		},
		Agents: lo.Map(agents, func(agent *db.CallableAgent, _ int) *CatalogAgent {
			return &CatalogAgent{
				Name:        agent.Name,
				Description: agent.Description,
				Version:     agent.Version,
				Status:      lo.Ternary(agent.PassingInstances > 0, db.HealthCheckStatusPassing, db.HealthCheckStatusCritical),
				Instances: CatalogAgentInstances{
					Total:   agent.Instances,
					Passing: agent.PassingInstances,
				},
				CardUrl: fmt.Sprintf("%s/%s/.well-known/agent-card.json", base, agent.Name),
				SendUrl: fmt.Sprintf("%s/%s/message:send", base, agent.Name),
			}
		}),
		Total: total,
	}, nil
}
