CREATE TABLE IF NOT EXISTS access_keys (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	name VARCHAR(60) NOT NULL,
	token VARCHAR(100) NOT NULL UNIQUE,
	expires_at TIMESTAMP NULL,
	meta JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_access_keys_workspace_id ON access_keys(workspace_id);
CREATE INDEX IF NOT EXISTS idx_access_keys_token ON access_keys(token);
