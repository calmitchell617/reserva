CREATE TABLE IF NOT EXISTS accounts(
    id serial PRIMARY KEY,
    account_caretaker int NOT NULL REFERENCES caretakers ON DELETE RESTRICT, -- can change
    balance bigint NOT NULL, -- can change
    hold_balance bigint NOT NULL, -- can change
    created_at timestamp(0) with time zone NOT NULL,
    last_modified timestamp(0) with time zone NOT NULL,
    version int NOT NULL
);

