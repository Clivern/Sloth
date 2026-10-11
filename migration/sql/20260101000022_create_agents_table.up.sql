CREATE TABLE agents (
	id UUID PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
	name VARCHAR(100) NOT NULL,
	card JSONB,
	card_checksum VARCHAR(64),
	card_path VARCHAR(255) NOT NULL DEFAULT '/.well-known/agent-card.json',
	version VARCHAR(40) NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (workspace_id, name)
);
