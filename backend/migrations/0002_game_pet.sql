-- +goose Up

CREATE TABLE pets (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    xp INTEGER NOT NULL DEFAULT 0 CHECK (xp >= 0),
    last_charged_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE xp_events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    source VARCHAR(64) NOT NULL,
    source_key VARCHAR(128) NOT NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    occurred_at TIMESTAMPTZ NOT NULL,
    local_date DATE NOT NULL,
    UNIQUE (user_id, source_key)
);

CREATE INDEX xp_events_user_date_idx ON xp_events(user_id, local_date);
CREATE INDEX xp_events_pet_date_idx ON xp_events(pet_id, local_date);
CREATE INDEX xp_events_amount_idx ON xp_events(amount);

-- +goose Down

DROP TABLE xp_events;
DROP TABLE pets;
