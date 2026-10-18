CREATE TABLE IF NOT EXISTS async_tasks (
	id UUID PRIMARY KEY,
	workspace_id UUID NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	type VARCHAR(60) NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	payload JSONB,
	result JSONB,
	error JSONB,
	attempts INTEGER NOT NULL DEFAULT 0,
	priority SMALLINT NOT NULL DEFAULT 50,
	run_at TIMESTAMP NULL,
	locked_at TIMESTAMP NULL,
	completed_at TIMESTAMP NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_async_tasks_status_priority_created_at
	ON async_tasks (status, priority DESC, created_at);
