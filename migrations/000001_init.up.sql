CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS caretakers(
    id bigserial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    endpoint text UNIQUE NOT NULL,
    active bool NOT NULL,
    version int NOT NULL
);

-- create the central bank as caretaker id 1, if it doesn't exist
INSERT INTO caretakers(id, name, endpoint, active, version)
    VALUES (1, 'Central bank', '', TRUE, 1)
ON CONFLICT (id)
    DO NOTHING;

CREATE TABLE permission_levels(
    id int PRIMARY KEY,
    name text UNIQUE NOT NULL
);

INSERT INTO permission_levels(id, name)
    VALUES (1, 'Standard'),
(2, 'Admin'),
(3, 'Super admin');

CREATE TABLE IF NOT EXISTS agents(
    id bigserial PRIMARY KEY,
    caretaker_id bigint NOT NULL REFERENCES caretakers ON DELETE CASCADE,
    permission_level_id int NOT NULL REFERENCES permission_levels ON DELETE RESTRICT,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts(
    id bigserial PRIMARY KEY,
    caretaker_id bigint NOT NULL REFERENCES caretakers ON DELETE RESTRICT,
    balance bigint NOT NULL,
    version int NOT NULL,
    -- ensure the only account that can have a negative balance is the central bank account
    CHECK (balance >= 0 OR caretaker_id = 1)
);

-- create central bank account, if it doesn't exist
INSERT INTO accounts(caretaker_id, balance, version)
    VALUES (1, 0, 1)
ON CONFLICT (caretaker_id)
    DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_accounts_caretaker_id ON accounts(caretaker_id);

CREATE TABLE IF NOT EXISTS account_holders(
    id bigserial PRIMARY KEY,
    external_id text UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS account_account_holder(
    account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    account_holder_id bigint NOT NULL REFERENCES account_holders ON DELETE CASCADE,
    PRIMARY KEY (account_id, account_holder_id)
);

CREATE TABLE IF NOT EXISTS cards(
    id bigint PRIMARY KEY,
    account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    account_holder_id bigint NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    active bool NOT NULL DEFAULT FALSE,
    expiration_date date NOT NULL,
    security_code smallint NOT NULL,
    version int NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);

CREATE INDEX IF NOT EXISTS idx_cards_account_holder_id ON cards(account_holder_id);

CREATE TABLE IF NOT EXISTS transfer_requests(
    id bigserial PRIMARY KEY,
    card_id bigint NOT NULL REFERENCES cards ON DELETE RESTRICT,
    acquiring_account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_holder_id bigint NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL,
    version int NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transfer_requests_card_id ON transfer_requests(card_id);

CREATE INDEX IF NOT EXISTS idx_transfer_requests_acquiring_account_id ON transfer_requests(acquiring_account_id);

CREATE INDEX IF NOT EXISTS idx_transfer_requests_acquiring_account_holder_id ON transfer_requests(acquiring_account_holder_id);

CREATE INDEX IF NOT EXISTS idx_transfer_requests_created_at ON transfer_requests(created_at);

CREATE TABLE IF NOT EXISTS transfers(
    id bigserial PRIMARY KEY,
    transfer_request_id bigint REFERENCES transfer_requests ON DELETE RESTRICT,
    from_account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    to_account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transfers_transfer_request_id ON transfers(transfer_request_id);

CREATE INDEX IF NOT EXISTS idx_transfers_from_account_id ON transfers(from_account_id);

CREATE INDEX IF NOT EXISTS idx_transfers_to_account_id ON transfers(to_account_id);

CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at);

CREATE TABLE IF NOT EXISTS hold_requests(
    id bigserial PRIMARY KEY,
    card_id bigint NOT NULL REFERENCES cards ON DELETE RESTRICT,
    acquiring_account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_holder_id bigint NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL,
    version int NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_hold_requests_card_id ON hold_requests(card_id);

CREATE INDEX IF NOT EXISTS idx_hold_requests_acquiring_account_id ON hold_requests(acquiring_account_id);

CREATE INDEX IF NOT EXISTS idx_hold_requests_acquiring_account_holder_id ON hold_requests(acquiring_account_holder_id);

CREATE INDEX IF NOT EXISTS idx_hold_requests_created_at ON hold_requests(created_at);

CREATE TABLE IF NOT EXISTS holds(
    id bigserial PRIMARY KEY,
    hold_request_id bigint NOT NULL REFERENCES hold_requests ON DELETE RESTRICT,
    account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL,
    expires_at timestamp NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_holds_account_id ON holds(account_id);

CREATE INDEX IF NOT EXISTS idx_holds_expires_at ON holds(expires_at);

CREATE TABLE IF NOT EXISTS tokens(
    hash bytea PRIMARY KEY,
    caretaker_id bigint NOT NULL REFERENCES caretakers ON DELETE CASCADE,
    expires_at timestamp NOT NULL,
    scope text NOT NULL
);

