CREATE TABLE IF NOT EXISTS account_account_holder(
    account_id int NOT NULL REFERENCES accounts ON DELETE CASCADE,
    account_holder_id int NOT NULL REFERENCES account_holders ON DELETE CASCADE,
    created_at timestamp(0) with time zone NOT NULL
);

