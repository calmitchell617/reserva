CREATE TABLE IF NOT EXISTS account_holders(
    id serial PRIMARY KEY,
    external_id text UNIQUE NOT NULL,
    email citext NOT NULL UNIQUE, -- can change
    created_at timestamp(0) with time zone NOT NULL,
    last_modified timestamp(0) with time zone NOT NULL,
    version int NOT NULL
);

