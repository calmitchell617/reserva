INSERT INTO user_permission(user_id, permission_id)
SELECT
    id,
    1
FROM
    users;

INSERT INTO user_permission(user_id, permission_id)
SELECT
    id,
    2
FROM
    users;

