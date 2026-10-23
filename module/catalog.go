// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
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
