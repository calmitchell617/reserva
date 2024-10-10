CREATE TABLE IF NOT EXISTS caretakers(
    id serial PRIMARY KEY,
    external_id text UNIQUE NOT NULL,
    is_admin bool NOT NULL DEFAULT FALSE,
    endpoint text UNIQUE NOT NULL, -- can change
    activated bool NOT NULL, -- can change
    email citext NOT NULL UNIQUE, -- can change
    password_hash bytea NOT NULL, -- can change
    created_at timestamp(0) with time zone NOT NULL,
    last_modified timestamp(0) with time zone NOT NULL,
    version int NOT NULL
);

