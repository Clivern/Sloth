CREATE TABLE IF NOT EXISTS ai_agent_identity_tokens (
	id UUID PRIMARY KEY,
	identity_id UUID NOT NULL REFERENCES ai_agent_identity(id) ON DELETE CASCADE,
	token_hash VARCHAR(64) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_agent_identity_tokens_identity_id ON ai_agent_identity_tokens(identity_id);
CREATE INDEX IF NOT EXISTS idx_ai_agent_identity_tokens_expires_at ON ai_agent_identity_tokens(expires_at);
