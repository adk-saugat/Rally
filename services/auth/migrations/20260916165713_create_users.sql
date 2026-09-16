-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT NOT NULL
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

-- +goose Down
DROP TABLE IF EXISTS users;
