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
| GET | `/api/v1/ad-campaigns/{id}/insights?since=&until=` | Métriques normalisées (YYYY-MM-DD) |

## 4. Variables d'environnement

```env
APP_URL=                        # URL publique du frontend (redirections OAuth)
ENCRYPTION_KEY=                 # 32+ caractères aléatoires (AES-256-GCM)
ENCRYPTION_KEY_VERSION=v1
META_APP_ID=
META_APP_SECRET=
META_GRAPH_API_VERSION=v23.0
META_OAUTH_REDIRECT_URI=        # ex. https://api.example.com/api/v1/integrations/meta/callback
META_OAUTH_SCOPES=              # défaut: ads_management,ads_read,business_management,pages_show_list,pages_read_engagement
```

Aucune variable pour les tokens clients (interdit — ils sont chiffrés en DB).

## 5. Setup Meta (developers.facebook.com)

1. Créer une app (type Business) ; noter `App ID` / `App secret`.
2. Produit **Facebook Login** : ajouter `META_OAUTH_REDIRECT_URI` aux *Valid OAuth Redirect URIs*.
3. Permissions à demander : `ads_management`, `ads_read`, `business_management`, `pages_show_list`, `pages_read_engagement` (re-validation Meta requise hors mode dev ; en mode dev, seuls les rôles de l'app peuvent se connecter).
4. L'utilisateur doit avoir une **page Facebook** + un compte publicitaire actif.

## 6. Setup Google Ads / TikTok Ads (phases suivantes)

Mêmes endpoints (`provider=google_ads` / `tiktok_ads`) dès que les providers
sont enregistrés dans `ProviderRegistry`. Google Ads requiert un **developer
token** (accès comptes test au début) + OAuth offline (`refresh_token`) ;
TikTok utilise un `auth_code` échangé contre `access_token`.

## 7. Sécurité & opérations

- Tokens chiffrés AES-256-GCM (nonce aléatoire, format `v1:<nonce>:<ct>`) ;
  déchiffrés uniquement en mémoire, juste avant un appel provider.
- Refresh : si expiration < 24 h, prolongation automatique ; échec → statut
  `expired` (l'utilisateur reconnecte).
- Chaque opération mutative est tracée dans `provider_operations`
  (`attempts`, statut, `external_resource_id`) pour réconciliation/idempotence.
- Garde-fous budget côté service (`SafetyPolicy`) : refus avant tout appel
  provider si le budget dépasse la limite.
- Erreurs Graph typées (`graphError.Transient()` : codes 1/2/4/17/32/613) —
  base des retries à backoff (worker asynq à venir).

## 8. Développement local

- Sans `META_APP_ID/SECRET`, aucun provider n'est enregistré : `/integrations`
  répond 422 sur connect (l'UI liste les plateformes déconnectées).
- L'upload de creatives requiert le stockage S3 (URLs présignées) —
  indisponible avec `LocalStorage` en dev.
- Tester le callback en local : utiliser un tunnel HTTPS (les redirect URIs
  doivent être enregistrés chez Meta).
