# Audit de sécurité — AfriLaunch

> Audit statique (lecture seule) du backend Go et du frontend React, basé sur le code au commit `8910a80` (5 septembre 2026). Aucun correctif n'a été appliqué dans le cadre de cet audit.

## 1. Périmètre & méthode

- **Backend** : `backend/` — routeur chi (`router.go`), middlewares (`middleware.go`, `ratelimit.go`, `auth.go`, `cors`), handlers, services, repos sqlc/pgx, vérifieur Neon Auth, `cmd/api/main.go`.
- **Frontend** : `frontend/` — `router.tsx`, layouts, garde `ProtectedRoute`, client API, gestion de session Neon Auth, `nginx.conf.template`, config build Vite.
- **Référentiel** : OWASP ASVS / Top 10 (2021), focus demandé sur *« routes frontend accessibles sans contrôle backend »* et *« redirection vers login »*.
- Ce document liste les **points forts vérifiés** puis les **écarts** par gravité, avec emplacement `fichier:ligne` et recommandation.

---

## 2. Réponse aux questions ciblées

### 2.1 Y a-t-il des routes frontend accessibles sans contrôle backend ?

**Non pour l'accès aux données.** Toutes les pages applicatives de la SPA (`/dashboard`, `/discover`, `/projects*`, `/integrations`, `/credits`, `/support*`, `/settings`, `/admin*`) sont regroupées sous `<ProtectedRoute>` (`frontend/src/app/router.tsx:51-80`), qui exige une session Neon Auth côté navigateur et redirige vers `/login` sinon (`frontend/src/features/auth/components/protected-route.tsx:14-16`).

En complément — et c'est le contrôle **effectif** — chaque route d'API manipulée par ces pages est protégée par `RequireAuth` (JWT Bearer vérifié EdDSA/JWKS) côté backend (`backend/internal/server/router.go:77-78`, `auth.go:16-32`). Aucune route applicative n'existe côté frontend sans équivalent backend authentifié. La sécurité ne repose **pas** sur la garde UI (purement cosmétique/UX), mais sur le middleware backend.

**Seules routes backend publiques** (volontairement sans JWT) :
- `GET /healthz`, `GET /readyz`, `GET /api/v1/health` — healthchecks.
- `GET /api/v1/markets` — référentiel public (landing).
- `GET /api/v1/integrations/{provider}/callback` — callback OAuth, utilisateur résolu par **état CSRF à usage unique** (voir §5).
- `POST /api/v1/payments/webhook/{provider}` — notification serveur-à-serveur, **jamais considérée comme preuve** (statut reconfirmé par API, voir §5).

### 2.2 La redirection vers login est-elle correcte ?

Oui pour le cas nominal : session absente → `ProtectedRoute` → `<Navigate to="/login" state={{from}}>`. Points à améliorer (non bloquants, voir §4) :
- Le **rôle admin** n'est pas vérifié côté client sur `/admin*` (seul le masquage de la nav l'utilise). Un utilisateur `user` qui force `/admin/users` charge la page puis reçoit des 403 du backend — **la sécurité tient**, mais l'UX est cassée (aucune vue « accès refusé »).
- Un **401/403 en cours de session** (JWT expiré ~15 min, cache 10 min, session révoquée) n'est pas intercepté globalement : pas de purge du cache token, pas de `signOut()`, pas de redirection `/login`. Les requêtes échouent en boucle jusqu'à expiration du cache mémoire.

---

## 3. Points forts vérifiés (positifs)

| # | Domaine | Constat | Emplacement |
|---|---|---|---|
| P1 | Injection SQL | 100 % des requêtes sont paramétrées via sqlc ; aucune concaténation SQL manuelle dans le code métier | `db/query/*.sql` → `internal/infra/postgres/db/*.sql.go` |
| P2 | Contrôle d'accès (IDOR) | Tous les chemins d'écriture par `{id}` passent par un `Get` scopé `user_id` : conversations, projets, support, paiements, campagnes ads, assets, jobs, idées, crédits | voir tableau §6 |
| P3 | Auth JWT | Algorithme restreint à `EdDSA`, clé choisie par `kid` parmi JWKS filtré (OKP/Ed25519), issuer exigé — pas de confusion d'algorithme | `backend/internal/infra/auth/neon.go:71-89,112-152` |
| P4 | Secrets | Aucun secret hardcodé, aucun endpoint debug/pprof. Fail-closed : `DenyVerifier` (401) si Neon Auth absent, paiements désactivés sans `PAYMENT_PROVIDER`, chiffreur désactivé sans clé | `cmd/api/main.go:99-105,200-218` ; `infra/auth/neon.go:178-186` |
| P5 | Stockage tokens FE | JWT conservé **en mémoire** (jamais `localStorage`/`sessionStorage`) ; session Better Auth en cookie httpOnly (domaine Neon) | `frontend/src/lib/auth.ts:14-30` |
| P6 | XSS | Aucun `dangerouslySetInnerHTML`/`innerHTML` ; contenu LLM du chat rendu en nœuds texte React (échappé) ; pas de renderer markdown | `frontend/src/features/chat/chat-messages.tsx:45-61` |
| P7 | Erreurs / fuite d'info | Erreurs RFC 9457 ; les 500 ne renvoient aucun détail technique au client (log serveur uniquement) ; logs HTTP sans corps ni headers d'auth | `backend/internal/server/handler/respond.go:48-58` ; `middleware.go:13-28` |
| P8 | Chiffrement au repos | Tokens OAuth des providers chiffrés AES-256-GCM (`v1:nonce:ct`), jamais loggés | `backend/internal/infra/crypto/crypto.go` |
| P9 | Paiements | Webhook public mais **reconfirmé par API** (`VerifyStatus`) avant octroi ; octroi de crédits **idempotent** via ledger `payment:<id>` ; retour navigateur = resync uniquement | `backend/internal/application/payments/service.go:142-221` |
| P10 | OAuth ads | Callback public protégé par état CSRF créé user+provider, TTL 10 min, consommation **usage unique** | `backend/internal/application/advertising/service.go:94-107` ; `advertising.sql:57-66` |
| P11 | Validation entrées | `decodeJSON` : `http.MaxBytesReader` 1 MiB + `DisallowUnknownFields` ; webhooks limités à 1 Mo ; uploads multipart ≤ 5 Mo / 4 fichiers | `handler/decode.go:12-23`, `handler/payments.go:131`, `handler/support.go:176-197` |
| P12 | CORS | Allowlist explicite (`ALLOWED_ORIGINS`), origine autorisée réfléchie uniquement, pas de `*` | `backend/internal/server/router.go:184-208` |
| P13 | Commande système | `ffmpeg`/`ffprobe` exécutés en **args tableau** (jamais de shell), sous-titres via fichier SRT, HTML échappé | `backend/internal/infra/render/ffmpeg.go:250-277` |
| P14 | Paiement FE | Checkout 100 % hébergé provider ; aucune clé de paiement côté client | `frontend/src/features/credits/plans-panel.tsx`, `credits/api.ts` |

---

## 4. Écarts et recommandations (par gravité)

### 4.1 Élevée

#### E1 — Rate limiting défini mais jamais branché
- **Constat** : `rateLimit` est implémenté (`ratelimit.go:49-62`) mais **aucune route ne l'utilise** (grep : seule définition). Aucune limite sur les endpoints coûteux/non facturés côté montant : `POST /payments/checkout`, générations projets (LLM/HeyGen/chromedp/ffmpeg), `POST /conversations/{id}/messages`, `POST /research`, `POST /credits/reserve`, `GET /events` (SSE), webhooks publics.
- **Impact** : abus de coût opérateur, déni de service local (worker + broker in-process), spam de checkout. Les crédits limitent l'impact *par utilisateur* mais pas les comptes jetables.
- **Recommandation (OWASP A07/A04)** : brancher `rateLimit` (Redis en prod) sur au minimum checkout, réservation de crédits, endpoints d'IA/génération, envoi de messages, et sur les webhooks/callback publics. Limiter aussi le nombre de connexions SSE ouvertes par utilisateur.

#### E2 — IP non fiable pour toute future limitation
- **Constat** : `clientIP()` lit `X-Forwarded-For` puis `X-Real-IP` sans notion de proxy de confiance (`ratelimit.go:64-71`) ; `middleware.RealIP` (chi) fait de même et est déprécié (`router.go:51`). Un client peut donc **spoofer** son IP dès qu'une limitation sera branchée.
- **Recommandation (OWASP A07)** : utiliser le pair direct ou une liste de proxies fiables (`middleware.ForwardedHeaders`/`TrustProxies`), striper les en-têtes de forwarding au reverse-proxy.

### 4.2 Moyenne

#### M1 — Aucun en-tête de sécurité HTTP global
- **Constat** : ni CSP, HSTS, `X-Frame-Options`, `X-Content-Type-Options`, `Referrer-Policy` ni `Permissions-Policy`. Exceptions isolées : `nosniff` sur 2 téléchargements support (`support.go:223,237`), `X-Accel-Buffering: no` sur SSE (`events.go:33`). Le frontend nginx (`nginx.conf.template:1-17`) et `index.html` non plus.
- **Recommandation (OWASP A05)** : middleware d'en-têtes sur l'API (`Content-Security-Policy: default-src 'none'`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`) et `add_header` au niveau `server` du nginx (CSP SPA, nosniff, XFO, Referrer-Policy, HSTS si TLS terminé là).

#### M2 — Timeouts serveur HTTP quasi absents
- **Constat** : seul `ReadHeaderTimeout: 5s` est défini (`cmd/api/main.go:280-284`). Pas de `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`.
- **Recommandation (OWASP A05)** : fixer `ReadTimeout`/`IdleTimeout`/`MaxHeaderBytes` ; ne pas poser de `WriteTimeout` global (le SSE et les téléchargements sont longs) — gérer l'inactivité du SSE au niveau broker/heartbeat.

#### M3 — Listing des assets d'un projet non scopé à l'utilisateur
- **Constat** : `GET /api/v1/projects/{id}/assets` ne lit pas `authctx.UserID` (`handler/assets.go:33-45`) ; `ListByProject` filtre sur `project_id` seul (`db/query/assets.sql:9-10`). Un utilisateur authentifié connaissant l'UUID d'un projet d'autrui peut lister ses **métadonnées** (types, noms, dates). Le **download** reste protégé (`assets.go:49-50`, `WHERE id AND user_id`).
- **Impact** : fuite de métadonnées limitée (UUID non devinable), mais violation du principe d'accès par objet.
- **Recommandation (OWASP A01)** : passer `userID` et vérifier l'ownership du projet (`ListByProject(ctx, userID, projectID)` avec jointure/`EXISTS` sur `projects`), comme le fait déjà `Download`.

#### M4 — JWT : pas de vérification d'audience, profil non recoupé
- **Constat** : `Verify` valide signature/issuer/exp mais **pas `aud`** (`neon.go:71-73`) ; les claims `email`/`name`/`image` remplissent le profil local sans recoupement (`auth/service.go:39-78`). Si l'issuer/JWKS Neon est partagé entre applications, un token d'une autre app serait accepté.
- **Recommandation (OWASP A01/A05)** : ajouter `jwt.WithAudience("<audience attendue>")` si Neon émet un `aud` stable, ou au minimum valider un claim d'usage (`token_use`/type) ; considérer le **leeway** d'horloge.

#### M5 — 401/403 et session expirée non gérés globalement côté FE
- **Constat** : le client API lève `AppError` mais aucun interceptor ne purge le cache token / ne déconnecte / ne redirige vers `/login` (`frontend/src/lib/api/client.ts:58-60`) ; le cache token de 10 min peut servir un JWT déjà expiré en boucle (`lib/auth.ts:22`). L'absence de garde de rôle admin côté client aggrave l'UX sur `/admin*` (`router.tsx:70-79`).
- **Recommandation (OWASP A01/A05)** : interceptor central — sur 401 : `clearAccessTokenCache()` + `signOut()` + `navigate("/login")` ; sur 403 : vue « accès refusé » ; ajouter un garde `RequireRole("superadmin")` sur les routes `/admin*` (via `useMe()`), **en complément** du contrôle serveur.

### 4.3 Basse

#### B1 — Webhooks : aucune validation de signature (compensée)
- **Constat** : le handler collecte les en-têtes de signature (`payments.go:136-143`) mais les adaptateurs fournis ne les vérifient pas ; la sécurité repose sur la reconfirmation API (`service.go:159-163`), ce qui bloque l'auto-octroi. Restent : un oracle non authentifié déclenchant des appels sortants et l'absence de rate limit.
- **Recommandation** : vérifier les signatures quand le provider en expose (PawaPay `Signature`, FedaPay `X-Fedapay-Signature`, PayDunya…), brancher un rate limit IP, et répondre rapidement (200) pour limiter la charge.

#### B2 — `Content-Disposition` construite par concaténation
- **Constat** : nom de fichier inséré tel quel entre guillemets (`assets.go:58`, `support.go:222,236`) ; `strings.TrimSpace` seulement côté support. Risque d'injection d'en-tête (CRLF) très faible, pas de traversée (clé randomisée + `safePath`).
- **Recommandation** : encoder le nom via `mime.FormatMediaType("attachment", ...)` ou striper `\r\n`/`"`.

#### B3 — Uploads : MIME déclaré par le client, pas de magic-bytes
- **Constat** : allowlist MIME (`support.go:198-203`) basée sur le `Content-Type` envoyé (repli extension), pas de vérification du contenu réel ; pas de quota global par utilisateur.
- **Recommandation** : détecter le type réel (magic bytes), limiter le volume/taille cumulé, stocker hors racine web (déjà le cas).

#### B4 — Quelques écritures SQL non scopées, protégées « en amont » uniquement
- **Constat** : `UpdateIdeaContent`, `TouchConversation`/`SetTitle`, `UpdateJobStatus`, `UpdateResearchRequestStatus`, `UpdatePaymentCheckout`/`MarkPaymentStatus`, `UpdateAdCampaign…` filtrent par `id` seul ; sûres aujourd'hui car précédées d'un `Get` scopé, mais fragiles (défense en profondeur).
- **Recommandation** : ajouter `AND user_id = $n` à ces `UPDATE` (ou documenter explicitement la dépendance au pré-contrôle).

#### B5 — Opportunité « privée » lisible par ID
- **Constat** : les opportunités créées par recherche sont masquées des listes publiques (`opportunities.sql:3`) mais `GetOpportunity` est global (`opportunities.sql:19-20`), utilisé par conversations/idées. Exfiltration limitée (UUID), mais accès par objet non respecté.
- **Recommandation** : si une opportunité a `user_id`, n'autoriser la lecture par ID qu'à son propriétaire (ou traiter ces lectures comme volontairement partagées, à documenter).

#### B6 — Écarts mineurs FE
- i18n : `escapeValue: false` global (`frontend/src/i18n/index.ts:78`) — sûr aujourd'hui (rendu React), à remettre par défaut ou documenter.
- `/login` et `/register` restent accessibles une fois connecté (pas de « GuestRoute ») — `login.tsx:3-4`, `register.tsx:3-4`.
- `state.from` de `ProtectedRoute` jamais consommé après login (UX).
- SSE : sur 401, la connexion s'arrête sans redirection (`frontend/src/lib/api/events.ts:97-100`).

#### B7 — Détails transport/divers
- Clé de chiffrement dérivée par **SHA-256** (non memory-hard) — `crypto.go:34-41` ; rotation multi-clés non finalisée (version stockée non résolue).
- Audit trail **best-effort** (échec d'écriture = simple slog, `audit/recorder.go:34-36`) et lacunes : confirmations d'idées, changements de config projet, réservations de crédits non journalisés.
- Rendu chromedp en Chrome headless `no-sandbox` via `data:` URL (`infra/render/chromedp.go:27-34`) : du contenu (LLM/utilisateur) contenant `<img>` pourrait déclencher des requêtes sortantes (SSRF aveugle) — surface faible, à documenter/mitiger (CSP dans la page HTML rendue).

---

## 5. Zoom sécurité des flux publics sensibles

### 5.1 Paiements (ADR-018) — correct
1. `POST /api/v1/payments/checkout` (JWT) → création `pending` + checkout provider, `return_url = FE/credits?payment=<id>`.
2. Redirection navigateur vers la page hébergée du provider (aucune carte/secret côté FE).
3. Retour navigateur `?payment=<id>` → `POST /payments/{id}/sync` → `VerifyStatus` provider → `applyStatus` (crédits octroyés **une seule fois** : ledger idempotent `payment:<id>`).
4. Webhook public → extrait référence → `VerifyStatus` par API (le corps n'est jamais une preuve) → même `applyStatus`.

Chaîne d'octroi idempotente et sourcée sur le provider : **bien conçu**. Axes restants : signature webhook (B1), rate limit (E1), statut de `PaymentRefunded` non traité par `applyStatus` (à vérifier).

### 5.2 OAuth intégrations publicitaires — correct
Callback public sans JWT, mais état CSRF usage unique (TTL 10 min) créé avec `user_id`, consommé atomiquement, tokens chiffrés AES-GCM au repos. Anti-replay confirmé.

---

## 6. Vérification d'ownership par ressource (synthèse IDOR)

| Ressource | Scopé `user_id` ? | Emplacement clé |
|---|---|---|
| Conversations (get/messages/title) | Oui | `chat/service.go:130-203` ; `conversations.sql:6-7,12-19` |
| Projets (get/config/générations) | Oui | `projects/service.go:47-137` ; `projects.sql:6-22` |
| Support tickets + messages (user) | Oui (contrôle service) | `support/service.go:60-118` |
| Paiements (get/list/sync) | Oui | `payments/service.go:66-188` ; `billing.sql:23-30` |
| Campagnes ads (pause/resume/insights/creatives) | Oui | `advertising/service.go:388-492` ; `advertising.sql:92-116` |
| Assets **download** | Oui | `asset_repo.go:34-43` ; `assets.sql:6-7` |
| Assets **liste par projet** | **Non** (M3) | `handler/assets.go:33-45` ; `assets.sql:9-10` |
| Jobs, idées, réservation crédits | Oui | `jobs.sql:6-7` ; `ideas.sql:12-13` ; `credit_repo.go:96-134` |

---

## 7. Feuille de route de remédiation (suggestion)

**Itération 1 — court terme (sécurité)**
1. Brancher le rate limiting (E1) + IP fiable (E2) — Redis si multi-replicas.
2. Middleware d'en-têtes de sécurité backend + `add_header` nginx (M1).
3. Scoper `ListByProject` à l'utilisateur (M3).
4. Timeouts serveur (M2).

**Itération 2 — session & contrôle d'accès FE**
5. Interceptor 401/403 + purge cache token + redirection login (M5).
6. Garde de rôle admin côté client + vue « accès refusé » (M5).

**Itération 3 — durcissement**
7. Audience JWT / recoupement claims (M4).
8. Signature webhooks (B1), `Content-Disposition` encodée (B2), magic-bytes uploads (B3).
9. `UPDATE` SQL scopés (B4), accès par ID des opportunités (B5).
10. Écarts mineurs FE (B6) et transport (B7).

---

## 8. Limites de l'audit

- Audit **statique** : aucune exécution de tests, pas de scan de dépendances (à compléter par `govulncheck` + `npm audit`/`pnpm audit`), pas de test d'intrusion dynamique.
- Les claims exacts émis par Neon Auth (présence d'un `aud`, politique de session/cookies) n'ont pas pu être confirmés chez le fournisseur.
- Le rate limiting et la protection anti-brute-force du login relèvent de la configuration **Neon Auth** (à activer côté fournisseur).
