create table cards (
  id bigserial primary key,
  account_id bigint not null references accounts,
  private_key bytea not null,
  password_hash bytea not null,
  expiry timestamp not null,
  version bigint not null default 0
);