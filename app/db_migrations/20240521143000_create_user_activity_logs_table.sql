CREATE TABLE user_activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    action_type INT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    entity_type VARCHAR(100),
    entity_id BIGINT,
    meta JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_activity_logs_user_id ON user_activity_logs(user_id);
CREATE INDEX idx_user_activity_logs_action_type ON user_activity_logs(action_type);
CREATE INDEX idx_user_activity_logs_public_feed ON user_activity_logs(created_at DESC) WHERE is_public = TRUE;
