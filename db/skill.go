// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package db

import (
	"database/sql"
	"encoding/json"
	"errors"
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

// ListCallableAgentsFilter narrows and pages the agents a caller may call for one skill.
type ListCallableAgentsFilter struct {
	Caller  string
	Passing bool
	Limit   int
	Offset  int
}

// AgentSkill is one skill as described on an agent card.
type AgentSkill struct {
	SkillId     string
	Name        string
	Description string
	Tags        []string
}

// CallableAgent is one agent a caller may call for a skill, with its instance counts.
type CallableAgent struct {
	Name             string
	Description      string
	Version          string
	Instances        int
	PassingInstances int
}

// AgentSkillRepository is the interface for the skills flattened from agent cards.
type AgentSkillRepository interface {
	ListCallable(workspaceId Id, filter *ListCallableSkillsFilter) ([]*CallableSkill, int64, error)
	GetSkill(workspaceId Id, skillId string) (*AgentSkill, error)
	ListCallableAgents(workspaceId Id, skillId string, filter *ListCallableAgentsFilter) ([]*CallableAgent, int64, error)
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

// GetSkill returns a skill as described by the first agent by name that offers it, or nil when none does.
func (r *AgentSkillRepositoryPostgres) GetSkill(workspaceId Id, skillId string) (*AgentSkill, error) {
	skill := &AgentSkill{}
	var tags []byte

	err := r.db.QueryRow(
		`SELECT s.skill_id, s.name, COALESCE(s.description, ''), COALESCE(s.tags, '[]'::jsonb)
		FROM ai_agent_skills s
		JOIN ai_agents a ON a.id = s.agent_id
		WHERE s.workspace_id = $1 AND s.skill_id = $2
		ORDER BY a.name
		LIMIT 1`,
		workspaceId.String(),
		skillId,
	).Scan(&skill.SkillId, &skill.Name, &skill.Description, &tags)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(tags, &skill.Tags)

	return skill, nil
}

// ListCallableAgents returns one page of the agents the caller may call for a skill, sorted by name, and the total.
func (r *AgentSkillRepositoryPostgres) ListCallableAgents(workspaceId Id, skillId string, filter *ListCallableAgentsFilter) ([]*CallableAgent, int64, error) {
	rows, err := r.db.Query(
		`WITH callable AS (
			SELECT
				a.name,
				COALESCE(a.card->>'description', '') AS description,
				a.version,
				(SELECT count(*) FROM ai_agent_instances i WHERE i.agent_id = a.id) AS instances,
				(
					SELECT count(*) FROM ai_agent_instances i
					WHERE i.agent_id = a.id
						AND NOT EXISTS (
							SELECT 1 FROM ai_health_checks hc
							WHERE hc.instance_id = i.id
								AND (
									hc.status IN ($6, $7)
									OR (hc.type = $8 AND (hc.ttl_expires_at IS NULL OR hc.ttl_expires_at <= $5))
								)
						)
				) AS passing_instances
			FROM ai_agent_skills s
			JOIN ai_agents a ON a.id = s.agent_id
			WHERE s.workspace_id = $1
				AND s.skill_id = $3
				AND a.name <> $2
				AND (
					SELECT it.action FROM ai_intentions it
					WHERE it.workspace_id = $1
						AND it.source_agent = $2
						AND it.destination_agent = a.name
						AND (it.skill IS NULL OR it.skill = s.skill_id)
					ORDER BY it.skill IS NULL
					LIMIT 1
				) = $9
		),
		filtered AS (
			SELECT * FROM callable WHERE NOT $4::boolean OR passing_instances > 0
		),
		page AS (
			SELECT * FROM filtered ORDER BY name LIMIT $10 OFFSET $11
		)
		SELECT t.total, p.name, p.description, p.version, p.instances, p.passing_instances
		FROM (SELECT count(*) AS total FROM filtered) t
		LEFT JOIN page p ON true
		ORDER BY p.name`,
		workspaceId.String(),
		filter.Caller,
		skillId,
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
	agents := []*CallableAgent{}
	for rows.Next() {
		var name, description, version sql.NullString
		var instances, passing sql.NullInt64

		err := rows.Scan(&total, &name, &description, &version, &instances, &passing)
		if err != nil {
			return nil, 0, err
		}
		if !name.Valid {
			continue
		}

		agents = append(agents, &CallableAgent{
			Name:             name.String,
			Description:      description.String,
			Version:          version.String,
			Instances:        int(instances.Int64),
			PassingInstances: int(passing.Int64),
		})
	}

	return agents, total, rows.Err()
}
