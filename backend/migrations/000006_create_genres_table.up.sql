CREATE TABLE IF NOT EXISTS genres (
    id integer PRIMARY KEY,
    name text NOT NULL
)

CREATE TABLE IF NOT EXISTS movies_genres (
    movie_id bigint NOT NULL REFERENCES movies ON DELETE CASCADE,
    genre_id integer NOT NULL REFERENCES genres ON DELETE CASCADE,
    PRIMARY KEY (movie_id, genre_id)
);