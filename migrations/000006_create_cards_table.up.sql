CREATE TABLE IF NOT EXISTS cards(
    id bigint PRIMARY KEY,
    account_id int NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    account_holder_id int NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    active bool NOT NULL DEFAULT FALSE, -- can change
    expiration_date date NOT NULL,
    security_code smallint NOT NULL,
    created_at timestamp(0) with time zone NOT NULL,
    last_modified timestamp(0) with time zone NOT NULL,
    version int NOT NULL
);

