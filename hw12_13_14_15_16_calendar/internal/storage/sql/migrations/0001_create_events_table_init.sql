-- +goose Up
DROP TABLE IF EXISTS events;

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    time_before INTERVAL,

    period TSRANGE GENERATED ALWAYS AS (
        tsrange(start_time AT TIME ZONE 'UTC', end_time AT TIME ZONE 'UTC', '[]')
    ) STORED,

    EXCLUDE USING GIST (
        user_id WITH =,
        period WITH &&
    )
);

-- +goose Down
DROP TABLE events;