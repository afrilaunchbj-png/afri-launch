# Intégrations publicitaires (Meta Ads, Google Ads, TikTok Ads)

> ADR-017 — couche multi-tenant provider-agnostic. Les vidéos générées par
> AfriLaunch (ADR-016) deviennent des creatives publiables en campagnes.

## 1. Architecture

```
application/port/advertising.go   # AdPlatformProvider + Capabilities, SecretEncryptor,
                                  # OAuthStateStore, repos connexions/campagnes/…
application/advertising/service.go # orchestration générique (aucune logique provider)
infra/ads/meta/                    # MetaAdsProvider (Graph API)
infra/crypto/                      # AES-256-GCM (ENCRYPTION_KEY, préfixe version v1:)
infra/postgres/ad_repo.go          # repos postgres (tokens chiffrés au repos)
server/handler/integrations.go     # endpoints + callback OAuth
```

**Règles** (prompts/marketing-flow.md) :

- aucune logique Meta/Google/TikTok hors de `infra/ads/<provider>` ;
- aucun token : pas de log, pas de réponse API, pas de frontend, jamais en clair en DB ;
- montants en **entiers unités mineures** (jamais de float) ;
- l'ID de compte envoyé par le frontend est toujours **re-vérifié** chez le provider ;
- les campagnes sont créées **en pause** ; la publication est un geste utilisateur.

## 2. Flux OAuth (Meta)

```
FE: GET /api/v1/integrations/meta/connect        → { authorization_url } (state CSRF lié user)
User autorise chez Meta
Meta → GET /api/v1/integrations/meta/callback?code&state   (sans JWT)
       ├─ state consommé (anti-replay, anti-CSRF, 10 min TTL) → identifie l'utilisateur
       ├─ code → token courte durée → prolongation longue durée (~60 j)
       ├─ tokens chiffrés (AES-GCM) + statut pending
       └─ 302 → {APP_URL}/integrations?connect=meta&status=success|error
FE: choix du compte → POST /integrations/meta/accounts/select
       (vérifié via GET /act_{id} → statut active + devise en metadata)
```

`page_id` (page Facebook requise par `object_story_spec`) est stocké dans les
métadonnées de la connexion ; sans page, l'upload de creative est refusé.

## 3. Endpoints

| Méthode | Route | Description |
|---|---|---|
| GET | `/api/v1/integrations` | Connexions + capacités par plateforme |
| GET | `/api/v1/integrations/{provider}/connect` | URL d'autorisation OAuth |
| GET | `/api/v1/integrations/{provider}/callback` | Callback OAuth (public, redirige vers le FE) |
| GET | `/api/v1/integrations/{provider}/accounts` | Comptes accessibles (discovery) |
| POST | `/api/v1/integrations/{provider}/accounts/select` | Sélection du compte (`{external_account_id}`) |
| DELETE | `/api/v1/integrations/{provider}` | Déconnexion (révoque, historique conservé) |
| POST | `/api/v1/integrations/{provider}/campaigns/sync` | Synchronise les campagnes |
| GET/POST | `/api/v1/ad-campaigns` | Liste / crée une campagne (créée en pause) |
| POST | `/api/v1/ad-campaigns/{id}/pause` · `/resume` | Pause / reprise |
| POST | `/api/v1/ad-campaigns/{id}/creatives` | Publie un asset interne comme créative (`{asset_id, headline, primary_text, cta}`) |
| GET | `/api/v1/ad-campaigns/{id}/insights?since=&until=` | Métriques normalisées (YYYY-MM-DD) |
| GET | `/api/v1/ad-creatives` | Créatives publiées (mapping interne/externe) |

## 4. Variables d'environnement

```env
APP_URL=                        # URL publique du frontend (redirections OAuth)
ENCRYPTION_KEY=                 # 32+ caractères aléatoires (AES-256-GCM)
ENCRYPTION_KEY_VERSION=v1

# Meta
META_APP_ID=
META_APP_SECRET=
META_GRAPH_API_VERSION=v23.0
META_OAUTH_REDIRECT_URI=
META_OAUTH_SCOPES=

# Google Ads
GOOGLE_ADS_CLIENT_ID=
GOOGLE_ADS_CLIENT_SECRET=
GOOGLE_ADS_DEVELOPER_TOKEN=
GOOGLE_ADS_OAUTH_REDIRECT_URI=
GOOGLE_ADS_LOGIN_CUSTOMER_ID=
GOOGLE_ADS_API_VERSION=         # défaut v19

# TikTok Ads
TIKTOK_APP_ID=
TIKTOK_APP_SECRET=
TIKTOK_OAUTH_REDIRECT_URI=
```

Aucune variable pour les tokens clients (interdit — ils sont chiffrés en DB).

## 5. Providers Google Ads et TikTok Ads (implémentés)

Mêmes endpoints (`provider=google_ads` / `tiktok_ads`) — implémentés via
**l'API REST officielle** (cohérence avec les autres clients HTTP du repo,
pas de gRPC au MVP).

### Google Ads (`infra/ads/googleads`)

- **OAuth offline** : `access_type=offline` + `prompt=consent` →
  `refresh_token` persisté chiffré ; refresh automatique (access token 1 h).
- **Discovery** : `customers:listAccessibleCustomers` puis info de chaque
  compte via `searchStream` (`SELECT customer... FROM customer`).
- **Campagnes** : lecture GAQL (`campaign` + `campaign_budget`) ; création en
  deux appels — `campaignBudgets:mutate` (micros = minor × 10 000) puis
  `campaigns:mutate` (canal SEARCH, `manualCpc`, statut PAUSED) ; pause/resume
  via `updateMask=status`.
- **Insights** : `metrics.impressions/clicks/costMicros/conversions` par jour
  (`segments.date`) ; conversion micros → unités mineures (÷ 10 000, 2 décimales).
- **Créatives non supportées au MVP** (`Capabilities.Creatives=false`) : la
  publication d'annonces exige ad groups + ads (phase suivante).
- Requiert `GOOGLE_ADS_DEVELOPER_TOKEN` (accès comptes test avant validation
  Google) ; `GOOGLE_ADS_LOGIN_CUSTOMER_ID` si MCC.

### TikTok Ads (`infra/ads/tiktok`, Business API v1.3)

- **OAuth** : `auth_code` échangé (`/oauth2/access_token/`) → access_token +
  refresh_token ; refresh via `/oauth2/refresh_token/`.
- **Discovery** : `/oauth2/advertiser/get/` (comptes liés au token) ;
  `VerifyAdAccount` vérifie l'appartenance à cette liste.
- **Campagnes** : `/campaign/get/`, création `/campaign/create/`
  (statut `CAMPAIGN_STATUS_DISABLE` = en pause), pause/resume via
  `/campaign/status/update/` (DISABLE/ENABLE).
- **Insights** : `/report/integrated/get/` (BASIC, AUCTION_CAMPAIGN) ;
  spend en unités majeures → mineures (× 100).
- **Créatives non supportées au MVP** (upload vidéo + adgroup/ad : phase
  suivante).

### Créatives (Meta)

`POST /ad-campaigns/{id}/creatives` publie un asset interne (ex. vidéo d'un
projet, ADR-016) : URL signée S3 24 h → `advideos` (file_url) → `adcreatives`
(`object_story_spec` avec la `page_id` de la connexion). Google/TikTok
refusent poliment tant que `Capabilities.Creatives` est à false.

## 6. Setup Meta (developers.facebook.com)

1. Créer une app (type Business) ; noter `App ID` / `App secret`.
2. Produit **Facebook Login** : ajouter `META_OAUTH_REDIRECT_URI` aux *Valid OAuth Redirect URIs*.
3. Permissions à demander : `ads_management`, `ads_read`, `business_management`, `pages_show_list`, `pages_read_engagement` (re-validation Meta requise hors mode dev ; en mode dev, seuls les rôles de l'app peuvent se connecter).
4. L'utilisateur doit avoir une **page Facebook** + un compte publicitaire actif.

## 7. Setup Google Ads

1. Projet Google Cloud : OAuth consent + ID client (Web) avec `GOOGLE_ADS_OAUTH_REDIRECT_URI`.
2. **Developer token** (Google Ads API Center, niveau compte test pour débuter).
3. L'utilisateur Google doit avoir accès au compte Google Ads ciblé.

## 8. Setup TikTok Ads

1. Créer une app sur le **TikTok for Business** developer portal.
2. Déclarer `TIKTOK_OAUTH_REDIRECT_URI` (redirect callback).
3. Les advertisers doivent être rattachés à l'app (business API v1.3).

## 9. Sécurité & opérations

- Tokens chiffrés AES-256-GCM (nonce aléatoire, format `v1:<nonce>:<ct>`) ;
  déchiffrés uniquement en mémoire, juste avant un appel provider.
- Refresh : si expiration < 24 h, prolongation/renouvellement automatique
  (Meta = prolongation, Google/TikTok = refresh_token) ; échec → statut
  `expired` (l'utilisateur reconnecte).
- Chaque opération mutative est tracée dans `provider_operations`
  (`attempts`, statut, `external_resource_id`) pour réconciliation/idempotence.
- Garde-fous budget côté service (`SafetyPolicy`) : refus avant tout appel
  provider si le budget dépasse la limite.
- Erreurs provider typées (`Transient()` : Meta codes 1/2/4/17/32/613,
  Google 408/429/5xx) — base des retries à backoff (worker asynq à venir).

## 10. Développement local

- Sans credentials provider, aucun provider n'est enregistré : `/integrations`
  répond 422 sur connect (l'UI liste les plateformes déconnectées).
- L'upload de créatives requiert le stockage S3 (URLs présignées) —
  indisponible avec `LocalStorage` en dev ; Google/TikTok refusent au MVP.
- Tester les callbacks en local : tunnel HTTPS (redirect URIs enregistrés
  chez chaque plateforme).
