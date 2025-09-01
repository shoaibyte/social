CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    email citext UNIQUE NOT NULL,
    USERNAME varchar(255) UNIQUE NOT NULL,
    PASSWORD bytea NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
)