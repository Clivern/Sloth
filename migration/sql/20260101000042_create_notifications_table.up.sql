CREATE TABLE IF NOT EXISTS notifications (
	id UUID PRIMARY KEY,
	audience VARCHAR(20) NOT NULL,
	workspace_id UUID NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
	type VARCHAR(60) NOT NULL,
	severity VARCHAR(20) NOT NULL DEFAULT 'info',
	title VARCHAR(200) NOT NULL,
	message TEXT,
	resource_type VARCHAR(60) NULL,
	resource_id UUID NULL,
	data JSONB,
	checksum VARCHAR(64) NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	CONSTRAINT notifications_audience_check CHECK (
		(audience = 'user' AND user_id IS NOT NULL) OR
		(audience = 'workspace' AND user_id IS NULL AND workspace_id IS NOT NULL) OR
		(audience = 'system' AND user_id IS NULL AND workspace_id IS NULL)
	)
);
CREATE INDEX IF NOT EXISTS idx_notifications_workspace_id_created_at ON notifications(workspace_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_audience_created_at ON notifications(audience, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type_resource_created_at ON notifications(type, resource_type, resource_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_resource ON notifications(resource_type, resource_id);
