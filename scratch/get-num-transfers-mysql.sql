SELECT
    hour(CREATED_AT) AS hour_created_at,
    minute(CREATED_AT) AS minute_created_at,
    COUNT(*) AS NUMBER_OF_TRANSFERS
FROM
    transfers
GROUP BY
    hour_created_at, minute_created_at
ORDER BY
    hour_created_at, minute_created_at;