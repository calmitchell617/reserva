INSERT INTO cards(id, account_id, expiration_date, frozen)
SELECT
    series_column,
    series_column,
    now() + interval '1 year',
    FALSE
FROM
    generate_series(1, 1000000) AS series_column;

