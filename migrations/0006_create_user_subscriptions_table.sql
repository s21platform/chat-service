-- +goose Up
CREATE TABLE IF NOT EXISTS user_subscriptions
(
    user_id       UUID NOT NULL,
    channel       TEXT NOT NULL,
    subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, channel),
    FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +goose Down
DROP TABLE IF EXISTS user_subscriptions;
