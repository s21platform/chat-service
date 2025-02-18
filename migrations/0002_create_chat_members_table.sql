-- +goose Up
CREATE TYPE role_type AS ENUM ('member', 'admin');

CREATE TABLE IF NOT EXISTS chat_members
(
    id        SERIAL PRIMARY KEY,
    chat_id   INT REFERENCES chats(id) ON DELETE CASCADE,
    user_uuid UUID      NOT NULL,
    role      role_type DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS chat_members;
DROP TYPE IF EXISTS role_type;
