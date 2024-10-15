INSERT INTO accounts(organization_id, balance, frozen)
SELECT
    floor(1 + random() * 1000)::int,
    1000000000,
    FALSE
FROM
    generate_series(1, 1000000) AS series_column;

