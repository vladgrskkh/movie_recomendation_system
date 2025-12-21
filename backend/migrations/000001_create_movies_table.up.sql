CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS movies (
    id bigint PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    original_title text NOT NULL DEFAULT '',
    overview text NOT NULL DEFAULT '',
    release_date text NOT NULL DEFAULT '',
    year integer NOT NULL,
    runtime integer NOT NULL DEFAULT 0,
    vote_average float NOT NULL DEFAULT 0,
    vote_count integer NOT NULL DEFAULT 0,
    poster_path text NOT NULL DEFAULT '',
    backdrop_path text NOT NULL DEFAULT '',
    version integer NOT NULL DEFAULT 1
);