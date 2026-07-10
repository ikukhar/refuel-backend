CREATE TABLE IF NOT EXISTS activities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    distance DOUBLE PRECISION DEFAULT NULL,
    duration INTEGER DEFAULT NULL,
    elevation DOUBLE PRECISION DEFAULT NULL,
    calories DOUBLE PRECISION DEFAULT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'manual',
    source_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_id)
);

CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_activities_started_at ON activities(started_at);
