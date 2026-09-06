-- name: ListPaymentCountries :many
SELECT *
FROM payment_countries
ORDER BY name ASC;

-- name: ListEnabledPaymentCountries :many
SELECT *
FROM payment_countries
WHERE enabled = true
ORDER BY name ASC;

-- name: SetPaymentCountryEnabled :exec
UPDATE payment_countries
SET enabled = $2, updated_at = now()
WHERE code = $1;
