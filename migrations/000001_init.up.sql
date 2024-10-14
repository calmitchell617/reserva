CREATE TABLE IF NOT EXISTS organizations(
    id serial PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users(
    id serial PRIMARY KEY,
    organization_id int NOT NULL REFERENCES organizations ON DELETE CASCADE,
    frozen bool NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions(
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

INSERT INTO permissions(id, name)
    VALUES (1, 'transfer_requests:create'),
(2, 'transfers:create')
ON CONFLICT (name)
    DO NOTHING;

CREATE TABLE IF NOT EXISTS user_permission(
    user_id int NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id int NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

CREATE TABLE IF NOT EXISTS accounts(
    id bigserial PRIMARY KEY,
    organization_id int NOT NULL REFERENCES organizations ON DELETE CASCADE,
    balance bigint NOT NULL,
    frozen bool NOT NULL
);

CREATE TABLE IF NOT EXISTS cards(
    id bigint PRIMARY KEY,
    account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    expiration_date date NOT NULL,
    frozen bool NOT NULL
);

CREATE TABLE IF NOT EXISTS transfer_requests(
    id bigserial PRIMARY KEY,
    card_id bigint REFERENCES cards ON DELETE CASCADE,
    issuing_account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    acquiring_account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS transfers(
    id bigserial PRIMARY KEY,
    transfer_request_id bigint REFERENCES transfer_requests ON DELETE CASCADE,
    from_account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    to_account_id bigint NOT NULL REFERENCES accounts ON DELETE CASCADE,
    amount bigint NOT NULL,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS tokens(
    hash uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id int NOT NULL REFERENCES users ON DELETE CASCADE,
    expires_at timestamp NOT NULL
);

