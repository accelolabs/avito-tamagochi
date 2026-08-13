-- +goose Up

ALTER TABLE pets
    ADD COLUMN longest_streak INTEGER NOT NULL DEFAULT 0 CHECK (longest_streak >= 0),
    ADD COLUMN streak_started_date DATE;

UPDATE pets
SET longest_streak = charge_streak,
    streak_started_date = CASE
        WHEN charge_streak > 0 AND last_streak_date IS NOT NULL
            THEN last_streak_date - (charge_streak - 1)
        ELSE NULL
    END;

-- +goose Down

ALTER TABLE pets
    DROP COLUMN streak_started_date,
    DROP COLUMN longest_streak;
