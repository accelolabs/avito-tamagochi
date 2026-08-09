-- +goose Up

ALTER TABLE task_progress
    ADD CONSTRAINT task_progress_required_count_check
    CHECK (
        (task_type = 'view' AND progress <= 5)
        OR (task_type IN ('favorite', 'create_listing', 'create_listing_in_category') AND progress <= 1)
    );

-- +goose Down

ALTER TABLE task_progress
    DROP CONSTRAINT task_progress_required_count_check;
