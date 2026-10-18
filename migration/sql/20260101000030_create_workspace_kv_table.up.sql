CREATE TABLE IF NOT EXISTS workspace_kv (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	key VARCHAR(200) NOT NULL,
	value TEXT NOT NULL,
	expires_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (workspace_id, key)
);
CREATE INDEX IF NOT EXISTS idx_workspace_kv_expires_at ON workspace_kv(expires_at);
