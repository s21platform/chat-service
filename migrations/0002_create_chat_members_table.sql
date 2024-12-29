-- +goose Up
CREATE TYPE role_type AS ENUM ('member', 'admin');

CREATE TABLE IF NOT EXISTS chat_members
(
    id        SERIAL PRIMARY KEY,
    chat_id   INT       NOT NULL,
    user_uuid UUID      NOT NULL,
    role      role_type NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_chat_members_chats FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS chat_members;
DROP TYPE IF EXISTS role_type;
