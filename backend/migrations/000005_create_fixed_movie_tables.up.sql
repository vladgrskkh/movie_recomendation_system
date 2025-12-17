CREATE TABLE IF NOT EXISTS popular_movies (
    id bigserial PRIMARY KEY REFERENCES movies ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS recommended_movies (
    id bigserial PRIMARY KEY REFERENCES movies ON DELETE CASCADE
);