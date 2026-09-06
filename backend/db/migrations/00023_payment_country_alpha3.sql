-- +goose Up
-- PawaPay attend les pays en ISO 3166-1 alpha-3 (ex. BEN) : la migration
-- 00022 avait utilisé des codes alpha-2 (BJ…) → checkout rejeté.

UPDATE payment_countries SET code = 'BEN', updated_at = now() WHERE code = 'BJ';
UPDATE payment_countries SET code = 'BFA', updated_at = now() WHERE code = 'BF';
UPDATE payment_countries SET code = 'CIV', updated_at = now() WHERE code = 'CI';
UPDATE payment_countries SET code = 'GNB', updated_at = now() WHERE code = 'GW';
UPDATE payment_countries SET code = 'MLI', updated_at = now() WHERE code = 'ML';
UPDATE payment_countries SET code = 'NER', updated_at = now() WHERE code = 'NE';
UPDATE payment_countries SET code = 'SEN', updated_at = now() WHERE code = 'SN';
UPDATE payment_countries SET code = 'TGO', updated_at = now() WHERE code = 'TG';

-- +goose Down
UPDATE payment_countries SET code = 'BJ', updated_at = now() WHERE code = 'BEN';
UPDATE payment_countries SET code = 'BF', updated_at = now() WHERE code = 'BFA';
UPDATE payment_countries SET code = 'CI', updated_at = now() WHERE code = 'CIV';
UPDATE payment_countries SET code = 'GW', updated_at = now() WHERE code = 'GNB';
UPDATE payment_countries SET code = 'ML', updated_at = now() WHERE code = 'MLI';
UPDATE payment_countries SET code = 'NE', updated_at = now() WHERE code = 'NER';
UPDATE payment_countries SET code = 'SN', updated_at = now() WHERE code = 'SEN';
UPDATE payment_countries SET code = 'TG', updated_at = now() WHERE code = 'TGO';
