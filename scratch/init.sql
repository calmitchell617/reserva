CREATE EXTENSION IF NOT EXISTS citext;

DROP TABLE IF EXISTS tokens;

DROP TABLE IF EXISTS payment_requests;

DROP TABLE IF EXISTS hold_requests;

DROP TABLE IF EXISTS cards;

DROP TABLE IF EXISTS account_account_holder;

DROP TABLE IF EXISTS account_holders;

DROP TABLE IF EXISTS accounts;

DROP TABLE IF EXISTS caretakers;

CREATE TABLE IF NOT EXISTS caretakers(
    id serial PRIMARY KEY,
    is_admin bool NOT NULL DEFAULT FALSE,
    external_id text UNIQUE NOT NULL,
    endpoint text UNIQUE NOT NULL, -- can change
    password_hash bytea NOT NULL, -- can change
    activated bool NOT NULL, -- can change
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT now(),
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts(
    id serial PRIMARY KEY,
    account_caretaker int NOT NULL REFERENCES caretakers ON DELETE RESTRICT, -- can change
    balance bigint NOT NULL, -- can change
    hold_balance bigint NOT NULL, -- can change
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT now(),
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS account_holders(
    id serial PRIMARY KEY,
    external_id text UNIQUE NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS account_account_holder(
    account_id int NOT NULL REFERENCES accounts ON DELETE CASCADE,
    account_holder_id int NOT NULL REFERENCES account_holders ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cards(
    id bigint PRIMARY KEY,
    account_id int NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    account_holder_id int NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    expiration_date date NOT NULL,
    security_code smallint NOT NULL,
    active bool NOT NULL DEFAULT FALSE, -- can change
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT now(),
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS payment_requests(
    id bigserial PRIMARY KEY,
    acquiring_account int NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_holder int NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    card_id bigint NOT NULL REFERENCES cards ON DELETE RESTRICT,
    amount bigint NOT NULL CHECK (amount > 0),
    completed bool NOT NULL DEFAULT FALSE, -- can change
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT now(),
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS hold_requests(
    id bigserial PRIMARY KEY,
    acquiring_account int NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_holder int NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    card_id bigint NOT NULL REFERENCES cards ON DELETE RESTRICT,
    amount bigint NOT NULL CHECK (amount > 0),
    granted bool NOT NULL DEFAULT FALSE, -- can change
    active bool NOT NULL DEFAULT FALSE, -- can change
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    last_modified timestamp(0) with time zone NOT NULL DEFAULT now(),
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS tokens(
    hash bytea PRIMARY KEY,
    caretaker_id int NOT NULL REFERENCES caretakers ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);

