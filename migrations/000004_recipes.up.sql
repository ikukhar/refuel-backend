CREATE TABLE IF NOT EXISTS recipes (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    calories INTEGER NOT NULL,
    protein_g DOUBLE PRECISION NOT NULL,
    fat_g DOUBLE PRECISION NOT NULL,
    carbs_g DOUBLE PRECISION NOT NULL,
    image_url TEXT DEFAULT NULL,
    meal_type VARCHAR(50) NOT NULL DEFAULT 'other',
    steps TEXT DEFAULT '[]',
    ingredients TEXT DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
