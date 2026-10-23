CREATE TABLE IF NOT EXISTS ai_agents_tasks (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	-- the id the destination agent generated for the task
	task_id VARCHAR(100) NOT NULL,
	-- the calling agent; only it may read the task
	source_agent VARCHAR(100) NOT NULL,
	agent_id UUID NOT NULL REFERENCES ai_agents(id) ON DELETE CASCADE,
	-- the instance that accepted the task and holds its state
	instance_id UUID REFERENCES ai_agent_instances(id) ON DELETE SET NULL,
	skill VARCHAR(120) NOT NULL DEFAULT '',
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (workspace_id, source_agent, task_id)
);
CREATE INDEX IF NOT EXISTS idx_ai_agents_tasks_created_at ON ai_agents_tasks(created_at);
