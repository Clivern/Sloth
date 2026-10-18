CREATE TABLE IF NOT EXISTS workspace_token_purchases (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	stripe_session_id VARCHAR(200) NOT NULL UNIQUE,
	amount_cents BIGINT NOT NULL,
	tokens BIGINT NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_workspace_token_purchases_workspace_id ON workspace_token_purchases(workspace_id);
