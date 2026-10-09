CREATE TABLE health_checks (
	id UUID PRIMARY KEY,
	agent_instance_id UUID NOT NULL REFERENCES agent_instances(id) ON DELETE CASCADE,
	check_id VARCHAR(100) NOT NULL,
	name VARCHAR(120) NOT NULL,
	type VARCHAR(20) NOT NULL,
	-- lease: the instance lease, agent: copied from an agent check template
	source VARCHAR(20) NOT NULL DEFAULT 'agent',
	status VARCHAR(20) NOT NULL DEFAULT 'critical',
	definition JSONB,
	output TEXT,
	ttl_expires_at TIMESTAMP,
	last_run_at TIMESTAMP,
	-- when the runner should next probe an http or tcp check
	next_run_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	updated_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	UNIQUE (agent_instance_id, check_id)
);
CREATE INDEX idx_health_checks_agent_instance_id ON health_checks(agent_instance_id);
CREATE INDEX idx_health_checks_ttl_expires_at ON health_checks(ttl_expires_at)
	WHERE type = 'ttl';
CREATE INDEX idx_health_checks_next_run_at ON health_checks(next_run_at)
	WHERE next_run_at IS NOT NULL;
