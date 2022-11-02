drop table if exists transfers;

drop table if exists accounts;

drop table if exists banks;

create rowstore table banks (
  username varchar(32) primary key,
  admin bool not null,
  password_hash varbinary(64) not null,
  frozen bool not null default false,
  version bigint unsigned not null default 0,
  shard key (username)
);

create rowstore table accounts (
  id bigint unsigned auto_increment,
  controlling_bank varchar(32) not null,
  metadata json not null,
  frozen boolean not null default false,
  balance_in_cents bigint not null default 0,
  version bigint not null default 0,
  primary key (id, controlling_bank),
  shard key (controlling_bank)
);

create table transfers (
  id bigint auto_increment not null,
  source_account_id bigint not null,
  target_account_id bigint not null,
  created_at datetime not null default now(),
  amount_in_cents bigint not null,
  primary key (id, source_account_id),
  shard key (source_account_id)
);