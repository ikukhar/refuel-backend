CREATE TABLE IF NOT EXISTS daily_nutrition (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    calories_target DOUBLE PRECISION NOT NULL,
    protein_g DOUBLE PRECISION NOT NULL,
    fat_g DOUBLE PRECISION NOT NULL,
    carbs_g DOUBLE PRECISION NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'baseline',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX idx_nutrition_user_date ON daily_nutrition(user_id, date);
