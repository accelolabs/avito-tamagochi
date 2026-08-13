-- +goose Up

CREATE TABLE energy_notification_deliveries (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    threshold INTEGER NOT NULL CHECK (threshold IN (0, 5, 25, 50)),
    sent_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, threshold)
);

-- +goose Down

DROP TABLE energy_notification_deliveries;
