CREATE TABLE IF NOT EXISTS payment_requests(
    id bigserial PRIMARY KEY,
    card_id bigint NOT NULL REFERENCES cards ON DELETE RESTRICT,
    acquiring_account int NOT NULL REFERENCES accounts ON DELETE RESTRICT,
    acquiring_account_holder int NOT NULL REFERENCES account_holders ON DELETE RESTRICT,
    amount bigint NOT NULL CHECK (amount > 0),
    completed bool NOT NULL DEFAULT FALSE, -- can change
    created_at timestamp(0) with time zone NOT NULL,
    last_modified timestamp(0) with time zone NOT NULL,
    version int NOT NULL
);

