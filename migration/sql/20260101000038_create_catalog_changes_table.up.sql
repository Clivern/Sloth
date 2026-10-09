CREATE TABLE catalog_changes (
	id BIGSERIAL PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	agent_name VARCHAR(100) NOT NULL,
	kind VARCHAR(20) NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX idx_catalog_changes_workspace_id_id ON catalog_changes(workspace_id, id);
