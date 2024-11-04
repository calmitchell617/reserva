\c postgres;

DROP DATABASE IF EXISTS reserva WITH (FORCE);

CREATE DATABASE reserva;

-- switch to reserva db
\c reserva;
vacuum full;

CREATE UNLOGGED TABLE IF NOT EXISTS organizations(
    id smallserial PRIMARY KEY
);

CREATE UNLOGGED TABLE IF NOT EXISTS users(
    id smallserial PRIMARY KEY,
    organization_id smallint NOT NULL,
    frozen bool NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS permissions(
    id smallserial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE UNLOGGED TABLE IF NOT EXISTS accounts(
    id serial PRIMARY KEY,
    organization_id smallint NOT NULL,
    balance bigint NOT NULL,
    frozen bool NOT NULL,
    CHECK (balance >= 0)
);

CREATE UNLOGGED TABLE IF NOT EXISTS cards(
    id bigserial PRIMARY KEY,
    account_id int NOT NULL,
    expiration_date date NOT NULL,
    security_code smallint NOT NULL,
    frozen bool NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS transfers(
    id bigserial PRIMARY KEY,
    card_id bigint,
    from_account_id int NOT NULL,
    to_account_id int NOT NULL,
    requesting_user_id int NOT NULL,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE UNLOGGED TABLE IF NOT EXISTS tokens(
    hash uuid DEFAULT gen_random_uuid(),
    permission_id smallint NOT NULL,
    user_id smallint NOT NULL,
    expires_at timestamp NOT NULL,
    PRIMARY KEY (hash, permission_id)
);

CREATE OR REPLACE FUNCTION transfer_funds(p_card_id bigint, p_from_account_id int, p_to_account_id int, p_requesting_user_id smallint, p_amount bigint, p_created_at timestamp)
    RETURNS int
    AS $$
DECLARE
    v_transfer_id int;
BEGIN

    UPDATE
        accounts
    SET
        balance = CASE
                    WHEN id = p_from_account_id THEN balance - p_amount
                    WHEN id = p_to_account_id THEN balance + p_amount
                END
    WHERE
        id IN (p_from_account_id, p_to_account_id);

    
    INSERT INTO transfers(card_id, from_account_id, to_account_id, requesting_user_id, amount, created_at)
        VALUES (p_card_id, p_from_account_id, p_to_account_id, p_requesting_user_id, p_amount, p_created_at)
    RETURNING
        id INTO v_transfer_id;
    RETURN v_transfer_id;
END;
$$
LANGUAGE plpgsql;

SET synchronous_commit TO OFF;

DO $$
DECLARE
    num_organizations INT := 100;
    num_accounts INT := 10000000;
    num_transfers INT := 100000000;
BEGIN

SET CONSTRAINTS ALL DEFERRED;

INSERT INTO organizations(id)
SELECT
    generate_series(1, num_organizations);

INSERT INTO accounts(id, organization_id, balance, frozen)
SELECT
    series_column,
    ceil(random() * num_organizations)::int,
    50000000,
    FALSE
FROM
    generate_series(1, num_accounts) AS series_column;

INSERT INTO users(id, organization_id, frozen)
SELECT
    series_column,
    series_column,
    FALSE
FROM
    generate_series(1, num_organizations) AS series_column;

INSERT INTO permissions(id, name)
    VALUES (1, 'transfer_requests:create'),
(2, 'transfers:create')
ON CONFLICT (name)
    DO NOTHING;

INSERT INTO cards(id, account_id, expiration_date, security_code, frozen)
SELECT
    series_column,
    series_column,
    now() + interval '1 year',
    random() * 999 + 1,
    FALSE
FROM
    generate_series(1, num_accounts) AS series_column;

INSERT INTO tokens(user_id, permission_id, expires_at)
SELECT
    id,
    1,
    now() + interval '1 year'
FROM
    users;

INSERT INTO tokens(hash, user_id, permission_id, expires_at)
SELECT
    hash,
    user_id,
    2,
    expires_at
FROM
    tokens;

-- insert into transfers (card_id, from_account_id, to_account_id, requesting_user_id, amount, created_at)
-- select
--     ceil(random() * num_accounts)::int,
--     ceil(random() * num_accounts)::int,
--     ceil(random() * num_accounts)::int,
--     ceil(random() * num_organizations)::int,
--     floor(1 + random() * 100)::int,
--     now() - interval '1 year' * random()
-- from
--     generate_series(1, num_transfers);


SET CONSTRAINTS ALL IMMEDIATE;
COMMIT;

END $$;

SET synchronous_commit TO ON;

ALTER TABLE organizations SET UNLOGGED;

ALTER TABLE users SET UNLOGGED;

ALTER TABLE permissions SET UNLOGGED;

ALTER TABLE accounts SET UNLOGGED;

ALTER TABLE cards SET UNLOGGED;

ALTER TABLE transfers SET UNLOGGED;

ALTER TABLE tokens SET UNLOGGED;

BEGIN;

-- Add foreign key constraints
ALTER TABLE users
ADD CONSTRAINT fk_users_organization
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE accounts
ADD CONSTRAINT fk_accounts_organization
FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

ALTER TABLE cards
ADD CONSTRAINT fk_cards_account
FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE;

ALTER TABLE transfers
ADD CONSTRAINT fk_transfers_card
FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE;

ALTER TABLE transfers
ADD CONSTRAINT fk_transfers_from_account
FOREIGN KEY (from_account_id) REFERENCES accounts(id) ON DELETE CASCADE;

ALTER TABLE transfers
ADD CONSTRAINT fk_transfers_to_account
FOREIGN KEY (to_account_id) REFERENCES accounts(id) ON DELETE CASCADE;

ALTER TABLE transfers
ADD CONSTRAINT fk_transfers_requesting_user
FOREIGN KEY (requesting_user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE tokens
ADD CONSTRAINT fk_tokens_permission
FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE;

ALTER TABLE tokens
ADD CONSTRAINT fk_tokens_user
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes on foreign key columns to improve performance
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_accounts_organization_id ON accounts(organization_id);
CREATE INDEX idx_cards_account_id ON cards(account_id);
CREATE INDEX idx_transfers_card_id ON transfers(card_id);
CREATE INDEX idx_transfers_from_account_id ON transfers(from_account_id);
CREATE INDEX idx_transfers_to_account_id ON transfers(to_account_id);
CREATE INDEX idx_transfers_requesting_user_id ON transfers(requesting_user_id);
CREATE INDEX idx_tokens_permission_id ON tokens(permission_id);
CREATE INDEX idx_tokens_user_id ON tokens(user_id);

-- Create unique index on permissions.name
CREATE UNIQUE INDEX idx_permissions_name ON permissions(name);

-- Commit the transaction
COMMIT;