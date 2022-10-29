create table banks (
  id bigserial primary key,
  name text not null unique,
  email citext not null unique,
  password_hash bytea not null,
  balance_in_cents bigint not null default 0,
  activated bool not null default false,
  frozen boolean not null default false,
  ips inet[] not null default array[]::inet[],
  version bigint not null default 0
);