CREATE TABLE IF NOT EXISTS ai_agent_protection_tokens (
	id UUID PRIMARY KEY,
	protection_id UUID NOT NULL REFERENCES ai_agent_protection(id) ON DELETE CASCADE,
	token_hash VARCHAR(128) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	revoked_at TIMESTAMP NULL,
	meta JSONB,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS idx_ai_agent_protection_tokens_protection_id ON ai_agent_protection_tokens(protection_id);
CREATE INDEX IF NOT EXISTS idx_ai_agent_protection_tokens_expires_at ON ai_agent_protection_tokens(expires_at);
