INSERT INTO tokens(user_id, expires_at)
SELECT
    id,
    now() + interval '1 year'
FROM
    users;

