-- switch to postgresql db
\c postgres;

-- force drop reserva db
DROP DATABASE IF EXISTS reserva with (force);

-- create reserva db

vacuum full;

create database reserva;

-- switch to reserva db
\c reserva;

vacuum full;

CREATE TABLE IF NOT EXISTS organizations(id smallserial PRIMARY KEY);

CREATE TABLE IF NOT EXISTS users(
    id smallserial PRIMARY KEY,
    organization_id smallint NOT NULL REFERENCES organizations ON DELETE CASCADE,
    frozen bool NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions(
    id smallserial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS accounts(
    id serial PRIMARY KEY,
    organization_id smallint NOT NULL REFERENCES organizations ON DELETE CASCADE,
    balance bigint NOT NULL,
    frozen bool NOT NULL,
    check (balance >= 0)
);

CREATE TABLE IF NOT EXISTS cards(
    id bigserial PRIMARY KEY,
    account_id int NOT NULL REFERENCES accounts ON DELETE CASCADE,
    expiration_date date NOT NULL,
    security_code smallint NOT NULL,
    frozen bool NOT NULL
);

CREATE TABLE IF NOT EXISTS transfers(
    id bigserial PRIMARY KEY,
    card_id bigint REFERENCES cards ON DELETE CASCADE,
    from_account_id int NOT NULL REFERENCES accounts ON DELETE CASCADE,
    to_account_id int NOT NULL REFERENCES accounts ON DELETE CASCADE,
    requesting_user_id int NOT NULL REFERENCES users ON DELETE CASCADE,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS tokens(
    hash uuid DEFAULT gen_random_uuid(),
    permission_id smallint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    user_id smallint NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at timestamp NOT NULL,
    primary key (hash, permission_id)
);

CREATE
OR REPLACE FUNCTION transfer_funds(
    p_card_id BIGINT,
    p_from_account_id INT,
    p_to_account_id INT,
    p_requesting_user_id SMALLINT,
    p_amount bigint,
    p_created_at TIMESTAMP
) RETURNS INT AS $$ DECLARE v_transfer_id INT;

BEGIN
UPDATE
    accounts
SET
    balance = balance - p_amount
WHERE
    id = p_from_account_id;

UPDATE
    accounts
SET
    balance = balance + p_amount
WHERE
    id = p_to_account_id;

INSERT INTO
    transfers (
        card_id,
        from_account_id,
        to_account_id,
        requesting_user_id,
        amount,
        created_at
    )
VALUES
    (
        p_card_id,
        p_from_account_id,
        p_to_account_id,
        p_requesting_user_id,
        p_amount,
        p_created_at
    ) RETURNING id INTO v_transfer_id;

RETURN v_transfer_id;

END;

$$ LANGUAGE plpgsql;

INSERT INTO organizations(id)
SELECT
    generate_series(1, 10);

INSERT INTO accounts(id, organization_id, balance, frozen)
SELECT
    series_column,
    floor(1 + random() * 10)::int,
    1000000000, -- everyone in my country is a billionaire. inflation be damned
    FALSE
FROM
    generate_series(1, 100000) AS series_column;

INSERT INTO users(id, organization_id, frozen)
SELECT
    series_column,
    series_column,
    FALSE
FROM
    generate_series(1, 10) AS series_column;

INSERT INTO
    permissions(id, name)
VALUES
    (1, 'transfer_requests:create'),
    (2, 'transfers:create') ON CONFLICT (name) DO NOTHING;

INSERT INTO cards(id, account_id, expiration_date, security_code, frozen)
SELECT
    series_column,
    series_column,
    now() + interval '1 year',
    random() * 999 + 1,
    FALSE
FROM
    generate_series(1, 100000) AS series_column;

INSERT INTO
    tokens(user_id, permission_id, expires_at)
SELECT
    id,
    1,
    now() + interval '1 year'
FROM
    users;

INSERT INTO
    tokens(hash, user_id, permission_id, expires_at)
SELECT
    hash,
    user_id,
    2,
    expires_at
FROM
    tokens;

