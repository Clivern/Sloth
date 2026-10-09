CREATE TABLE agent_checks (
	id UUID PRIMARY KEY,
	agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
	check_id VARCHAR(100) NOT NULL,
	name VARCHAR(120) NOT NULL,
	type VARCHAR(20) NOT NULL,
	definition JSONB NOT NULL,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (agent_id, check_id)
);
CREATE INDEX idx_agent_checks_agent_id ON agent_checks(agent_id);
