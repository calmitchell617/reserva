delete from cards;

delete from accounts;

delete from depositors;

delete from bank_ip;

delete from banks;

do $$
begin

for i in 1..1000 loop
insert into
  banks (
    id,
    name,
    email,
    password_hash,
    activated,
    version,
    external_id
  )
values
  (i,
  gen_random_uuid(),
  gen_random_uuid(),
  decode(replace(gen_random_uuid()::text, '-', ''), 'hex'),
  true,
  2,
  gen_random_uuid());
end loop;

-- for i in 1..10000 loop
--   insert into
--   bank_ip (bank_id, ip)
--   values
--   (trunc(random() * 1000 + 1),
--   CONCAT(
--   TRUNC(RANDOM() * 250 + 2), '.' , 
--   TRUNC(RANDOM() * 250 + 2), '.', 
--   TRUNC(RANDOM() * 250 + 2), '.',
--   TRUNC(RANDOM() * 250 + 2))::INET);
-- end loop;

-- for i in 1..10000000 loop
--   insert into
--     accounts (
--       id,
--       bank_id
--     )
--   values
--     (i,
--     trunc(random() * 1000 + 1));
-- end loop;

-- for i in 1..10000000 loop
-- insert into
--   depositors (
--   id,
--   external_id,
--   name,
--   address
--   )
-- values
--   (i,
--   gen_random_uuid(),
--   gen_random_uuid(),
--   gen_random_uuid());
-- end loop;

-- for i in 1..10000000 loop
-- insert into
--   cards (
--     id,
--     account_id,
--     depositor,
--     public_key,
--     pin_hash,
--     expiry
--   )
-- values
--   (i,
--   trunc(random() * 10000000 + 1),
--   trunc(random() * 10000000 + 1),
--   decode(replace(gen_random_uuid()::text, '-', ''), 'hex'),
--   decode(replace(gen_random_uuid()::text, '-', ''), 'hex'),
--   '2025-2-2');
-- end loop;

end;
$$;