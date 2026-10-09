CREATE TABLE a2a_tasks (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	source_agent VARCHAR(100) NOT NULL,
	destination_agent VARCHAR(100) NOT NULL,
	skill VARCHAR(120),
	state VARCHAR(20) NOT NULL DEFAULT 'submitted',
	request JSONB NOT NULL,
	response JSONB,
	error JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX idx_a2a_tasks_workspace_id_created_at ON a2a_tasks(workspace_id, created_at DESC);
CREATE INDEX idx_a2a_tasks_state_updated_at ON a2a_tasks(state, updated_at);
