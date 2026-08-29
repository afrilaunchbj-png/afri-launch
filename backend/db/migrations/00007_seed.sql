-- +goose Up
-- Données de référence (non sensibles). Les opportunités du catalogue sont
-- des exemples qualitatifs (signal « estimated »/« hypothesis ») sans
-- statistiques inventées : aucune donnée n'est présentée comme vérifiée.

INSERT INTO markets (code, name, currency, language) VALUES
    ('SN', 'Sénégal',        'XOF', 'fr'),
    ('CI', 'Côte d''Ivoire', 'XOF', 'fr'),
    ('BJ', 'Bénin',          'XOF', 'fr'),
    ('NG', 'Nigeria',        'NGN', 'en'),
    ('KE', 'Kenya',          'KES', 'en'),
    ('GH', 'Ghana',          'GHS', 'en');

-- Coûts de génération configurables (valeurs d'exemple du master prompt).
INSERT INTO generation_costs (operation, name, credits) VALUES
    ('niche_research',    'Recherche de niche',      5),
    ('idea_generation',   'Génération d''idées',     2),
    ('ebook_generation',  'Génération d''ebook',    20),
    ('image_generation',  'Génération d''image',     3),
    ('video_generation',  'Génération de vidéo',    15),
    ('sales_page',        'Page de vente',           5);

-- Packs de crédits (prix en unités mineures, exemples).
INSERT INTO plans (name, credits, price_minor, currency, sort_order) VALUES
    ('Démarrage', 50,   2500, 'XOF', 1),
    ('Croissance', 200, 9000, 'XOF', 2),
    ('Pro',        500, 20000, 'XOF', 3);

-- Catalogue d'opportunités de référence (qualitatif).
INSERT INTO opportunities (title, summary, country, sector, language, difficulty, signal, score, scores) VALUES
    (
        'Gestion de stock pour boutiques de retail indépendantes',
        'La fragmentation du commerce de détail local crée des inefficacités. Les solutions existantes sont souvent perçues comme complexes ou trop coûteuses pour des exploitants en point de vente unique. (Hypothèse à valider par entretiens terrain.)',
        'Nigeria', 'Retail Tech', 'en', 'medium', 'hypothesis', 82,
        '{"demand": 85, "pain": 90, "competition": 30, "purchasing_power": 60, "digital_fit": 95, "evidence_strength": 20}'
    ),
    (
        'Traçabilité de la chaîne du froid pour petits producteurs agricoles',
        'Des pertes post-récolte sont régulièrement évoquées faute de visibilité sur la température de transport. Un outil de suivi léger (SMS/USSD) pourrait fournir une traçabilité de base. (Hypothèse sectorielle.)',
        'Sénégal', 'AgriTech', 'fr', 'high', 'hypothesis', 68,
        '{"demand": 60, "pain": 80, "competition": 45, "purchasing_power": 50, "digital_fit": 80, "evidence_strength": 15}'
    ),
    (
        'Comptabilité simplifiée pour micro-commerçants',
        'De nombreux commerçants tiennent encore leurs comptes sur papier. Une solution mobile simple, adaptée au langage local et aux petits montants, pourrait répondre à un besoin récurrent. (Hypothèse à valider.)',
        'Côte d''Ivoire', 'FinTech', 'fr', 'low', 'hypothesis', 74,
        '{"demand": 80, "pain": 75, "competition": 55, "purchasing_power": 45, "digital_fit": 90, "evidence_strength": 18}'
    ),
    (
        'Préparation aux concours et examens nationaux',
        'La demande d''accompagnement à la préparation des examens est forte et récurrente, avec un contenu digital encore sous-exploité dans certaines zones. (Hypothèse de marché.)',
        'Bénin', 'EdTech', 'fr', 'low', 'estimated', 71,
        '{"demand": 85, "pain": 65, "competition": 40, "purchasing_power": 50, "digital_fit": 85, "evidence_strength": 25}'
    ),
    (
        'Logistique du dernier kilomètre en zone urbaine dense',
        'La livraison du dernier kilomètre reste un goulot d''étranglement dans les grandes villes. Un guide opérationnel ou un outil de coordination léger pourrait aider les livreurs indépendants. (Hypothèse.)',
        'Kenya', 'Logistics', 'en', 'medium', 'hypothesis', 63,
        '{"demand": 70, "pain": 70, "competition": 65, "purchasing_power": 55, "digital_fit": 75, "evidence_strength": 15}'
    ),
    (
        'Rendez-vous et rappels pour cliniques de quartier',
        'Les petites structures de santé gèrent encore souvent les rendez-vous manuellement. Un rappel automatisé (SMS/WhatsApp) pourrait réduire les absences. (Hypothèse opérationnelle.)',
        'Ghana', 'Health', 'en', 'low', 'estimated', 66,
        '{"demand": 65, "pain": 70, "competition": 35, "purchasing_power": 50, "digital_fit": 80, "evidence_strength": 20}'
    );

-- +goose Down
DELETE FROM opportunities;
DELETE FROM plans;
DELETE FROM generation_costs;
DELETE FROM markets;
