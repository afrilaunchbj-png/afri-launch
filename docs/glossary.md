# Glossary — AfriLaunch

> Terminologie métier et technique du projet.

## Domaine métier

| Terme | Définition |
|---|---|
| **Opportunité** | Niche de marché identifiée comme exploitable, associée à un ou plusieurs pays, scorée (`Opportunity Score`). |
| **Opportunity Score** | Note 0–100 combinant demande, douleur, pouvoir d'achat, concurrence, adéquation produit digital, facilité de distribution, pertinence locale, complexité, prix/marge potentiels, force des preuves. |
| **Evidence / Signal** | Classification d'une donnée : `VERIFIED`, `ESTIMATED`, `INFERRED`, `HYPOTHESIS`. Une donnée vérifiée conserve source, titre, URL, date, pays, métrique, valeur, `retrieved_at`. |
| **Market** | Contexte géo-économique : pays, langue, devise, secteurs, audience, problème, pouvoir d'achat, canaux de distribution, méthodes de paiement. |
| **Product Idea** | Concept de produit généré (titre, sous-titre, audience, problème, promesse, solution, format, prix estimé, difficulté, preuves, angle concurrentiel). |
| **Idea Version** | Version historique d'une idée (jamais écrasée silencieusement). |
| **Ebook / Guide digital** | Produit livrable initial (4 000–6 000 mots configurables) : titre, intro, problème, contexte local, méthode, chapitres, exemples, cas, checklist, conclusion, CTA. |
| **Asset** | Fichier généré : couverture, illustration, post réseau social, visuel pub, page de vente, script/storyboard vidéo. |
| **Quality Score** | Note 0–100 du contrôle qualité (complétude, vérification des faits, validité des citations, qualité de langue, cohérence, détection de doublons, cohérence marque). |
| **Crédit** | Unité de compte pay-as-you-go ; chaque opération consomme un coût configurable. |
| **Ledger** | Système comptable double-entrée des crédits (comptes, transactions, réservations). |
| **Reservation** | Blocage temporaire de crédits avant exécution d'une génération (`reserve → consume|release`). |
| **GenerationJob** | Job asynchrone d'une génération (id, status, progress, attempts, provider, cost, metadata). |
| **Workflow / WorkflowStep** | Chaîne orchestrée d'étapes (Research → Scoring → Ideas → Outline → Content → Assets → Marketing → QC → Packaging). |
| **Mobile Money** | Paiement via mobile (Wave, Orange Money, MTN), méthode prioritaire sur les marchés ciblés. |

## Technique

| Terme | Définition |
|---|---|
| **chi** | Routeur HTTP Go minimaliste et idiomatique. |
| **pgx/v5** | Driver PostgreSQL natif Go. |
| **sqlc** | Générateur de code Go type-safe à partir de SQL. |
| **goose** | Outil de migrations SQL versionnées. |
| **asynq** | File d'attente basée sur Redis pour Go. |
| **Ports & Adapters** | Hexagonal : le domaine définit des interfaces (ports), l'infra les implémente (adapters). |
| **Problem Details** | RFC 9457 : format d'erreur HTTP standardisé. |
| **Idempotence** | Rejouer une opération produit le même effet (évite la double consommation). |
| **CVA** | Class Variance Authority : gestion de variants de composants. |
| **TanStack Query** | Gestion d'état serveur (cache, invalidation, mutations). |
| **Tenant** | Entité isolant les données d'un utilisateur/organisation. |
| **RLS** | Row-Level Security PostgreSQL (filet de sécurité d'isolation). |

## IA & génération

| Terme | Définition |
|---|---|
| **LLMProvider** | Interface Go d'accès aux modèles de langage (OpenAI) — jamais hardcodée. |
| **ImageProvider / VideoProvider** | Interfaces d'accès aux modèles image (OpenAI `gpt-image-2`) et vidéo (HeyGen). |
| **ModelRouter** | Aiguillage tâche → (provider, modèle) : `gpt-5.6-terra` (recherche/contenu long), `gpt-5.6-luna` (idéation). |
| **Impeccable** | Skill open-source (github.com/pbakaus/impeccable) de vocabulaire design anti « AI slop », injecté dans le prompt de génération HTML. |
| **chromedp** | Pilote Go d'un Chrome headless, utilisé pour rendre le HTML et exporter PDF/PNG. |
| **Template HTML** | Gabarit de rendu des documents (thème « Emerald & Amber Ledger ») dans lequel on injecte le contenu structuré. |
| **Image-par-slide** | Assemblage PPTX où chaque slide est une image PNG pleine page (rendu HTML 16:9). |
| **Render Worker** | Worker asynq qui rend le HTML (chromedp) et exporte PDF/PPTX. |
