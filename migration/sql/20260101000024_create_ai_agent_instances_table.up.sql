CREATE TABLE IF NOT EXISTS ai_agent_instances (
	id UUID PRIMARY KEY,
	agent_id UUID NOT NULL REFERENCES ai_agents(id) ON DELETE CASCADE,
	instance_id VARCHAR(100) NOT NULL,
	address VARCHAR(255) NOT NULL,
	port INTEGER NOT NULL,
	-- how Sloth reaches the instance's A2A endpoint: scheme://address:port/path
	scheme VARCHAR(10) NOT NULL DEFAULT 'http',
	path VARCHAR(255) NOT NULL DEFAULT '',
	datacenter VARCHAR(60),
	meta JSONB,
	status VARCHAR(20) NOT NULL DEFAULT 'passing',
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (agent_id, instance_id)
);
