INSERT INTO accounts(organization_id, balance, frozen)
SELECT
    random(1, 1000),
    1000000000,
    FALSE
FROM
    generate_series(1, 1000000) AS series_column;

