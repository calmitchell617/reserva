CREATE TABLE IF NOT EXISTS tokens(
    hash bytea PRIMARY KEY,
    caretaker_id int NOT NULL REFERENCES caretakers ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);

