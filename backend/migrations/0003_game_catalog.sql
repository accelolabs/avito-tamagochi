-- +goose Up

CREATE TABLE task_definitions (
    type VARCHAR(64) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    required_count INTEGER NOT NULL CHECK (required_count > 0),
    xp_reward INTEGER NOT NULL CHECK (xp_reward > 0)
);

CREATE TABLE task_rotation (
    cycle_day INTEGER NOT NULL CHECK (cycle_day BETWEEN 1 AND 3),
    task_type VARCHAR(64) NOT NULL REFERENCES task_definitions(type),
    PRIMARY KEY (cycle_day, task_type)
);

CREATE TABLE task_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    local_date DATE NOT NULL,
    task_type VARCHAR(64) NOT NULL REFERENCES task_definitions(type),
    progress INTEGER NOT NULL DEFAULT 0 CHECK (progress >= 0),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, local_date, task_type)
);

CREATE TABLE reward_definitions (
    id UUID PRIMARY KEY,
    level INTEGER NOT NULL UNIQUE CHECK (level >= 2),
    type VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL
);

CREATE TABLE user_rewards (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_id UUID NOT NULL REFERENCES reward_definitions(id),
    status VARCHAR(32) NOT NULL CHECK (status IN ('available', 'used')),
    unlocked_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    UNIQUE (user_id, reward_id)
);

INSERT INTO task_definitions (type, title, required_count, xp_reward) VALUES
    ('view', 'Посмотреть объявления', 5, 20),
    ('favorite', 'Добавить объявление в избранное', 1, 25),
    ('create_listing', 'Разместить объявление', 1, 40),
    ('create_listing_in_category', 'Разместить объявление в категории', 1, 50);

INSERT INTO task_rotation (cycle_day, task_type) VALUES
    (1, 'view'), (1, 'favorite'), (1, 'create_listing'),
    (2, 'view'), (2, 'create_listing'), (2, 'create_listing_in_category'),
    (3, 'favorite'), (3, 'create_listing'), (3, 'create_listing_in_category');

INSERT INTO reward_definitions (id, level, type, title) VALUES
    ('00000000-0000-0000-0000-000000000002', 2, 'promotion', 'Продвижение объявления'),
    ('00000000-0000-0000-0000-000000000003', 3, 'free_delivery', 'Бесплатная доставка'),
    ('00000000-0000-0000-0000-000000000004', 4, 'promotion', 'Продвижение объявления'),
    ('00000000-0000-0000-0000-000000000005', 5, 'free_delivery', 'Бесплатная доставка');

-- +goose Down

DROP TABLE user_rewards;
DROP TABLE reward_definitions;
DROP TABLE task_progress;
DROP TABLE task_rotation;
DROP TABLE task_definitions;
