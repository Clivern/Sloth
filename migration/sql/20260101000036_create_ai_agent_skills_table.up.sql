CREATE TABLE IF NOT EXISTS ai_agent_skills (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	agent_id UUID NOT NULL REFERENCES ai_agents(id) ON DELETE CASCADE,
	skill_id VARCHAR(120) NOT NULL,
	name VARCHAR(120) NOT NULL,
	description TEXT,
	tags JSONB,
	input_modes JSONB NOT NULL DEFAULT '[]'::jsonb,
	output_modes JSONB NOT NULL DEFAULT '[]'::jsonb,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (agent_id, skill_id)
);
CREATE INDEX IF NOT EXISTS idx_ai_agent_skills_workspace_id_skill_id ON ai_agent_skills(workspace_id, skill_id);
