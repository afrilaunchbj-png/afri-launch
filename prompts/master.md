# PROMPT — Concevoir et construire un SaaS de création de produits digitaux pour les marchés africains

## 1. TON RÔLE

Tu es un **Product Architect, SaaS Engineer, AI Engineer et Product Strategist senior**, spécialisé dans :

* SaaS B2C/B2B2C
* marchés africains
* produits digitaux
* IA générative
* architectures distribuées
* workflows asynchrones
* paiements et systèmes de crédits
* UX mobile-first
* Next.js / TypeScript
* PostgreSQL / Prisma
* systèmes de queues et workers
* intégration de fournisseurs IA

Ton objectif n'est PAS de simplement produire beaucoup de code.

Ton objectif est de concevoir puis construire un **MVP réellement exploitable, maintenable, sécurisé et économiquement viable**.

Tu dois donc **challenger mes hypothèses lorsqu'elles sont mauvaises**, proposer de meilleures alternatives et expliquer brièvement tes décisions.

---

# 2. VISION DU PRODUIT

Je souhaite créer un SaaS permettant à un créateur, entrepreneur, consultant ou expert africain de transformer une opportunité de marché en **produit digital commercialisable**, avec un minimum de travail manuel.

Le workflow principal est :

```text
Recherche de marché
        ↓
Identification d'opportunités
        ↓
Sélection d'une niche
        ↓
Génération d'idées de produits
        ↓
Sélection / itération
        ↓
Création du produit
        ↓
Génération des assets marketing
        ↓
Contrôle qualité
        ↓
Packaging
        ↓
Export / publication
```

Le premier MVP doit principalement se concentrer sur :

1. recherche d'opportunités ;
2. génération et validation d'idées ;
3. création d'un ebook ou guide digital ;
4. génération des assets marketing ;
5. génération d'une page de vente ;
6. système de crédits.

La vidéo publicitaire doit être conçue comme une fonctionnalité extensible mais ne doit pas complexifier inutilement le cœur du MVP.

---

# 3. PRINCIPES STRATÉGIQUES

Avant de concevoir l'architecture ou d'écrire du code, analyse les hypothèses suivantes.

## 3.1 Ne considère PAS "l'Afrique" comme un marché homogène

Le système doit être conçu autour de :

```text
Country
Language
Currency
Audience
Sector
Problem
Purchasing Power
Distribution Channel
Payment Methods
```

Une opportunité doit donc pouvoir être associée à un ou plusieurs pays.

Exemple :

```text
Country: Nigeria
Language: English
Currency: NGN
Audience: Small business owners
Sector: Retail
Problem: Inventory management
```

et non simplement :

```text
Market: Africa
```

---

# 4. RECHERCHE DE MARCHÉ

Le système doit être capable d'identifier des opportunités à partir de données disponibles publiquement et/ou fournies par l'utilisateur.

## Sources potentielles

Selon disponibilité :

* rapports économiques ;
* données gouvernementales ;
* banques centrales ;
* organismes statistiques ;
* rapports d'entreprises ;
* rapports sectoriels ;
* Google Trends ou équivalent ;
* marketplaces ;
* réseaux sociaux ;
* forums et communautés ;
* données de recherche ;
* données fournies par l'utilisateur.

### RÈGLE CRITIQUE

Le système ne doit JAMAIS inventer une statistique.

Chaque donnée utilisée dans une recommandation doit être classifiée comme :

```text
VERIFIED
ESTIMATED
INFERRED
HYPOTHESIS
```

Pour une donnée vérifiée, conserver :

```text
source
title
url
publication_date
country
metric
value
retrieved_at
```

Si aucune source fiable n'est disponible, le système doit le dire explicitement.

---

# 5. SCORING DES OPPORTUNITÉS

Ne recommande pas simplement les niches avec le plus gros marché.

Construis un Opportunity Score basé notamment sur :

```text
Demand
Pain Level
Purchasing Power
Competition
Digital Product Suitability
Ease of Distribution
Local Relevance
Content Complexity
Potential Price
Potential Margin
Evidence Strength
```

Chaque critère doit avoir un score explicable.

Exemple :

```text
Opportunity Score: 82/100

Demand: 90
Pain: 85
Competition: 60
Purchasing Power: 70
Digital Product Fit: 95
Evidence Strength: 88
```

Le score doit être explicable à l'utilisateur.

---

# 6. SÉLECTION DE LA NICHE

L'utilisateur doit pouvoir filtrer les opportunités par :

* pays ;
* langue ;
* secteur ;
* niveau de difficulté ;
* taille du marché ;
* potentiel commercial ;
* concurrence ;
* budget nécessaire ;
* type de produit digital.

Il doit pouvoir sauvegarder une opportunité.

---

# 7. GÉNÉRATION D'IDÉES DE PRODUITS

Pour une opportunité sélectionnée, générer 5 à 10 concepts de produits.

Chaque concept doit contenir :

```text
title
subtitle
target_audience
problem
promise
solution
format
estimated_price
difficulty
market_evidence
why_now
competitive_angle
```

Les titres doivent être attractifs mais ne doivent pas utiliser de statistiques inventées.

Si une statistique est utilisée dans un titre, elle doit provenir d'une source identifiée.

---

# 8. SYSTÈME D'ITÉRATION

L'utilisateur doit pouvoir :

* aimer ;
* rejeter ;
* modifier ;
* demander une nouvelle proposition ;
* fusionner deux idées ;
* modifier l'audience ;
* modifier la promesse ;
* modifier le ton.

Le système doit conserver l'historique des versions.

Exemple :

```text
Idea v1
   ↓
User feedback
   ↓
Idea v2
   ↓
User feedback
   ↓
Idea v3
```

Ne jamais écraser silencieusement les versions précédentes.

---

# 9. CRÉATION DU PRODUIT DIGITAL

Une fois une idée validée, générer un plan détaillé.

Pour un ebook :

* titre ;
* sous-titre ;
* introduction ;
* problème ;
* contexte local ;
* méthode ;
* chapitres ;
* exemples ;
* études de cas ;
* checklist ;
* conclusion ;
* call-to-action.

Ne raisonne pas en nombre de pages.

Utilise plutôt une cible de :

```text
4 000 à 6 000 mots
```

pour un premier ebook, avec possibilité de configuration par l'utilisateur.

---

# 10. LOCALISATION

Le contenu doit être adapté au pays ciblé.

Prendre en compte :

* monnaie ;
* pouvoir d'achat ;
* habitudes commerciales ;
* exemples locaux ;
* terminologie ;
* réglementation lorsqu'elle est pertinente ;
* canaux de vente ;
* méthodes de paiement ;
* contexte culturel.

Ne jamais appliquer automatiquement un exemple nigérian à un utilisateur ciblant le Bénin, le Sénégal ou le Kenya.

---

# 11. GÉNÉRATION DES ASSETS

Après validation du contenu, lancer un workflow asynchrone.

Les assets peuvent inclure :

### Ebook

* couverture ;
* illustrations ;
* diagrammes ;
* images de chapitres.

### Marketing

* posts réseaux sociaux ;
* visuels publicitaires ;
* hooks ;
* captions ;
* emails ;
* page de vente.

### Vidéo

Générer éventuellement :

* script ;
* storyboard ;
* scènes ;
* texte ;
* voix ;
* montage ;
* export vertical 9:16 ;
* export horizontal 16:9.

Les fournisseurs IA doivent être abstraits derrière des interfaces.

Exemple :

```typescript
interface ImageGenerationProvider {
  generate(input: ImageGenerationInput): Promise<GeneratedAsset>
}
```

Ne couple jamais toute l'application à un seul fournisseur.

---

# 12. ARCHITECTURE IA

Ne construis PAS une architecture composée d'agents autonomes partout.

Utilise en priorité un workflow orchestré.

Exemple :

```text
ResearchWorkflow
      ↓
OpportunityScoringWorkflow
      ↓
ProductIdeaWorkflow
      ↓
OutlineWorkflow
      ↓
ContentWorkflow
      ↓
AssetWorkflow
      ↓
MarketingWorkflow
      ↓
QualityControlWorkflow
      ↓
PackagingWorkflow
```

Les agents peuvent être utilisés à l'intérieur d'une étape lorsqu'ils apportent une réelle valeur.

---

# 13. WORKFLOW ASYNCHRONE

Les générations longues ne doivent jamais être exécutées dans une requête HTTP classique.

Architecture souhaitée :

```text
Next.js
   ↓
API
   ↓
Workflow Orchestrator
   ↓
Queue
   ↓
Workers
   ├── Research Worker
   ├── LLM Worker
   ├── Image Worker
   ├── Video Worker
   ├── PDF Worker
   └── QC Worker
   ↓
Object Storage
```

Chaque job doit avoir :

```text
id
status
progress
attempts
started_at
completed_at
error
provider
cost
metadata
```

Les jobs doivent être :

* idempotents ;
* retryables ;
* observables ;
* annulables ;
* reprenables.

---

# 14. LIMITE DE 30 MINUTES

L'objectif est qu'un workflow complet puisse généralement être terminé en moins de 30 minutes.

Mais ne prétends jamais garantir 30 minutes si un fournisseur externe peut être lent.

Implémente plutôt :

```text
deadline
timeout
retry
fallback
partial completion
```

Si une génération vidéo échoue après que l'ebook et les assets sont terminés, le système doit conserver les résultats déjà produits.

---

# 15. SYSTÈME DE CRÉDITS

Le SaaS utilise un modèle pay-as-you-go.

Créer un véritable ledger de crédits.

Exemple :

```text
CreditAccount
CreditTransaction
CreditReservation
GenerationCost
```

Une génération doit suivre :

```text
available credits
      ↓
reserve
      ↓
execute
      ↓
success → consume
failure → release/refund
```

Le système doit être idempotent afin d'éviter les doubles consommations.

Chaque opération doit avoir un coût configurable.

Exemple :

```text
Niche Research       5 credits
Idea Generation      2 credits
Ebook Generation    20 credits
Image                3 credits
Video               15 credits
Sales Page           5 credits
```

Ces valeurs sont des exemples et doivent être configurables.

---

# 16. PAIEMENTS

Ne couple pas le système à Stripe.

Créer une abstraction :

```typescript
interface PaymentProvider {
  createPayment(...)
  verifyPayment(...)
  refundPayment(...)
}
```

Permettre plusieurs providers selon les pays.

Le système doit prendre en compte :

* devise locale ;
* Mobile Money ;
* cartes bancaires ;
* paiements internationaux ;
* webhooks ;
* idempotence ;
* remboursements.

---

# 17. STACK TECHNIQUE

Propose une architecture basée prioritairement sur :

### Frontend

* Next.js
* TypeScript
* Tailwind CSS
* composants UI accessibles
* responsive/mobile-first

### Backend

Next.js API / Route Handlers ou backend Node.js séparé si cela devient préférable.

Tu dois challenger ce choix et expliquer ta recommandation.

### Database

* PostgreSQL
* Prisma

### Queue

Choisir et justifier une solution adaptée.

Par exemple :

* Redis + BullMQ
* ou autre solution managée

### Storage

Utiliser un object storage compatible S3 ou équivalent.

### Authentication

Utiliser une solution moderne et maintenable.

Ne considère pas automatiquement une technologie comme obligatoire si elle est obsolète ou si une meilleure alternative existe.

### AI

Créer une couche d'abstraction permettant de changer facilement de modèle.

Exemple :

```text
LLMProvider
 ├── Provider A
 ├── Provider B
 └── Local Model
```

Le modèle LLM ne doit jamais être hardcodé dans toute l'application.

---

# 18. ARCHITECTURE DU PROJET

Propose une structure propre permettant d'évoluer.

Par exemple :

```text
apps/
  web/

packages/
  database/
  ai/
  billing/
  workflows/
  storage/
  shared/
  ui/

workers/
  generation/
```

Mais tu dois choisir l'architecture finale après analyse.

Explique pourquoi tu choisis :

* monorepo ou non ;
* backend intégré ou séparé ;
* workers séparés ou non ;
* queue ;
* storage ;
* observabilité.

---

# 19. MODÈLE DE DONNÉES

Conçois le modèle PostgreSQL complet.

Il doit au minimum couvrir :

```text
User
Organization
Project
Market
Opportunity
ProductIdea
IdeaVersion
Product
Chapter
ContentVersion
Asset
GenerationJob
Workflow
WorkflowStep
CreditAccount
CreditTransaction
Payment
Subscription / Plan
Notification
AuditLog
```

Tu dois ajouter les entités nécessaires.

Fournis ensuite :

* ERD logique ;
* relations ;
* indexes ;
* contraintes ;
* stratégies d'archivage ;
* stratégie de soft delete si pertinente.

---

# 20. UX/UI

Créer une expérience extrêmement simple.

Le produit doit être utilisable depuis un smartphone avec une connexion relativement lente.

Pages minimales :

### Marketing

```text
/
 /pricing
 /how-it-works
 /login
 /register
```

### Application

```text
/dashboard
/opportunities
/opportunities/:id
/ideas
/ideas/:id
/projects
/projects/:id
/projects/:id/content
/projects/:id/assets
/projects/:id/marketing
/projects/:id/export
/credits
/settings
```

L'interface doit afficher clairement :

* progression ;
* crédits consommés ;
* étapes terminées ;
* étapes en cours ;
* erreurs ;
* possibilité de relancer une étape.

---

# 21. MOBILE-FIRST

Optimiser particulièrement pour :

* mobile ;
* faible bande passante ;
* images optimisées ;
* lazy loading ;
* pagination ;
* compression ;
* reprise après perte de connexion ;
* temps de chargement réduit.

Ne pas envoyer inutilement de gros fichiers au navigateur.

---

# 22. NOTIFICATIONS

Implémenter :

* notification in-app ;
* email lorsque le workflow est terminé ;
* email en cas d'échec important.

Prévoir une architecture extensible pour :

```text
Email
Push
WhatsApp
SMS
```

sans implémenter tout cela dans le MVP si ce n'est pas nécessaire.

---

# 23. QUALITY CONTROL

Après chaque génération importante, effectuer des contrôles automatiques.

Exemples :

```text
Content completeness
Fact verification
Citation validity
Language quality
Consistency
Duplicate detection
Brand consistency
Marketing consistency
```

Créer un score :

```text
Quality Score: 87/100
```

Afficher les problèmes détectés à l'utilisateur.

---

# 24. SÉCURITÉ

Le système doit prendre en compte :

* authentication ;
* authorization ;
* tenant isolation ;
* rate limiting ;
* validation des inputs ;
* protection des webhooks ;
* secrets management ;
* SQL injection ;
* XSS ;
* CSRF lorsque pertinent ;
* upload sécurisé ;
* signed URLs ;
* contrôle des coûts IA ;
* protection contre les abus ;
* audit logs.

Un utilisateur ne doit jamais pouvoir accéder aux données d'un autre utilisateur simplement en modifiant un ID dans l'URL.

---

# 25. OBSERVABILITÉ

Prévoir :

```text
structured logging
metrics
tracing
job monitoring
AI generation cost tracking
error tracking
```

Pour chaque génération IA, conserver au minimum :

```text
provider
model
input tokens
output tokens
latency
estimated cost
status
```

---

# 26. COÛTS

Construis une estimation du coût de génération d'un produit complet.

Par exemple :

```text
Research
LLM
Images
Video
Storage
Email
Infrastructure
```

Puis estime :

```text
Cost per generated product
Gross margin
Break-even
```

Le système de crédits doit être conçu à partir de ces coûts.

---

# 27. MVP VS FUTURE

Sépare clairement :

## MVP

Fonctionnalités indispensables pour tester le marché.

## V2

Fonctionnalités utiles mais non indispensables.

## Future

Fonctionnalités ambitieuses.

Ne surcharge pas le MVP.

---

# 28. VALIDATION BUSINESS

Avant de coder, réponds à ces questions :

1. Qui est le premier utilisateur cible ?
2. Quel problème paie-t-il réellement pour résoudre ?
3. Pourquoi utiliserait-il ce SaaS plutôt que ChatGPT + Canva ?
4. Quelle est la proposition de valeur unique ?
5. Quel est le produit digital minimal généré ?
6. Quel prix peut être demandé ?
7. Quel est le coût réel de génération ?
8. Quelle marge est possible ?
9. Quel canal d'acquisition semble réaliste ?
10. Quel pays devrait être ciblé en premier ?
11. Pourquoi ce pays ?
12. Quel est le plus gros risque du business ?
13. Quelle hypothèse doit être validée avant de développer ?

Si une hypothèse du concept est faible, dis-le explicitement.

---

# 29. CONCURRENCE

Analyse le positionnement potentiel face à des catégories de produits comme :

* ChatGPT ;
* Claude ;
* Canva ;
* Jasper ;
* Copy.ai ;
* outils de création d'ebooks ;
* outils de génération vidéo ;
* outils de recherche marketing.

Ne cherche pas uniquement à copier leurs fonctionnalités.

Identifie :

```text
What they do better
What they do worse
What African users need that they don't solve
What can become our moat
```

---

# 30. MOAT

Propose des avantages compétitifs défendables.

Par exemple :

* datasets locaux ;
* connaissance des marchés africains ;
* templates localisés ;
* pricing local ;
* payment integrations ;
* distribution ;
* workflows spécialisés ;
* données propriétaires issues des utilisateurs ;
* verticalisation.

Ne considère pas simplement « utiliser l'IA » comme un avantage compétitif.

---

# 31. LIVRABLES

Tu dois travailler en plusieurs phases.

## PHASE 1 — Product Strategy

Fournir :

* critique du concept ;
* hypothèses ;
* risques ;
* cible ;
* proposition de valeur ;
* MVP ;
* roadmap.

**Ne génère aucun code à cette phase.**

---

## PHASE 2 — Architecture

Fournir :

* architecture globale ;
* diagramme des composants ;
* flux de données ;
* architecture IA ;
* architecture queue/workers ;
* architecture paiement ;
* architecture crédits ;
* stratégie storage ;
* stratégie observabilité.

Utilise des diagrammes Mermaid lorsque pertinent.

---

## PHASE 3 — Data Model

Fournir :

* schéma Prisma complet ;
* ERD ;
* indexes ;
* contraintes ;
* stratégie migrations.

---

## PHASE 4 — API

Définir les API :

```text
Authentication
Markets
Opportunities
Ideas
Projects
Generation
Assets
Credits
Payments
Notifications
```

Pour chaque endpoint :

```text
method
path
authentication
request
response
errors
authorization
idempotency
```

---

## PHASE 5 — UX/UI

Définir :

* architecture des pages ;
* navigation ;
* composants ;
* états loading ;
* états empty ;
* erreurs ;
* responsive behavior ;
* mobile UX.

---

## PHASE 6 — IMPLEMENTATION

Construire le projet.

Le code doit être :

* TypeScript strict ;
* propre ;
* modulaire ;
* testable ;
* sécurisé ;
* maintenable ;
* documenté.

Ne génère pas artificiellement des fichiers inutiles.

---

# 32. TESTS

Inclure :

### Unit tests

Pour :

* scoring ;
* credits ;
* billing ;
* authorization ;
* workflow logic.

### Integration tests

Pour :

* database ;
* API ;
* queues ;
* webhooks.

### E2E

Tester au minimum :

```text
Register
Create project
Research opportunity
Select idea
Generate ebook
Generate assets
Consume credits
Complete workflow
Export product
```

---

# 33. DÉPLOIEMENT

Fournir une stratégie de déploiement réaliste.

Séparer :

```text
Frontend
API
Workers
Database
Redis
Object Storage
Monitoring
```

Expliquer ce qui peut être hébergé sur Vercel et ce qui devrait être déporté vers une infrastructure adaptée aux workloads longs.

Ne force pas Vercel à héberger des workloads qui ne lui conviennent pas.

---

# 34. ENVIRONMENT VARIABLES

Fournir un `.env.example` complet.

Ne jamais mettre de vraies clés API dans le code.

---

# 35. DOCUMENTATION

Créer :

```text
README.md
ARCHITECTURE.md
DEVELOPMENT.md
DEPLOYMENT.md
DATABASE.md
AI.md
BILLING.md
SECURITY.md
```

---

# 36. RÈGLES DE QUALITÉ

Respecte impérativement ces règles :

1. Ne jamais inventer une statistique.
2. Ne jamais inventer une API.
3. Ne jamais inventer une fonctionnalité d'un fournisseur.
4. Vérifier la documentation officielle des services externes lorsque nécessaire.
5. Si une technologie recommandée est obsolète, le signaler.
6. Préférer une architecture simple à une architecture inutilement complexe.
7. Ne pas utiliser des agents IA simplement parce que le mot « agent » est populaire.
8. Chaque workflow long doit être asynchrone.
9. Chaque génération doit être observable.
10. Les crédits doivent être financièrement contrôlables.
11. Toutes les opérations sensibles doivent être idempotentes.
12. La sécurité doit être intégrée dès la conception.
13. Le produit doit être mobile-first.
14. Le système doit pouvoir changer de fournisseur IA.
15. Le système doit pouvoir changer de fournisseur de paiement.
16. Le système doit pouvoir évoluer vers plusieurs pays africains.
17. Les données utilisateur doivent être isolées par tenant.
18. Une erreur sur une étape ne doit pas détruire les résultats déjà générés.
19. Le système doit privilégier les preuves aux suppositions.
20. Lorsque plusieurs architectures sont possibles, comparer les options avant de choisir.

---

# 37. FORMAT FINAL DE TA RÉPONSE

Réponds exactement dans cet ordre :

## 1. Executive Summary

## 2. Critique du business model

## 3. Target User

## 4. Proposition de valeur

## 5. MVP recommandé

## 6. Risques et hypothèses à valider

## 7. Analyse concurrentielle

## 8. Architecture technique

## 9. Diagrammes Mermaid

## 10. Architecture IA

## 11. Architecture des workflows

## 12. Architecture des queues/workers

## 13. Architecture des paiements

## 14. Architecture des crédits

## 15. Modèle de données

## 16. API contract

## 17. Architecture frontend

## 18. UX/UI

## 19. Structure complète du repository

## 20. Code

## 21. Tests

## 22. Sécurité

## 23. Observabilité

## 24. Estimation des coûts

## 25. Stratégie de déploiement

## 26. Documentation

## 27. Roadmap MVP → V2 → V3

---

# 38. IMPORTANT — ORDRE DE PRIORITÉ

Si tu dois choisir entre :

```text
Feature richness
Architecture quality
Product viability
```

priorise :

```text
Product viability
>
Architecture quality
>
Feature richness
```

Je préfère un MVP avec 20 fonctionnalités extrêmement bien conçues plutôt qu'une plateforme de 100 fonctionnalités fragile.

Avant de générer plusieurs milliers de lignes de code, identifie d'abord les décisions qui peuvent remettre en cause l'architecture.

Lorsque tu identifies une hypothèse discutable, **challenge-la explicitement au lieu de simplement l'accepter**.
