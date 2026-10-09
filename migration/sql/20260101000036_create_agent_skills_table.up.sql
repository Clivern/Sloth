CREATE TABLE agent_skills (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
	skill_id VARCHAR(120) NOT NULL,
	name VARCHAR(120) NOT NULL,
	description TEXT,
	tags JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (agent_id, skill_id)
);
CREATE INDEX idx_agent_skills_workspace_id_skill_id ON agent_skills(workspace_id, skill_id);
