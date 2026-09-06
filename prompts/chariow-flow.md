# Mission : intégrer Chariow comme provider de commerce et de paiement de mon SaaS

Tu travailles sur mon SaaS de création et de vente de produits digitaux destiné principalement au marché africain.

Je veux maintenant intégrer **Chariow** comme plateforme de commerce/paiement.

## 1. Objectif

Permettre à un utilisateur de mon SaaS de connecter **sa boutique Chariow existante** à mon application afin que mon SaaS puisse, depuis son interface, exploiter l'API Chariow pour :

* créer/gérer ses produits digitaux ;
* générer des checkouts ;
* récupérer les ventes ;
* synchroniser les données commerciales ;
* exploiter les webhooks Chariow ;
* rattacher chaque vente à l'utilisateur et au produit correspondant dans mon SaaS ;
* éventuellement déclencher des automatisations via n8n.

### Architecture métier souhaitée

Je ne veux PAS, pour le moment, créer un compte Chariow central pour tous mes utilisateurs.

Le modèle cible est :

```text
Utilisateur SaaS A
    │
    └── sa boutique Chariow

Utilisateur SaaS B
    │
    └── sa boutique Chariow

Utilisateur SaaS C
    │
    └── sa boutique Chariow
```

Chaque utilisateur connecte donc sa propre boutique Chariow à mon SaaS.

Pour la première version, nous utiliserons le mécanisme officiellement disponible permettant à l'utilisateur de fournir/connecter une **API key Chariow**.

IMPORTANT :

* Ne jamais inventer un endpoint Chariow.
* Ne jamais supposer qu'un OAuth Chariow pour applications tierces existe si la documentation officielle ne le confirme pas.
* Ne jamais supposer qu'une API permet de créer automatiquement un compte Chariow.
* Ne jamais supposer qu'une API permet de créer automatiquement une boutique Chariow.
* Toute fonctionnalité non confirmée par la documentation officielle doit être explicitement marquée comme "non supportée / à vérifier".
* L'intégration doit être conçue pour pouvoir migrer ultérieurement vers OAuth/Partner API si Chariow le permet.

---

# 2. Documentation officielle à utiliser

Avant toute implémentation, consulte la documentation officielle Chariow actuelle.

Sources prioritaires :

* https://chariow.dev/
* https://chariow.dev/en/introduction/quickstart
* https://chariow.dev/api-reference/introduction
* https://chariow.dev/api-reference/checkout/init-checkout
* https://help.chariow.com/fr/articles/222-gerer-vos-cles-api-sur-chariow
* https://help.chariow.com/fr/articles/242-integrer-chariow-a-n8n

Si les URLs ont changé, recherche la documentation officielle actuelle.

Ne te base pas sur des articles tiers lorsqu'une information peut être obtenue depuis la documentation officielle.

---

# 3. Première étape obligatoire : audit du projet

Avant de modifier le code :

1. Analyse entièrement la structure du projet.
2. Identifie :

   * frontend ;
   * backend ;
   * ORM ;
   * base de données ;
   * système d'authentification ;
   * système de configuration ;
   * gestion des secrets ;
   * système de queues/jobs ;
   * système de webhooks existant ;
   * système d'intégrations externes existant ;
   * système de logs ;
   * système de permissions ;
   * tests ;
   * conventions existantes.
3. Lis les fichiers de conventions du projet, notamment :

   * `AGENTS.md`
   * `CLAUDE.md`
   * `CONVENTIONS.md`
   * `README.md`
   * tous les fichiers présents dans les éventuels dossiers `skills/`, `.agents/`, `docs/`, etc.
4. Ne réécris pas l'architecture existante sans raison.
5. Réutilise les abstractions déjà présentes dans le projet.

Avant de coder, produis un court diagnostic de l'architecture existante et explique où l'intégration Chariow doit être placée.

---

# 4. Architecture d'intégration souhaitée

Je veux une architecture découplée.

Ne disperse pas des appels HTTP Chariow partout dans le code.

Créer une couche dédiée :

```text
ChariowIntegration
    │
    ├── ChariowClient
    ├── ChariowProductsService
    ├── ChariowCheckoutService
    ├── ChariowSalesService
    ├── ChariowWebhookService
    └── ChariowSyncService
```

Adapte cette structure à l'architecture réelle du projet.

Le reste du SaaS doit dépendre d'une abstraction interne et non directement de détails HTTP Chariow.

Exemple conceptuel :

```text
Application
    ↓
CommerceProvider
    ↓
ChariowProvider
    ↓
Chariow API
```

L'objectif est de pouvoir ajouter plus tard :

```text
ChariowProvider
StripeProvider
PaystackProvider
...
```

sans réécrire tout le domaine métier.

---

# 5. Modèle de données

Analyse le schéma actuel puis ajoute les entités nécessaires.

Il faut pouvoir représenter au minimum :

## CommerceIntegration

Une intégration entre un utilisateur SaaS et Chariow.

Exemple conceptuel :

```text
CommerceIntegration
-------------------
id
userId
provider
status
credentials
externalAccountId
externalStoreId
storeName
storeUrl
connectedAt
lastSyncAt
createdAt
updatedAt
```

`provider` doit pouvoir contenir :

```text
CHARIOW
```

Prévoir l'architecture pour d'autres providers.

---

# 6. Gestion de la clé API

La clé API Chariow est une donnée secrète.

NE JAMAIS :

* la stocker en clair si le projet dispose déjà d'un mécanisme de chiffrement des secrets ;
* l'envoyer au frontend après connexion ;
* la logger ;
* l'inclure dans les erreurs ;
* la retourner dans les endpoints publics ;
* l'inclure dans les analytics ;
* la mettre dans les URLs ;
* la committer dans Git.

Si le projet possède déjà un système de chiffrement des secrets, utilise-le.

Sinon, implémente une solution robuste de chiffrement côté backend.

Le frontend ne doit recevoir que des informations non sensibles :

```json
{
  "provider": "CHARIOW",
  "status": "CONNECTED",
  "storeName": "...",
  "storeUrl": "...",
  "connectedAt": "..."
}
```

Jamais :

```json
{
  "apiKey": "..."
}
```

---

# 7. Configuration / variables d'environnement

Analyse d'abord les conventions du projet.

N'ajoute pas de variables inutiles.

Si nécessaire, prévoir uniquement des variables du type :

```env
CHARIOW_API_BASE_URL=
CHARIOW_WEBHOOK_SECRET=
```

IMPORTANT :

La clé API Chariow d'un utilisateur ne doit PAS être une variable d'environnement globale.

Elle appartient à l'utilisateur et doit être stockée de manière sécurisée en base.

Exemple :

```text
ENV
 ├── CHARIOW_API_BASE_URL
 └── autres configurations globales

DATABASE
 └── encrypted Chariow API key par utilisateur
```

Ne mets donc surtout pas :

```env
CHARIOW_API_KEY=...
```

comme clé globale de tous les utilisateurs.

---

# 8. Connexion de Chariow

Créer une fonctionnalité :

```text
Settings
  → Integrations
      → Chariow
          → Connect
```

UX souhaitée :

### État non connecté

Afficher :

```text
Chariow
Connectez votre boutique Chariow
[Connecter Chariow]
```

### Connexion

Afficher des instructions simples :

```text
1. Connectez-vous à Chariow
2. Ouvrez vos paramètres
3. Créez une clé API
4. Copiez-la
5. Collez-la ici
```

Puis :

```text
[API Key]
[Connecter]
```

Le backend doit immédiatement tester la clé auprès de Chariow.

NE PAS considérer l'intégration comme connectée simplement parce que la chaîne fournie ressemble à une API key.

Il faut vérifier réellement les credentials auprès de Chariow.

---

# 9. Test de connexion

Créer un mécanisme :

```text
POST /integrations/chariow/connect
```

ou l'équivalent adapté à l'architecture existante.

Workflow :

```text
Frontend
    ↓
API SaaS
    ↓
validate Chariow credentials
    ↓
Chariow API
    ↓
success
    ↓
encrypt credentials
    ↓
save integration
```

En cas d'erreur :

```text
Chariow credentials invalid
```

ou un message utilisateur compréhensible.

Ne jamais exposer les détails techniques ou secrets dans l'erreur frontend.

---

# 10. Client HTTP Chariow

Créer un client HTTP dédié.

Responsabilités :

* authentification ;
* headers ;
* base URL ;
* timeout ;
* retry contrôlé ;
* gestion des erreurs ;
* rate limiting si nécessaire ;
* logging sécurisé ;
* parsing des réponses.

Exemple conceptuel :

```typescript
class ChariowClient {
  get(...)
  post(...)
  patch(...)
  delete(...)
}
```

Mais adapte cela aux conventions du projet.

Tous les appels doivent avoir :

* timeout ;
* gestion propre des erreurs ;
* logs sans secrets ;
* correlation/request ID lorsque disponible.

---

# 11. Produits

Implémenter uniquement les opérations réellement disponibles dans l'API Chariow actuelle.

Créer une abstraction interne :

```text
createProduct()
updateProduct()
getProduct()
listProducts()
deleteProduct()
```

Si certaines opérations ne sont pas supportées par Chariow :

* ne pas les simuler ;
* documenter clairement la limitation ;
* proposer éventuellement un fallback.

---

# 12. Publication d'un produit généré par mon SaaS

Mon SaaS génère des produits digitaux.

Je veux pouvoir avoir un workflow du type :

```text
SaaS Product
     │
     ▼
Validate product
     │
     ▼
Connect Chariow
     │
     ▼
Create Chariow product
     │
     ▼
Store external product ID
     │
     ▼
Store checkout information
```

Dans ma base, prévoir une relation permettant de savoir :

```text
SaaS Product
      ↓
Chariow Product
      ↓
externalProductId
```

Un produit ne doit pas être recréé à chaque synchronisation.

Il faut pouvoir faire :

```text
SaaS Product
    │
    ├── chariowProductId
    └── syncStatus
```

ou créer une table de mapping si cela correspond mieux à l'architecture.

---

# 13. Checkout

Implémenter l'utilisation de l'API Chariow permettant de générer un checkout.

Le workflow cible :

```text
Client final
    ↓
SaaS landing page
    ↓
Buy
    ↓
Backend
    ↓
Chariow checkout API
    ↓
Checkout URL
    ↓
Client
```

Le frontend ne doit pas contenir la clé API Chariow.

Le backend doit générer le checkout.

---

# 14. Custom metadata

Si l'API Chariow actuelle permet les `custom_metadata` sur les checkouts, les utiliser pour relier une transaction Chariow à mon SaaS.

Exemple conceptuel :

```json
{
  "custom_metadata": {
    "saas_user_id": "...",
    "saas_product_id": "...",
    "checkout_id": "..."
  }
}
```

N'utilise que les champs réellement supportés par Chariow.

Objectif :

```text
Chariow sale
     ↓
metadata
     ↓
SaaS user
     ↓
SaaS product
```

C'est essentiel pour le multi-tenant.

---

# 15. Webhooks

Créer un endpoint sécurisé pour recevoir les webhooks Chariow.

Exemple conceptuel :

```text
POST /webhooks/chariow
```

IMPORTANT :

Ne fais pas confiance aveuglément au payload.

Implémenter, selon les mécanismes réellement fournis par Chariow :

* vérification de signature ;
* secret ;
* validation du payload ;
* validation de l'événement ;
* idempotence ;
* protection contre les doublons ;
* logs ;
* gestion des erreurs ;
* retry.

Si Chariow fournit une signature webhook, utilise-la.

Si Chariow ne fournit pas de mécanisme de signature dans la version actuelle de l'API, documente précisément cette limitation au lieu d'en inventer un.

---

# 16. Événements de vente

Prendre en charge les événements réellement documentés par Chariow.

Notamment, si toujours disponibles :

```text
successful.sale
abandoned.sale
failed.sale
refunded.sale
```

Architecture :

```text
Chariow
   │
   │ webhook
   ▼
Webhook Controller
   │
   ▼
Webhook Processor
   │
   ├── validate
   ├── deduplicate
   ├── map user
   ├── map product
   └── persist event
```

Le traitement lourd ne doit pas nécessairement être effectué directement dans la requête HTTP.

Si le projet possède Redis/BullMQ ou un système de queue existant, utilise-le.

---

# 17. Idempotence

Les webhooks peuvent être reçus plusieurs fois.

Créer une stratégie d'idempotence.

Par exemple :

```text
WebhookEvent
----------------
id
provider
externalEventId
eventType
payload
processedAt
status
createdAt
```

Ajouter une contrainte unique appropriée sur :

```text
provider + externalEventId
```

ou utiliser l'identifiant unique réellement fourni par Chariow.

Ne jamais enregistrer deux fois une même vente.

---

# 18. Synchronisation des ventes

En plus des webhooks, prévoir une synchronisation API.

Pourquoi ?

Parce qu'un webhook peut :

* être perdu ;
* être retardé ;
* échouer ;
* être mal traité.

Prévoir :

```text
Sync Chariow
    ↓
fetch sales
    ↓
compare
    ↓
upsert
```

Le job doit être idempotent.

Si le projet utilise déjà un système de jobs :

```text
ChariowSyncJob
```

Sinon, implémenter la solution cohérente avec l'architecture existante.

---

# 19. n8n

Je veux pouvoir utiliser n8n, mais **n8n ne doit pas devenir le cœur de l'intégration métier**.

Architecture cible :

```text
SaaS Backend
      │
      ├── Chariow API
      │
      └── n8n
```

Le backend reste responsable de :

* utilisateurs ;
* permissions ;
* produits ;
* commandes ;
* intégrations ;
* synchronisation ;
* sécurité ;
* cohérence des données.

n8n peut être utilisé pour :

* email ;
* notifications ;
* CRM ;
* Google Sheets ;
* marketing ;
* automatisations ;
* workflows secondaires.

Si une automatisation critique dépend de n8n, prévoir une stratégie de retry/failure.

---

# 20. Architecture n8n

Si n8n est déjà présent dans le projet, analyse sa configuration avant de modifier quoi que ce soit.

Si aucune intégration n8n n'existe :

NE PAS installer automatiquement une architecture n8n complexe juste pour Chariow.

Préparer plutôt des webhooks/events internes permettant de connecter n8n proprement.

Exemple :

```text
Chariow sale
    ↓
SaaS webhook
    ↓
SaaS event
    ↓
n8n webhook
    ↓
Automation
```

---

# 21. Multi-tenancy

C'est critique.

Chaque appel Chariow doit utiliser les credentials du bon utilisateur.

NE JAMAIS faire :

```text
global Chariow API key
```

Faire :

```text
authenticated SaaS user
        ↓
CommerceIntegration
        ↓
encrypted credentials
        ↓
ChariowClient
        ↓
Chariow
```

Toutes les routes doivent vérifier :

```text
currentUser.id === integration.userId
```

ou utiliser le mécanisme d'autorisation existant.

Un utilisateur A ne doit absolument jamais pouvoir :

* récupérer la clé de B ;
* accéder à la boutique B ;
* récupérer les produits de B ;
* récupérer les ventes de B ;
* utiliser les credentials de B.

---

# 22. Permissions

Utiliser le système de permissions existant.

Prévoir au minimum :

```text
integration:read
integration:connect
integration:disconnect
integration:sync
```

Si le projet n'utilise pas ce modèle, adapte-le au système existant.

---

# 23. Déconnexion

Implémenter :

```text
Disconnect Chariow
```

La déconnexion doit :

1. invalider/supprimer les credentials chiffrés ;
2. conserver éventuellement les données historiques nécessaires ;
3. conserver les mappings historiques ;
4. empêcher les nouveaux appels API ;
5. arrêter les jobs de synchronisation associés ;
6. supprimer/désactiver les webhooks si l'API Chariow le permet ;
7. ne pas supprimer aveuglément les données commerciales locales.

---

# 24. Reconnexion

Prévoir :

```text
CONNECTED
DISCONNECTED
ERROR
RECONNECT_REQUIRED
```

Si une clé devient invalide :

```text
Chariow API
    ↓
401/403
    ↓
integration status = RECONNECT_REQUIRED
    ↓
notify user
```

Ne pas supprimer automatiquement les données historiques.

---

# 25. Observabilité

Tous les appels doivent être observables sans exposer les secrets.

Logger :

```text
provider=chariow
operation=create_product
userId=...
externalProductId=...
duration=...
status=success/error
```

NE JAMAIS logger :

```text
Authorization
Bearer token
API key
webhook secret
checkout secret
```

---

# 26. Tests

Ajouter des tests unitaires et d'intégration.

Minimum :

### Credentials

* connexion valide ;
* clé invalide ;
* Chariow indisponible ;
* timeout.

### Multi-tenancy

* utilisateur A ne peut pas accéder à l'intégration B ;
* utilisateur A ne peut pas utiliser les credentials B.

### Products

* création ;
* mise à jour ;
* synchronisation ;
* erreurs API.

### Checkout

* création ;
* metadata ;
* erreur API.

### Webhooks

* webhook valide ;
* webhook invalide ;
* signature invalide si applicable ;
* doublon ;
* événement inconnu ;
* payload invalide ;
* traitement réussi ;
* traitement échoué.

### Sync

* vente nouvelle ;
* vente existante ;
* modification ;
* remboursement ;
* répétition du job.

---

# 27. Gestion des erreurs

Créer des erreurs métier propres.

Exemple :

```text
ChariowIntegrationError
ChariowAuthenticationError
ChariowRateLimitError
ChariowValidationError
ChariowUnavailableError
ChariowWebhookError
```

Adapter aux conventions existantes.

Le frontend doit recevoir des erreurs compréhensibles.

Ne jamais retourner directement les réponses brutes de Chariow contenant potentiellement des données sensibles.

---

# 28. Rate limits et retry

Avant d'implémenter les retries, vérifie les limites officielles de Chariow.

Ne pas mettre un retry aveugle.

Utiliser une stratégie adaptée :

```text
429 → backoff
5xx → retry limité
4xx → généralement pas de retry
401/403 → reconnect required
```

Respecter les `Retry-After` headers s'ils existent.

---

# 29. Documentation interne

Créer une documentation interne expliquant :

```text
docs/integrations/chariow.md
```

Elle doit expliquer :

* architecture ;
* configuration ;
* credentials ;
* connexion ;
* produits ;
* checkout ;
* webhooks ;
* synchronisation ;
* n8n ;
* erreurs ;
* sécurité ;
* limitations connues ;
* comment ajouter un autre commerce provider.

Créer également une section :

## Limitations Chariow

Y documenter explicitement ce que l'API publique actuelle ne permet pas.

Notamment :

* création automatique d'un compte Chariow si non supportée ;
* création automatique d'une boutique si non supportée ;
* OAuth tierce partie si non disponible ;
* marketplace/split settlement si non disponible.

Ne rien inventer.

---

# 30. UX finale souhaitée

Dans mon SaaS :

```text
Settings
│
└── Integrations
    │
    └── Chariow
         │
         ├── Not connected
         │      └── [Connect]
         │
         └── Connected
                │
                ├── Store name
                ├── Store URL
                ├── Status
                ├── Last synchronization
                │
                ├── [Sync now]
                └── [Disconnect]
```

Lorsqu'un produit est publié :

```text
Product
   │
   ├── Draft
   ├── Generated
   ├── Published
   │
   └── Chariow
        ├── externalProductId
        ├── checkoutUrl
        └── syncStatus
```

---

# 31. Important : ne pas sur-engineer

Je veux une implémentation production-ready mais pragmatique.

Ne crée pas :

* un microservice Chariow séparé ;
* une architecture event-driven complexe ;
* une couche d'abstraction inutilement énorme ;
* des providers fictifs ;
* un système OAuth fictif ;
* des endpoints qui n'existent pas.

Utilise l'architecture actuelle du projet.

---

# 32. Workflow d'implémentation obligatoire

Procède dans cet ordre :

### Phase 1 — Analyse

Analyse le repository et la documentation officielle Chariow.

Ne modifie rien.

Donne-moi :

1. architecture actuelle ;
2. fichiers concernés ;
3. modèle de données actuel ;
4. système d'intégrations existant ;
5. système de secrets ;
6. système de jobs ;
7. système de webhooks ;
8. endpoints Chariow réellement disponibles ;
9. limitations identifiées ;
10. plan d'implémentation.

### Phase 2 — Design

Propose :

* modifications DB ;
* API backend ;
* services ;
* frontend ;
* sécurité ;
* webhooks ;
* jobs ;
* tests.

Attends ma validation si le projet est suffisamment complexe pour le justifier.

### Phase 3 — Implémentation

Implémente progressivement.

Après chaque grosse étape :

```text
typecheck
lint
unit tests
integration tests
build
```

Corrige les régressions.

### Phase 4 — Validation

Vérifie :

* aucune clé secrète dans les logs ;
* aucun credential côté frontend ;
* isolation multi-tenant ;
* idempotence ;
* gestion des erreurs ;
* tests ;
* migrations ;
* documentation.

---

# 33. Critère de réussite

À la fin, un utilisateur doit pouvoir faire :

```text
1. Ouvrir Settings
2. Cliquer "Connect Chariow"
3. Entrer sa clé API
4. Le SaaS vérifie la clé
5. La clé est chiffrée et stockée
6. La boutique est détectée
7. Le statut passe à CONNECTED
8. L'utilisateur crée un produit dans mon SaaS
9. Le SaaS publie le produit sur Chariow
10. Le SaaS récupère son identifiant externe
11. Le SaaS génère un checkout
12. Un client achète
13. Chariow envoie un webhook
14. Mon SaaS reçoit et valide le webhook
15. La vente est associée au bon utilisateur
16. La vente est associée au bon produit
17. Le système est idempotent
18. Les données peuvent être synchronisées ultérieurement
```

---

# 34. Règle absolue

Tu es un agent de développement.

Tu dois privilégier :

```text
Documentation officielle
        >
Code existant du projet
        >
Hypothèse
```

Si une capacité Chariow n'est pas documentée :

**NE L'INVENTE PAS.**

Signale :

```text
NOT CONFIRMED
```

et propose une architecture permettant de l'ajouter plus tard.

L'objectif n'est pas seulement de "faire fonctionner Chariow".

L'objectif est de construire une **intégration Chariow multi-tenant, sécurisée, maintenable et extensible**, qui pourra devenir plus tard une abstraction de commerce pour l'ensemble de mon SaaS.
