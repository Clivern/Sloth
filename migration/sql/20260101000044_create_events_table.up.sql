CREATE TABLE IF NOT EXISTS events (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	type VARCHAR(60) NOT NULL,
	actor_id UUID NULL,
	actor_name VARCHAR(100) NULL,
	actor_type VARCHAR(20) NULL,
	resource_type VARCHAR(100) NULL,
	resource_id UUID NULL,
	ip_address VARCHAR(45) NULL,
	user_agent VARCHAR(200) NULL,
	data JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_events_workspace_id_created_at ON events(workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_resource ON events(resource_type, resource_id);
