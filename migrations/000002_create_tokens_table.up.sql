create table if not exists tokens (
    hash bytea primary key,
    bank_id bigint not null references banks on delete cascade,
    expiry timestamp not null,
    scope text not null
);