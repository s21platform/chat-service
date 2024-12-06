-- +goose Up
CREATE TYPE role_type AS ENUM ('member', 'admin');

CREATE TABLE IF NOT EXISTS chat_members (
    id INT SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_uuid UUID NOT NULL,
    role role_type NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_chat_members_chats FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chat_members;
DROP TYPE role_type;
