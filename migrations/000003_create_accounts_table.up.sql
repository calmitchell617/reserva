create table accounts (
  id bigserial primary key,
  bank_id bigint references banks,
  balance_in_cents bigint not null default 0,
  frozen boolean not null default false,
  version bigint not null default 0
);