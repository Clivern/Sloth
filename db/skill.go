// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ListCallableSkillsFilter narrows and pages the skills a caller may call.
type ListCallableSkillsFilter struct {
	Caller  string
	Tags    []string
	Passing bool
	Limit   int
	Offset  int
}

// CallableSkill is one skill a caller may call, grouped across the agents that offer it.
type CallableSkill struct {
	SkillId     string
	Name        string
	Description string
	Tags        []string
	InputModes  []string
	OutputModes []string
	Agents      int
}

// AgentSkillRepository is the interface for the skills flattened from agent cards.
type AgentSkillRepository interface {
	ListCallable(workspaceId Id, filter *ListCallableSkillsFilter) ([]*CallableSkill, int64, error)
}

type AgentSkillRepositoryPostgres struct {
	db *sql.DB
}

// NewAgentSkillRepository returns the repository for the ai_agent_skills table.
func NewAgentSkillRepository(db *sql.DB) AgentSkillRepository {
	return &AgentSkillRepositoryPostgres{db: db}
}

// ListCallable returns one page of the workspace skills the caller may call and the total across all pages.
func (r *AgentSkillRepositoryPostgres) ListCallable(workspaceId Id, filter *ListCallableSkillsFilter) ([]*CallableSkill, int64, error) {
	rows, err := r.db.Query(
		`WITH grouped AS (
			SELECT
				s.skill_id,
				(array_agg(s.name ORDER BY a.name))[1] AS name,
				(array_agg(COALESCE(s.description, '') ORDER BY a.name))[1] AS description,
				(array_agg(COALESCE(s.tags, '[]'::jsonb) ORDER BY a.name))[1] AS tags,
				(array_agg(s.input_modes ORDER BY a.name))[1] AS input_modes,
				(array_agg(s.output_modes ORDER BY a.name))[1] AS output_modes,
				count(*) AS agents
			FROM ai_agent_skills s
			JOIN ai_agents a ON a.id = s.agent_id
			WHERE s.workspace_id = $1
				AND a.name <> $2
				AND (cardinality($3::text[]) = 0 OR s.tags ?| $3::text[])
				AND (NOT $4::boolean OR EXISTS (
					SELECT 1 FROM ai_agent_instances i
					WHERE i.agent_id = a.id
						AND NOT EXISTS (
							SELECT 1 FROM ai_health_checks hc
							WHERE hc.instance_id = i.id
								AND (
									hc.status IN ($6, $7)
									OR (hc.type = $8 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $5))
								)
						)
				))
				AND (
					SELECT it.action FROM ai_intentions it
					WHERE it.workspace_id = $1
						AND it.source_agent = $2
						AND it.destination_agent = a.name
						AND (it.skill IS NULL OR it.skill = s.skill_id)
					ORDER BY it.skill IS NULL
					LIMIT 1
				) = $9
			GROUP BY s.skill_id
		),
		page AS (
			SELECT * FROM grouped ORDER BY skill_id LIMIT $10 OFFSET $11
		)
		SELECT t.total, p.skill_id, p.name, p.description, p.tags, p.input_modes, p.output_modes, p.agents
		FROM (SELECT count(*) AS total FROM grouped) t
		LEFT JOIN page p ON true
		ORDER BY p.skill_id`,
		workspaceId.String(),
		filter.Caller,
		filter.Tags,
		filter.Passing,
		time.Now().UTC(),
		HealthCheckStatusCritical,
		HealthCheckStatusUnknown,
		HealthCheckTypeTTL,
		IntentionActionAllow,
		filter.Limit,
		filter.Offset,
	)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	// The total row is always returned; page columns are NULL when the page is empty.
	var total int64
	skills := []*CallableSkill{}
	for rows.Next() {
		var skillId, name, description sql.NullString
		var tags, inputModes, outputModes []byte
		var agents sql.NullInt64

		err := rows.Scan(&total, &skillId, &name, &description, &tags, &inputModes, &outputModes, &agents)
		if err != nil {
			return nil, 0, err
		}
		if !skillId.Valid {
			continue
		}

		skill := &CallableSkill{
			SkillId:     skillId.String,
			Name:        name.String,
			Description: description.String,
			Agents:      int(agents.Int64),
		}
		_ = json.Unmarshal(tags, &skill.Tags)
		_ = json.Unmarshal(inputModes, &skill.InputModes)
		_ = json.Unmarshal(outputModes, &skill.OutputModes)

		skills = append(skills, skill)
	}

	return skills, total, rows.Err()
}
