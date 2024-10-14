INSERT INTO users(id, organization_id, frozen)
SELECT
    series_column,
    series_column,
    FALSE
FROM
    generate_series(1, 1000) AS series_column;

