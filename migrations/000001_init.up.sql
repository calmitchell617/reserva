CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS organizations(
    id serial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    can_request_transfers bool NOT NULL,
    transfer_request_callback_endpoint text UNIQUE,
    can_manage_accounts bool NOT NULL,
    transfer_request_authorization_endpoint text UNIQUE,
    minimum_reserve_ratio numeric(3, 2) NOT NULL,
    email citext UNIQUE,
    phone text UNIQUE,
    alert_endpoint text UNIQUE,
    external_id text UNIQUE,
    created_at timestamp NOT NULL,
    version int NOT NULL
);

INSERT INTO organizations(id, name, can_request_transfers, can_manage_accounts, minimum_reserve_ratio, created_at, version)
    VALUES (1, 'Central Bank', TRUE, TRUE, 0.0, now(), 1)
ON CONFLICT (id)
    DO NOTHING;

CREATE TABLE IF NOT EXISTS users(
    id serial PRIMARY KEY,
    organization_id bigint NOT NULL REFERENCES organizations ON DELETE CASCADE,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL,
    created_at timestamp NOT NULL,
    version int NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions(
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    central_bank_only bool NOT NULL
);

INSERT INTO permissions(name, central_bank_only)
    VALUES ('Create organization', TRUE),
('Show organization', FALSE),
('List organizations', TRUE),
('Update organization info', FALSE),
('Update organization permissions', TRUE),
('Delete organization', TRUE),
('Send alert to organization', TRUE),
('Create user', FALSE),
('Show user', FALSE),
('List users', FALSE),
('Update user info', FALSE),
('Update user permissions', FALSE),
('Delete user', FALSE),
('Create account', FALSE),
('Show account', FALSE),
('List accounts', FALSE),
('Update account', FALSE),
('Delete account', FALSE),
(`Show accounts at central bank`, FALSE),
(`List accounts at central bank`, FALSE),
('Create account holder type', TRUE),
('Show account holder type', FALSE),
('List account holder types', FALSE),
('Update account holder type', TRUE),
('Delete account holder type', TRUE),
('Create account holder', FALSE),
('Show account holder', FALSE),
('List account holders', FALSE),
('Update account holder', FALSE),
('Delete account holder', TRUE),
('Associate account holder with account', FALSE),
('Dissociate account holder from account', FALSE),
('Create card', FALSE),
('Show card', FALSE),
('List cards', FALSE),
('Update card', FALSE),
('Create transfer request', FALSE),
('Show transfer request', FALSE),
('List transfer requests', FALSE),
('Update transfer request', FALSE),
('Create transfer', FALSE),
('Show transfer', FALSE),
('List transfers', FALSE),
('Create hold request', FALSE),
('Show hold request', FALSE),
('List hold requests', FALSE),
('Create hold', FALSE),
('Show hold', FALSE),
('List holds', FALSE),
('Update hold', FALSE)
ON CONFLICT (name)
    DO NOTHING;

CREATE TABLE IF NOT EXISTS user_permission(
    user_id int NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id int NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

CREATE TABLE IF NOT EXISTS accounts(
    id bigserial PRIMARY KEY,
    organization_id int NOT NULL REFERENCES organizations ON DELETE RESTRICT,
    name text NOT NULL,
    balance bigint NOT NULL,
    active bool NOT NULL,
    created_at timestamp NOT NULL,
    version int NOT NULL,
    -- the only account that should have a negative balance is the central bank's default account
    CHECK (balance >= 0 OR id = 1)
);

CREATE INDEX IF NOT EXISTS idx_accounts_organization_id ON accounts(organization_id);

-- create the central bank's money supply account, if it doesn't exist
INSERT INTO accounts(id, organization_id, name, balance, active, created_at, version)
    VALUES (1, 1, 'Money supply' 0, TRUE, now(), 1)
ON CONFLICT (id)
    DO NOTHING;

CREATE TABLE IF NOT EXISTS account_holder_types(
    id serial PRIMARY KEY,
    name text UNIQUE NOT NULL,
    description text
);

INSERT INTO account_holder_types(name)
    VALUES ('Person'),
('Other');

CREATE TABLE IF NOT EXISTS account_holders(
    id bigserial PRIMARY KEY,
    account_holder_type_id int NOT NULL REFERENCES account_holder_types ON DELETE RESTRICT,
    name text NOT NULL,
    external_id text UNIQUE NOT NULL name text NOT NULL,
    created_at timestamp NOT NULL
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
    created_at timestamp NOT NULL,
    version int NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);

CREATE INDEX IF NOT EXISTS idx_cards_account_holder_id ON cards(account_holder_id);

CREATE TABLE IF NOT EXISTS transfer_requests(
    id bigserial PRIMARY KEY,
    card_id bigint REFERENCES cards ON DELETE RESTRICT,
    issuing_account_id bigint REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_id bigint NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_organization_id bigint NOT NULL REFERENCES organizations ON DELETE RESTRICT,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL,
    cancelled_at timestamp,
    version int NOT NULL,
    -- check that a card id or issuing account id is provided, but not both
    CHECK ((card_id IS NOT NULL AND issuing_account_id IS NULL) OR (card_id IS NULL AND issuing_account_id IS NOT NULL))
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
    cancelled_at timestamp NOT NULL,
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

