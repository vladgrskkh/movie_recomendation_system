CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS movies (
    id bigint PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    original_title text NOT NULL,
    overview text NOT NULL,
    release_date text NOT NULL,
    year integer NOT NULL DEFAULT 0,
    runtime integer NOT NULL DEFAULT 0,
    genre_ids integer[],
    vote_average float NOT NULL DEFAULT 0,
    vote_count integer NOT NULL DEFAULT 0,
    poster_path text,
    backdrop_path text,
    version integer NOT NULL DEFAULT 1
);