CREATE TABLE IF NOT EXISTS ai_intentions (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	source_agent VARCHAR(100) NOT NULL,
	destination_agent VARCHAR(100) NOT NULL,
	skill VARCHAR(120),
	action VARCHAR(20) NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_intentions_workspace_id_source_destination
	ON ai_intentions (workspace_id, source_agent, destination_agent);
