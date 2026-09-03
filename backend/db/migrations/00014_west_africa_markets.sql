-- +goose Up
-- Extension de la couverture marchés : toute l'Afrique de l'Ouest
-- (CEDEAO + Mauritanie) et l'Afrique centrale (CEMAC + RD Congo).
-- Idempotent : les marchés déjà seedés sont conservés.

INSERT INTO markets (code, name, currency, language) VALUES
    -- Afrique de l'Ouest (complément CEDEAO + Mauritanie)
    ('BF', 'Burkina Faso',    'XOF', 'fr'),
    ('CV', 'Cap-Vert',        'CVE', 'pt'),
    ('GM', 'Gambie',          'GMD', 'en'),
    ('GN', 'Guinée',          'GNF', 'fr'),
    ('GW', 'Guinée-Bissau',   'XOF', 'pt'),
    ('LR', 'Libéria',         'LRD', 'en'),
    ('ML', 'Mali',            'XOF', 'fr'),
    ('MR', 'Mauritanie',      'MRU', 'fr'),
    ('NE', 'Niger',           'XOF', 'fr'),
    ('SL', 'Sierra Leone',    'SLE', 'en'),
    ('TG', 'Togo',            'XOF', 'fr'),
    -- Afrique centrale (CEMAC + RD Congo)
    ('CM', 'Cameroun',            'XAF', 'fr'),
    ('GA', 'Gabon',               'XAF', 'fr'),
    ('CG', 'Congo',               'XAF', 'fr'),
    ('TD', 'Tchad',               'XAF', 'fr'),
    ('CF', 'République centrafricaine', 'XAF', 'fr'),
    ('GQ', 'Guinée équatoriale',  'XAF', 'fr'),
    ('CD', 'RD Congo',            'CDF', 'fr')
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DELETE FROM markets
WHERE code IN ('BF','CV','GM','GN','GW','LR','ML','MR','NE','SL','TG','CM','GA','CG','TD','CF','GQ','CD');
