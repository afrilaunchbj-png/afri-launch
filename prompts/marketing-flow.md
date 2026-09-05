# Mission — Construire le système d'intégration publicitaire multi-tenant

Tu es un Software Architect / Senior Backend Engineer spécialisé en SaaS multi-tenant, OAuth 2.0, APIs publicitaires, systèmes distribués et intégrations tierces.

Je veux implémenter dans ce projet un système permettant à chaque organisation/utilisateur du SaaS de connecter ses propres comptes publicitaires et permettant ensuite à notre backend de :

* découvrir les comptes publicitaires accessibles ;
* laisser l'utilisateur sélectionner le compte à utiliser ;
* créer des campagnes ;
* créer des ad sets / ad groups ;
* uploader des creatives ;
* créer des publicités ;
* lancer / mettre en pause des campagnes ;
* modifier certains paramètres ;
* récupérer les statistiques ;
* suivre les conversions ;
* calculer les performances ;
* préparer ultérieurement une couche d'optimisation IA.

Les plateformes initiales sont :

1. Meta Ads / Facebook / Instagram
2. Google Ads
3. TikTok Ads

L'architecture doit être conçue pour pouvoir ajouter d'autres plateformes plus tard.

---

# 1. RÈGLE FONDAMENTALE

Le SaaS est MULTI-TENANT.

Chaque client possède ses propres comptes publicitaires.

Nous ne devons JAMAIS :

* utiliser les credentials d'un client pour un autre client ;
* exposer un access token au frontend ;
* stocker des secrets en clair ;
* demander le mot de passe du client ;
* hard-coder un account ID ;
* mettre la logique Meta/Google/TikTok directement dans le domaine métier.

Le système doit utiliser OAuth lorsque la plateforme le permet.

Le modèle mental est :

```text
Notre SaaS
    │
    ├── Application Meta
    ├── Application Google
    └── Application TikTok
             │
             ▼
       OAuth utilisateur
             │
             ▼
       Compte publicitaire
             │
             ▼
   AdPlatformConnection
             │
             ▼
       Notre backend
```

Les credentials de notre application sont globaux au SaaS.

Les tokens et identifiants de comptes publicitaires sont propres à chaque connexion/tenant.

---

# 2. AVANT DE CODER

NE COMMENCE PAS PAR ÉCRIRE DU CODE.

Commence par auditer le repository.

Inspecte :

* architecture ;
* framework backend ;
* ORM ;
* modèle User ;
* modèle Organization / Workspace / Tenant ;
* authentification ;
* autorisation ;
* système de queue ;
* Redis ;
* stockage objet ;
* configuration ;
* système de secrets ;
* logging ;
* tests ;
* conventions de nommage ;
* structure des modules ;
* éventuel système d'intégrations déjà existant.

Si une architecture multi-tenant existe déjà, réutilise-la.

Ne crée pas un deuxième système de tenant.

Si un système d'intégrations existe déjà, étends-le proprement.

---

# 3. ARCHITECTURE CIBLE

Je veux une architecture provider-agnostic.

Créer une abstraction conceptuelle :

```typescript
interface AdPlatformProvider {
  getAuthorizationUrl(input: OAuthAuthorizationInput): Promise<string>;

  exchangeAuthorizationCode(
    input: OAuthCallbackInput
  ): Promise<OAuthTokenResult>;

  refreshAccessToken?(
    connection: AdPlatformConnection
  ): Promise<OAuthTokenResult>;

  revokeAccess?(
    connection: AdPlatformConnection
  ): Promise<void>;

  getAdAccounts(
    connection: AdPlatformConnection
  ): Promise<AdAccount[]>;

  getCampaigns(
    connection: AdPlatformConnection,
    input: CampaignQuery
  ): Promise<Campaign[]>;

  createCampaign(
    connection: AdPlatformConnection,
    input: CreateCampaignInput
  ): Promise<Campaign>;

  updateCampaign(
    connection: AdPlatformConnection,
    input: UpdateCampaignInput
  ): Promise<Campaign>;

  pauseCampaign(
    connection: AdPlatformConnection,
    campaignId: string
  ): Promise<void>;

  resumeCampaign(
    connection: AdPlatformConnection,
    campaignId: string
  ): Promise<void>;

  getInsights(
    connection: AdPlatformConnection,
    input: InsightsQuery
  ): Promise<CampaignInsights[]>;
}
```

Les méthodes peuvent être adaptées aux capacités réelles de chaque provider.

IMPORTANT :

Toutes les plateformes ne supportent pas exactement les mêmes fonctionnalités.

Ne force donc pas artificiellement une interface gigantesque.

Si nécessaire, utiliser des capabilities :

```typescript
interface AdPlatformCapabilities {
  campaigns: boolean;
  adSets: boolean;
  adGroups: boolean;
  creatives: boolean;
  videoAds: boolean;
  imageAds: boolean;
  conversionTracking: boolean;
  reporting: boolean;
  budgetManagement: boolean;
  audienceManagement: boolean;
}
```

---

# 4. STRUCTURE RECOMMANDÉE

Adapter cette structure aux conventions existantes :

```text
src/
  modules/
    advertising/
      domain/
        entities/
        value-objects/
        interfaces/

      application/
        use-cases/
          connect-platform/
          complete-oauth/
          list-ad-accounts/
          select-ad-account/
          create-campaign/
          update-campaign/
          pause-campaign/
          resume-campaign/
          get-insights/
          disconnect-platform/

      infrastructure/
        providers/
          meta/
          google/
          tiktok/

        encryption/
        persistence/

      presentation/
        controllers/
        dto/

      jobs/
        sync-campaigns/
        sync-insights/
        refresh-tokens/
```

Ne copie pas exactement cette structure si le repository possède déjà une convention différente.

---

# 5. MODÈLE MULTI-TENANT

Utiliser `organizationId` / `workspaceId` / `tenantId` selon le modèle existant.

Ne pas supposer qu'un User = un tenant.

Prévoir :

```text
Organization
    │
    ├── Members
    │
    └── AdPlatformConnections
              │
              ├── Meta
              ├── Google
              └── TikTok
```

---

# 6. ENTITÉ AdPlatformConnection

Créer une entité permettant de représenter une connexion.

Conceptuellement :

```typescript
AdPlatformConnection {
  id
  organizationId

  provider
  status

  externalUserId?
  externalBusinessId?
  externalAccountId?

  accessTokenEncrypted?
  refreshTokenEncrypted?

  accessTokenExpiresAt?

  scopes

  metadata

  lastSyncAt?
  lastError?
  lastErrorAt?

  createdAt
  updatedAt
}
```

Enum :

```text
META
GOOGLE_ADS
TIKTOK_ADS
```

Status :

```text
PENDING
ACTIVE
EXPIRED
REVOKED
ERROR
DISCONNECTED
```

IMPORTANT :

Ne pas supposer que les trois providers ont la même hiérarchie.

Par exemple, conserver les IDs spécifiques dans des champs structurés ou metadata JSON lorsque nécessaire :

```json
{
  "businessId": "...",
  "adAccountId": "...",
  "customerId": "...",
  "managerCustomerId": "..."
}
```

Mais les identifiants essentiels utilisés fréquemment doivent rester indexables.

---

# 7. TOKENS

Les tokens OAuth sont des secrets.

Ils doivent être chiffrés au repos.

Créer une abstraction :

```typescript
interface SecretEncryptionService {
  encrypt(value: string): Promise<string>;
  decrypt(value: string): Promise<string>;
}
```

NE JAMAIS :

```text
console.log(accessToken)
logger.info(accessToken)
return accessToken to frontend
store plaintext token
```

Les tokens ne doivent être déchiffrés que juste avant un appel provider.

Prévoir une stratégie de rotation de clé.

---

# 8. VARIABLES D'ENVIRONNEMENT

Créer/compléter `.env.example`.

IMPORTANT :

Ne mettre AUCUNE vraie valeur.

Utiliser exactement ce genre de structure :

```env
# ============================================================
# APPLICATION
# ============================================================

NODE_ENV=development

APP_URL=http://localhost:3000
API_URL=http://localhost:3000

# Frontend URL used for OAuth redirects if needed
FRONTEND_URL=http://localhost:5173


# ============================================================
# SECURITY
# ============================================================

# 32+ bytes random secret
ENCRYPTION_KEY=

# Optional key identifier for future key rotation
ENCRYPTION_KEY_VERSION=v1


# ============================================================
# META ADS
# ============================================================

META_APP_ID=
META_APP_SECRET=

# OAuth redirect URI registered in Meta Developer Console
META_OAUTH_REDIRECT_URI=http://localhost:3000/api/v1/integrations/meta/oauth/callback

# Optional versioning depending on the currently supported Meta Graph API
META_GRAPH_API_VERSION=

META_OAUTH_SCOPES=


# ============================================================
# GOOGLE ADS
# ============================================================

GOOGLE_ADS_CLIENT_ID=
GOOGLE_ADS_CLIENT_SECRET=

# Google Ads Developer Token
GOOGLE_ADS_DEVELOPER_TOKEN=

# OAuth redirect URI registered in Google Cloud Console
GOOGLE_ADS_OAUTH_REDIRECT_URI=http://localhost:3000/api/v1/integrations/google-ads/oauth/callback

# Optional default manager customer ID / MCC.
# Do NOT use this as the client's customer ID.
GOOGLE_ADS_LOGIN_CUSTOMER_ID=

# Google Ads API version
GOOGLE_ADS_API_VERSION=

# OAuth scope:
# https://www.googleapis.com/auth/adwords
GOOGLE_ADS_OAUTH_SCOPE=https://www.googleapis.com/auth/adwords


# ============================================================
# TIKTOK ADS
# ============================================================

TIKTOK_APP_ID=
TIKTOK_APP_SECRET=

TIKTOK_OAUTH_REDIRECT_URI=http://localhost:3000/api/v1/integrations/tiktok/oauth/callback

TIKTOK_API_BASE_URL=
TIKTOK_API_VERSION=

TIKTOK_OAUTH_SCOPES=


# ============================================================
# QUEUE
# ============================================================

REDIS_URL=redis://localhost:6379

AD_PLATFORM_QUEUE_NAME=ad-platform


# ============================================================
# STORAGE
# ============================================================

OBJECT_STORAGE_PROVIDER=
OBJECT_STORAGE_BUCKET=

OBJECT_STORAGE_ENDPOINT=
OBJECT_STORAGE_REGION=

OBJECT_STORAGE_ACCESS_KEY=
OBJECT_STORAGE_SECRET_KEY=


# ============================================================
# OBSERVABILITY
# ============================================================

LOG_LEVEL=info

# Optional
SENTRY_DSN=
```

IMPORTANT :

Ne crée pas de variables d'environnement pour les tokens des utilisateurs.

Par exemple, ceci est INTERDIT :

```env
CLIENT_META_ACCESS_TOKEN=
CLIENT_GOOGLE_REFRESH_TOKEN=
```

Les tokens clients sont stockés dans la base de données, chiffrés.

---

# 9. GOOGLE ADS

Implémenter le flux OAuth 2.0.

Google Ads nécessite OAuth 2.0 ainsi qu'un Developer Token pour les appels API.

Utiliser le scope :

```text
https://www.googleapis.com/auth/adwords
```

Le système doit supporter l'accès offline afin que les workers puissent effectuer des opérations même lorsque l'utilisateur n'est pas connecté.

Google documente actuellement l'utilisation de :

```text
client ID
client secret
refresh token
developer token
```

pour effectuer les appels API.

Sources officielles à consulter :

* Google Ads OAuth documentation
* Google Ads Developer Token documentation
* Google Ads API authentication documentation

Ne jamais demander le mot de passe Google.

---

# 10. GOOGLE ADS — FLOW

Créer :

```text
GET /integrations/google-ads/connect
```

Le backend génère l'URL OAuth.

Puis :

```text
Google
   ↓
User authorization
   ↓
GET /integrations/google-ads/oauth/callback?code=...
   ↓
exchange code
   ↓
encrypted refresh token
   ↓
list accessible customers
   ↓
return account selection
```

Prévoir une route permettant au frontend de récupérer les comptes accessibles :

```text
GET /integrations/google-ads/accounts
```

Puis :

```text
POST /integrations/google-ads/accounts/select
```

avec :

```json
{
  "customerId": "1234567890"
}
```

Ne jamais considérer `customerId` envoyé par le frontend comme fiable.

Vérifier côté backend que le compte est effectivement accessible via la connexion OAuth.

---

# 11. GOOGLE ADS — MANAGER ACCOUNT

Google Ads utilise une hiérarchie pouvant inclure :

```text
Manager Account / MCC
       │
       ├── Customer A
       ├── Customer B
       └── Customer C
```

Ne pas confondre :

```text
GOOGLE_ADS_LOGIN_CUSTOMER_ID
```

et :

```text
client customerId
```

Le premier correspond au contexte manager lorsque nécessaire.

Le second est le compte publicitaire du client.

Les deux doivent être gérés séparément.

---

# 12. META ADS

Créer un provider :

```text
MetaAdsProvider
```

Le provider doit encapsuler :

* OAuth ;
* access token ;
* Business Manager ;
* Ad Accounts ;
* Pages lorsque nécessaire ;
* Instagram assets lorsque nécessaire ;
* campaigns ;
* ad sets ;
* creatives ;
* ads ;
* insights ;
* conversion tracking.

Le frontend doit simplement afficher :

```text
Connecter Meta Ads
```

Le backend gère le flux OAuth.

---

# 13. META OAUTH

Créer :

```text
GET /integrations/meta/connect
```

Puis :

```text
GET /integrations/meta/oauth/callback
```

Le callback doit :

1. vérifier `state` ;
2. récupérer `code` ;
3. échanger le code côté backend ;
4. récupérer les informations nécessaires ;
5. stocker les tokens chiffrés ;
6. récupérer les comptes accessibles ;
7. créer/mettre à jour la connexion ;
8. rediriger vers le frontend sans exposer de token.

IMPORTANT :

Utiliser un `state` OAuth cryptographiquement sûr.

Le `state` doit être lié à :

```text
organizationId
userId
provider
nonce
expiration
```

et idéalement être stocké côté serveur ou signé de manière sûre.

---

# 14. TIKTOK ADS

Créer :

```text
TikTokAdsProvider
```

Implémenter :

```text
OAuth
Advertiser accounts
Campaigns
Ad groups
Ads
Creatives
Reporting
```

TikTok utilise un `auth_code` après autorisation de l'annonceur, qui est ensuite échangé contre un `access_token`.

Le provider doit encapsuler toute cette logique.

Ne jamais exposer :

```text
app_secret
access_token
```

au frontend.

---

# 15. OAUTH STATE

Créer un mécanisme générique :

```typescript
interface OAuthStateService {
  create(input: OAuthStateInput): Promise<string>;

  verify(state: string): Promise<OAuthStatePayload>;
}
```

Payload conceptuel :

```json
{
  "organizationId": "...",
  "userId": "...",
  "provider": "META",
  "nonce": "...",
  "expiresAt": "..."
}
```

Le callback doit refuser :

* state absent ;
* state expiré ;
* state déjà utilisé ;
* provider différent ;
* tenant différent ;
* user différent.

Protection obligatoire contre CSRF.

---

# 16. ACCOUNT DISCOVERY

Après connexion, le provider doit découvrir les comptes accessibles.

Normaliser :

```typescript
interface AdAccount {
  id: string;
  externalId: string;
  name: string;
  currency?: string;
  timezone?: string;
  status?: string;
  provider: AdPlatform;
  metadata?: Record<string, unknown>;
}
```

Le frontend peut afficher :

```text
Select advertising account

○ My Business - Main
○ My Business - Products
○ My Agency Account
```

Une sélection doit être persistée dans `AdPlatformConnection`.

---

# 17. CONNECTION UX

Créer une page :

```text
Settings
  ↓
Integrations
  ↓
Advertising
```

UI :

```text
Google Ads
──────────────────────
Not connected

[ Connect Google Ads ]


Meta Ads
──────────────────────
Connected
Business: My Business
Account: Digital Products

[ Manage ] [ Disconnect ]


TikTok Ads
──────────────────────
Not connected

[ Connect TikTok Ads ]
```

---

# 18. DISCONNECT

Créer :

```text
DELETE /integrations/:provider
```

Avant déconnexion :

* vérifier ownership ;
* invalider l'état de connexion ;
* supprimer ou révoquer les tokens selon ce que permet le provider ;
* conserver l'historique des campagnes ;
* conserver les statistiques historiques ;
* ne plus permettre de nouveaux appels avec cette connexion.

Ne pas supprimer les données métier simplement parce qu'une intégration est déconnectée.

---

# 19. CAMPAIGN DOMAIN MODEL

Créer des abstractions internes.

Exemple :

```typescript
Campaign {
  id
  organizationId

  productId?

  platformConnectionId

  externalCampaignId

  name

  objective

  status

  budget
  currency

  startAt?
  endAt?

  createdAt
  updatedAt
}
```

Le système doit conserver le mapping :

```text
Internal Campaign
        ↓
External Campaign
```

Exemple :

```text
campaign.id = internal UUID

externalCampaignId = Meta campaign ID
```

Ne jamais utiliser l'ID externe comme ID primaire interne.

---

# 20. CREATIVE MODEL

Prévoir :

```text
Creative
```

avec :

```text
id
organizationId
campaignId?

type
assetId
externalCreativeId
headline
primaryText
description
cta
metadata
status
createdAt
updatedAt
```

Types :

```text
VIDEO
IMAGE
CAROUSEL
TEXT
```

---

# 21. ASSETS

Les vidéos générées précédemment par notre Creative Engine doivent pouvoir être utilisées par AdPlatformProvider.

Pipeline :

```text
Product
   ↓
Creative Engine
   ↓
Video
   ↓
Object Storage
   ↓
Advertising Engine
   ↓
Upload to provider
   ↓
Create Creative
   ↓
Create Ad
```

Ne jamais faire dépendre le provider publicitaire de l'implémentation interne de HeyGen/Veo/Kling.

Le provider reçoit uniquement un asset abstrait :

```typescript
CreativeAsset {
  type
  url
  mimeType
  width?
  height?
  duration?
}
```

---

# 22. CAMPAIGN CREATION

Créer un use case :

```text
CreateAdvertisingCampaign
```

Input :

```json
{
  "platformConnectionId": "...",
  "adAccountId": "...",
  "name": "...",
  "objective": "...",
  "budget": 10000,
  "currency": "XOF",
  "creativeIds": []
}
```

Le use case doit :

1. vérifier le tenant ;
2. vérifier la connexion ;
3. vérifier les permissions ;
4. vérifier le budget ;
5. vérifier les assets ;
6. appeler le provider ;
7. sauvegarder le mapping externe ;
8. enregistrer l'opération ;
9. retourner le résultat.

---

# 23. ASYNCHRONE

Les opérations longues ou sensibles aux rate limits doivent être exécutées via queue.

Par exemple :

```text
POST /campaigns
       ↓
Create Campaign Job
       ↓
202 Accepted
       ↓
BullMQ
       ↓
Advertising Worker
       ↓
Provider API
       ↓
Persist result
```

Créer des jobs :

```text
ad-platform.oauth.refresh
ad-platform.accounts.sync
ad-platform.campaign.create
ad-platform.campaign.update
ad-platform.campaign.pause
ad-platform.campaign.resume
ad-platform.creative.upload
ad-platform.ad.create
ad-platform.insights.sync
ad-platform.conversions.sync
```

---

# 24. IDEMPOTENCE

Toutes les opérations mutables doivent être idempotentes lorsque possible.

Exemple :

```text
CreateCampaign
idempotencyKey = organizationId + campaignRequestId
```

Si un worker redémarre après avoir créé la campagne chez Meta mais avant d'enregistrer le résultat localement, il ne doit pas créer une deuxième campagne.

Prévoir une stratégie de reconciliation.

---

# 25. PROVIDER JOB TRACKING

Créer une entité :

```text
ProviderOperation
```

avec :

```text
id
organizationId
connectionId
provider
operationType
internalResourceId?
externalResourceId?
status
attempts
requestId?
errorCode?
errorMessage?
startedAt?
completedAt?
createdAt
```

NE PAS stocker les access tokens dans cette table.

---

# 26. RETRY

Mettre en place :

```text
exponential backoff
```

pour :

* timeout ;
* rate limit ;
* erreurs temporaires ;
* 5xx.

Ne pas retry automatiquement :

* invalid credentials ;
* permission denied ;
* malformed request ;
* policy violation ;
* invalid account.

---

# 27. TOKEN EXPIRATION

Les providers ne gèrent pas tous les tokens de la même façon.

Le provider doit donc pouvoir exposer :

```typescript
getValidAccessToken(connection)
```

Conceptuellement :

```text
Token valid?
    │
    ├── YES → use token
    │
    └── NO
         ↓
     refresh if supported
         ↓
     update encrypted token
         ↓
     use new token
```

Si refresh impossible :

```text
connection.status = EXPIRED
```

Puis notifier l'utilisateur :

```text
Your Meta connection needs to be reauthorized.
[Reconnect]
```

---

# 28. REPORTING

Créer une abstraction commune :

```typescript
CampaignInsights {
  campaignId
  date

  impressions
  reach
  clicks
  ctr

  spend
  cpc
  cpm

  conversions
  conversionValue

  cpa
  roas

  currency

  metadata
}
```

Attention :

Les plateformes n'ont pas nécessairement les mêmes définitions.

Toujours conserver :

```text
normalized metrics
+
provider raw metadata
```

---

# 29. SYNCHRONISATION

Créer un worker périodique :

```text
sync-ad-platform-insights
```

Il doit :

1. récupérer les connexions actives ;
2. grouper par provider ;
3. récupérer les campagnes ;
4. récupérer les insights ;
5. normaliser ;
6. upsert dans notre DB ;
7. enregistrer les erreurs ;
8. respecter les rate limits.

Ne jamais faire un appel API par utilisateur dans une seule requête HTTP.

---

# 30. CONVERSIONS

Préparer l'architecture pour :

```text
Meta Conversions API
TikTok Events API
Google Ads conversion tracking
Google Analytics / GA4
```

Créer une abstraction :

```typescript
interface ConversionTrackingProvider {
  sendEvent(input: ConversionEvent): Promise<void>;
}
```

Événements :

```text
VIEW_PRODUCT
ADD_TO_CART
INITIATE_CHECKOUT
PURCHASE
LEAD
SIGNUP
```

Exemple :

```json
{
  "event": "PURCHASE",
  "value": 5000,
  "currency": "XOF",
  "orderId": "...",
  "timestamp": "..."
}
```

---

# 31. IMPORTANT POUR LE MARCHÉ AFRICAIN

Le système doit supporter correctement :

```text
XOF
NGN
GHS
KES
ZAR
USD
EUR
```

Ne jamais stocker un montant monétaire sous forme de float.

Utiliser :

```text
integer minor units
```

ou Decimal selon l'architecture existante.

Pour XOF :

```text
5000 XOF
```

doit rester exactement représentable.

---

# 32. BUDGET SAFETY

Comme notre objectif final est de permettre à une IA d'optimiser les campagnes, créer dès maintenant une couche de garde-fous.

Exemple :

```text
AdvertisingSafetyPolicy

maxDailySpend
maxCampaignSpend
maxBudgetIncreasePercent
maxBudgetDecreasePercent
minimumRoas
maximumCpa
requiresApproval
```

Exemple :

```json
{
  "maxDailySpend": 10000,
  "maxCampaignSpend": 100000,
  "maxBudgetIncreasePercent": 20,
  "requiresApproval": true
}
```

NE JAMAIS laisser un agent LLM décider directement d'un appel API de dépense.

Architecture :

```text
AI Decision
     ↓
Safety Policy
     ↓
Validation
     ↓
Human Approval OR Auto Approval
     ↓
Provider API
```

---

# 33. AUDIT LOG

Toute action ayant un impact publicitaire doit être auditée.

Créer :

```text
AdvertisingAuditLog
```

avec :

```text
organizationId
userId?
actorType

provider
connectionId

action
resourceType
resourceId

before
after

createdAt
```

Actor types :

```text
USER
SYSTEM
AI_AGENT
```

Exemple :

```text
AI_AGENT
META
CAMPAIGN
campaign_123
UPDATE_BUDGET

before:
5000 XOF/day

after:
6000 XOF/day
```

---

# 34. PERMISSIONS

Prévoir des permissions internes :

```text
advertising.read
advertising.connect
advertising.create
advertising.update
advertising.pause
advertising.publish
advertising.manage_budget
advertising.manage_integrations
```

Un utilisateur pouvant voir les campagnes ne doit pas automatiquement pouvoir dépenser de l'argent.

---

# 35. API INTERNE

Créer des endpoints REST cohérents.

Exemple :

```http
GET    /api/v1/integrations
GET    /api/v1/integrations/:provider
GET    /api/v1/integrations/:provider/connect
GET    /api/v1/integrations/:provider/callback
DELETE /api/v1/integrations/:provider

GET    /api/v1/ad-accounts
POST   /api/v1/ad-accounts/:id/select

GET    /api/v1/ad-campaigns
POST   /api/v1/ad-campaigns
GET    /api/v1/ad-campaigns/:id
PATCH  /api/v1/ad-campaigns/:id

POST   /api/v1/ad-campaigns/:id/pause
POST   /api/v1/ad-campaigns/:id/resume

GET    /api/v1/ad-campaigns/:id/insights
```

Adapter les URLs aux conventions du projet.

---

# 36. FRONTEND

Créer un écran :

```text
Settings
→ Integrations
→ Advertising
```

Puis :

```text
Meta Ads
Google Ads
TikTok Ads
```

Pour chaque provider :

```text
Not connected
    ↓
Connect
    ↓
OAuth
    ↓
Account selection
    ↓
Connected
```

Afficher :

```text
provider
business
account
currency
timezone
status
last synchronization
```

Ne jamais afficher :

```text
accessToken
refreshToken
appSecret
```

---

# 37. ERREURS UX

Prévoir des erreurs explicites :

```text
CONNECTION_EXPIRED
PERMISSION_DENIED
ACCOUNT_NOT_FOUND
ACCOUNT_NOT_AUTHORIZED
RATE_LIMITED
PROVIDER_UNAVAILABLE
INVALID_CREATIVE
POLICY_REJECTED
BUDGET_LIMIT_EXCEEDED
```

Le frontend doit transformer cela en messages compréhensibles.

---

# 38. TESTS

Écrire au minimum :

### Unit tests

* OAuth state generation ;
* OAuth state validation ;
* encryption/decryption ;
* tenant isolation ;
* token expiration ;
* provider mapping ;
* campaign normalization ;
* insight normalization ;
* budget safety ;
* permission checks.

### Integration tests

Tester chaque provider derrière des mocks.

Ne pas appeler les APIs réelles pendant les tests unitaires.

Créer :

```text
MockMetaAdsProvider
MockGoogleAdsProvider
MockTikTokAdsProvider
```

---

# 39. CONTRACT TESTS

Chaque provider doit respecter le même contrat.

Créer une suite de tests générique :

```text
AdPlatformProviderContractTests
```

Puis exécuter le contrat sur :

```text
MetaAdsProvider
GoogleAdsProvider
TikTokAdsProvider
```

Cela garantit que les implémentations restent cohérentes.

---

# 40. SECURITY TESTS

Tester spécifiquement :

### Tenant isolation

Utilisateur A ne doit jamais pouvoir :

```text
read connection B
use connection B
read campaign B
use campaign B
```

Même s'il connaît les IDs.

### OAuth CSRF

Un callback avec un mauvais state doit être refusé.

### Token exposure

Vérifier que les tokens n'apparaissent pas dans :

```text
logs
errors
API responses
database plaintext
```

### Authorization

Un utilisateur sans permission ne doit pas pouvoir publier ou modifier une campagne.

---

# 41. RATE LIMITING

Chaque provider doit avoir son propre mécanisme de rate limiting.

Architecture :

```text
AdvertisingService
       │
       ▼
Provider
       │
       ▼
RateLimiter(provider)
       │
       ▼
External API
```

Ne pas utiliser un rate limiter global unique.

---

# 42. OBSERVABILITÉ

Chaque appel provider doit avoir :

```text
traceId
organizationId
connectionId
provider
operation
externalRequestId?
duration
status
```

Mais JAMAIS :

```text
accessToken
refreshToken
clientSecret
appSecret
```

---

# 43. COST TRACKING

Préparer :

```text
AdvertisingOperationCost
```

pour connaître :

```text
API operation
provider
organization
campaign
date
estimated cost
```

Même si les APIs publicitaires ne facturent pas directement chaque appel, cela permettra plus tard de calculer :

```text
AI cost
video generation cost
advertising spend
SaaS margin
```

---

# 44. ARCHITECTURE FINALE

Le système doit évoluer vers :

```text
                         USER
                           │
                           ▼
                     DIGITAL PRODUCT
                           │
                           ▼
                    MARKETING ENGINE
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
       Creative Engine             Campaign Engine
             │                           │
       ┌─────┼─────┐                     │
       ▼     ▼     ▼                     │
    HeyGen  Veo   Image                  │
       │     │     │                     │
       └─────┼─────┘                     │
             ▼                           ▼
         CREATIVE                 AD PLATFORM LAYER
                                       │
                    ┌──────────────────┼──────────────────┐
                    ▼                  ▼                  ▼
                 Meta Ads          Google Ads         TikTok Ads
                    │                  │                  │
                    └──────────────────┼──────────────────┘
                                       ▼
                                  PERFORMANCE
                                       │
                                       ▼
                               CONVERSION TRACKING
                                       │
                                       ▼
                                  AI ANALYSIS
                                       │
                                       ▼
                               SAFETY / POLICY
                                       │
                                       ▼
                              HUMAN / AUTOPILOT
                                       │
                                       ▼
                                  OPTIMIZATION
```

---

# 45. ORDRE D'IMPLÉMENTATION

Ne tente pas de tout implémenter simultanément.

Ordre recommandé :

## Phase 1

Architecture générique :

```text
AdPlatformProvider
AdPlatformConnection
OAuthStateService
SecretEncryptionService
```

## Phase 2

Meta OAuth + account discovery.

## Phase 3

Google Ads OAuth + account discovery.

## Phase 4

TikTok OAuth + account discovery.

## Phase 5

Campaign CRUD.

## Phase 6

Creative upload.

## Phase 7

Insights.

## Phase 8

Conversion tracking.

## Phase 9

Safety policies.

## Phase 10

AI optimization.

---

# 46. MVP PRIORITAIRE

Pour la première version, le système doit simplement permettre :

```text
1. User opens Integrations

2. User clicks:
   Connect Meta

3. OAuth

4. User authorizes

5. Backend stores encrypted credentials

6. Backend retrieves advertising accounts

7. User selects an account

8. Connection becomes ACTIVE

9. User can see campaigns

10. User can create a campaign

11. User can pause/resume campaign

12. User can retrieve insights

13. User can disconnect
```

Faire ensuite exactement le même flux pour :

```text
Google Ads
TikTok Ads
```

---

# 47. NE PAS FAIRE

INTERDIT :

```text
if provider === "META" everywhere
```

ou :

```text
if provider === "GOOGLE"
```

dans le domaine métier.

INTERDIT :

```text
Meta API calls inside controllers
```

INTERDIT :

```text
access tokens in frontend
```

INTERDIT :

```text
provider secrets in database
```

INTERDIT :

```text
one global ad account
```

INTERDIT :

```text
hard-coded campaign IDs
```

INTERDIT :

```text
HTTP request waiting for provider job
```

INTERDIT :

```text
LLM directly calling provider APIs
```

---

# 48. IA ET OUTILS

À terme, l'IA pourra appeler des tools internes :

```text
list_ad_accounts
list_campaigns
get_campaign_insights
create_campaign
pause_campaign
resume_campaign
update_campaign_budget
create_creative
```

Mais le LLM ne doit jamais avoir accès directement aux credentials.

Architecture :

```text
LLM
 │
 ▼
Tool
 │
 ▼
Application Service
 │
 ▼
Permission Check
 │
 ▼
Safety Policy
 │
 ▼
AdPlatformProvider
 │
 ▼
External API
```

---

# 49. DOCUMENTATION

Créer une documentation :

```text
docs/integrations/advertising.md
```

Documenter :

* architecture ;
* OAuth ;
* providers ;
* environment variables ;
* setup Meta ;
* setup Google ;
* setup TikTok ;
* local development ;
* production configuration ;
* security ;
* token lifecycle ;
* account selection ;
* troubleshooting.

Créer également :

```text
.env.example
```

complet.

---

# 50. CONFIGURATION FOURNIE PAR LE PROJET

Avant de choisir une librairie ou une implémentation :

* inspecte `package.json` ;
* inspecte les dépendances existantes ;
* réutilise les SDK officiels lorsque disponibles et adaptés ;
* évite d'ajouter une librairie uniquement pour une petite abstraction ;
* respecte les versions déjà utilisées ;
* vérifie les APIs officielles actuelles avant d'implémenter les endpoints ;
* n'invente jamais un endpoint provider.

Pour Google Ads, privilégier la librairie officielle/supportée lorsque pertinente.

Pour Meta et TikTok, utiliser leur documentation/API officielle actuelle.

---

# 51. LIVRABLES

À la fin de l'implémentation, fournir :

### Code

* providers ;
* OAuth ;
* database models ;
* migrations ;
* services ;
* workers ;
* controllers ;
* DTO ;
* tests.

### Configuration

```text
.env.example
```

### Documentation

```text
docs/integrations/advertising.md
```

### Résumé

Indiquer :

```text
Files created
Files modified
Database migrations
Environment variables
OAuth configuration required
API permissions required
Known limitations
Next recommended steps
```

---

# 52. CRITÈRES D'ACCEPTATION

L'implémentation est considérée comme correcte uniquement si :

* [ ] plusieurs organisations peuvent connecter le même provider ;
* [ ] chaque organisation possède ses propres connexions ;
* [ ] OAuth fonctionne ;
* [ ] OAuth state est sécurisé ;
* [ ] tokens chiffrés ;
* [ ] aucun token dans le frontend ;
* [ ] aucun secret dans les logs ;
* [ ] tenant isolation testée ;
* [ ] account discovery fonctionne ;
* [ ] sélection du compte fonctionne ;
* [ ] Meta provider isolé ;
* [ ] Google provider isolé ;
* [ ] TikTok provider isolé ;
* [ ] campaigns utilisent des IDs internes ;
* [ ] external IDs correctement persistés ;
* [ ] jobs asynchrones ;
* [ ] retries ;
* [ ] rate limiting ;
* [ ] token expiration ;
* [ ] disconnect ;
* [ ] audit logs ;
* [ ] permissions ;
* [ ] tests unitaires ;
* [ ] tests d'intégration ;
* [ ] `.env.example` complet ;
* [ ] documentation complète.

---

# 53. IMPORTANT — COMPORTEMENT DE CODEX

Si tu rencontres une ambiguïté :

1. inspecte d'abord le code existant ;
2. consulte la documentation officielle du provider ;
3. respecte l'architecture existante ;
4. ne fais pas de grosse réécriture ;
5. ne devine pas les APIs ;
6. ne hard-code aucune valeur sensible ;
7. ne supprime aucune fonctionnalité existante sans raison ;
8. écris les tests avant ou en même temps que les nouvelles fonctionnalités lorsque possible.

Commence maintenant par **l'audit du repository**.

Ne code rien avant d'avoir présenté :

```text
1. Architecture actuelle
2. Architecture proposée
3. Modèle de données proposé
4. Flux OAuth
5. Environment variables
6. Découpage des tâches
7. Risques / points à vérifier
```

Puis attends ma validation avant de commencer l'implémentation.
