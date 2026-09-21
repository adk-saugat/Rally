-- +goose Up
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    invite_code TEXT NOT NULL
);

CREATE UNIQUE INDEX groups_invite_code_key ON groups (invite_code);

CREATE TABLE group_members (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX group_members_user_id_idx ON group_members (user_id);

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
