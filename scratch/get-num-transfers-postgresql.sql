SELECT
    DATE_TRUNC('minute', CREATED_AT) AS MINUTE,
    COUNT(*) AS NUMBER_OF_TRANSFERS
FROM
    TRANSFERS
GROUP BY
    MINUTE
ORDER BY
    MINUTE;