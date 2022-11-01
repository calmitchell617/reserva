-- drop table if exists tokens;

drop table if exists transfers;

drop table if exists accounts;

drop table if exists banks;

create rowstore table banks (
  username varchar(32) primary key,
  admin bool not null,
  password_hash varbinary(64) not null,
  balance_in_cents bigint not null default 0,
  frozen bool not null default false,
  version bigint unsigned not null default 0
);

create rowstore table accounts (
  id bigint unsigned auto_increment,
  bank_username varchar(32) not null,
  kyc_data json not null,
  frozen boolean not null default false,
  balance_in_cents bigint not null default 0,
  version bigint not null default 0,
  primary key (id, bank_username),
  shard key (bank_username)
);

create table transfers (
  id bigint auto_increment not null,
  source_account_id bigint not null,
  target_account_id bigint not null,
  pos_location geographypoint not null,
  created_at datetime not null default now(),
  amount_in_cents bigint not null,
  primary key (id, source_account_id),
  shard key (source_account_id)
);

-- create table tokens (
--   secret_hash bytea primary key,
--   bank_id bigint not null references banks on delete cascade,
--   expiry timestamp not null,
--   scope text not null
-- );