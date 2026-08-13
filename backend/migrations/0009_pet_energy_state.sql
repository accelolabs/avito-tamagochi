-- +goose Up

ALTER TABLE pets
    ADD COLUMN energy_percent INTEGER NOT NULL DEFAULT 100 CHECK (energy_percent BETWEEN 0 AND 100),
    ADD COLUMN energy_updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE pets
SET energy_percent = GREATEST(0, LEAST(100,
        100 - FLOOR(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - last_charged_at)) / 172800 * 100)::INTEGER
    )),
    energy_updated_at = CURRENT_TIMESTAMP;

-- +goose Down

ALTER TABLE pets
    DROP COLUMN energy_updated_at,
    DROP COLUMN energy_percent;
