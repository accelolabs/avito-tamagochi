-- +goose Up

UPDATE pets
SET longest_streak = GREATEST(longest_streak, charge_streak),
    streak_started_date = CASE
        WHEN charge_streak = 0 THEN NULL
        ELSE COALESCE(streak_started_date, last_streak_date - (charge_streak - 1))
    END;

ALTER TABLE pets
    ADD CONSTRAINT pets_longest_streak_check CHECK (longest_streak >= charge_streak),
    ADD CONSTRAINT pets_streak_dates_check CHECK (
        (charge_streak = 0 AND streak_started_date IS NULL)
        OR (charge_streak > 0 AND streak_started_date IS NOT NULL AND last_streak_date IS NOT NULL)
    );

-- +goose Down

ALTER TABLE pets
    DROP CONSTRAINT pets_streak_dates_check,
    DROP CONSTRAINT pets_longest_streak_check;
