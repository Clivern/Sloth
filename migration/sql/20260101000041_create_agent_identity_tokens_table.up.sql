CREATE TABLE agent_identity_tokens (
	id UUID PRIMARY KEY,
	identity_id UUID NOT NULL REFERENCES agent_identity(id) ON DELETE CASCADE,
	token_hash VARCHAR(64) NOT NULL UNIQUE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX idx_agent_identity_tokens_identity_id ON agent_identity_tokens(identity_id);
CREATE INDEX idx_agent_identity_tokens_expires_at ON agent_identity_tokens(expires_at);
