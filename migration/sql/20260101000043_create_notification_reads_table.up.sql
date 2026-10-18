CREATE TABLE IF NOT EXISTS notification_reads (
	notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	read_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
	is_dismissed BOOLEAN NOT NULL DEFAULT FALSE,
	PRIMARY KEY (notification_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_notification_reads_user_id ON notification_reads(user_id);
