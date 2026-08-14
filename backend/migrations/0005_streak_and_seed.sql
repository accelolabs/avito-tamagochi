-- +goose Up

ALTER TABLE pets
    ADD COLUMN charge_streak INTEGER NOT NULL DEFAULT 0 CHECK (charge_streak >= 0),
    ADD COLUMN last_streak_date DATE;

CREATE INDEX xp_events_occurred_at_idx ON xp_events(occurred_at);

-- +goose Down

DROP INDEX xp_events_occurred_at_idx;

ALTER TABLE pets
    DROP COLUMN last_streak_date,
    DROP COLUMN charge_streak;
