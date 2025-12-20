CREATE TABLE IF NOT EXISTS watched_movies (
    id bigserial PRIMARY KEY REFERENCES movies ON DELETE CASCADE,
    user_id bigserial NOT NULL REFERENCES users ON DELETE CASCADE
);