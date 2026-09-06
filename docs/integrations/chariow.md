# Intégration commerce — Chariow (ADR-019)

> Chaque utilisateur connecte **sa propre boutique Chariow** à AfriLaunch via
> une **API key** (jamais de clé globale). L'intégration est multi-tenant :
> les appels Chariow utilisent toujours les credentials du bon utilisateur.

## Architecture

```
Application (domain/application)
   ↓  port.CommerceProvider + port.CommerceRepository
CommerceProvider (infra/commerce/chariow)   →  Chariow API
CommerceRepository (infra/postgres/commerce_repo.go)
CommerceService (application/commerce)       →  chiffre/valide/associe
```

Conçu pour accepter plus tard `StripeProvider`, `PaystackProvider`, etc. :
le domaine ne dépend jamais des détails HTTP Chariow.

## Configuration (env)

- `CHARIOW_API_BASE_URL` (défaut `https://api.chariow.com/v1`).
- `ENCRYPTION_KEY` : chiffre les clés API et secrets Pulse en base (AES-256-GCM).
- **Pas de `CHARIOW_API_KEY` globale** : la clé appartient à chaque utilisateur.

## Modèle de données (migration `00020_commerce.sql`)

- `commerce_connections` : 1 connexion/user (provider `chariow`), `api_key_enc`
  + `webhook_secret_enc` (chiffrés), statut `connected|disconnected|error|reconnect_required`,
  infos boutique (`external_store_id`, `store_name`, `store_url`), `last_sync_at`.
- `commerce_product_links` : association **projet SaaS ↔ produit Chariow existant**
  (`external_product_id`, snapshot prix, `is_public` + `public_token` pour la page
  d'achat publique `/buy/:token`).
- `commerce_sales` : ventes (`external_sale_id`, statut, montant minor, acheteur,
  `custom_metadata`, `payload`), mises à jour par webhook/sync (idempotent).
- `commerce_webhook_events` : idempotence des livraisons Pulse
  (`UNIQUE(provider, external_delivery_id)`).

## Connexion

1. UI : Intégrations → **Vente (Chariow)** → Connecter → instructions + champ clé
   (`sk_live_…`) et, optionnel, secret Pulse (`whsec_…`).
2. Backend : `POST /api/v1/commerce/chariow/connect` → **test réel** `GET /store`
   (jamais de validation « ça ressemble à une clé ») → chiffrement → sauvegarde.
3. Erreur claire si `401/403` (« Clé invalide »). La clé n'est **jamais** renvoyée
   au frontend, loggée, ni incluse dans une erreur/analytics/URL.

Endpoints protégés : `GET /commerce/status`, `GET /commerce/products`,
`GET|POST /commerce/links`, `PUT /commerce/links/{id}/public`,
`DELETE /commerce/links/{id}`, `GET /commerce/sales`, `POST /commerce/sync`,
`POST /commerce/chariow/disconnect`, `POST /commerce/chariow/webhook-secret`.

## Produits & « publication »

⚠️ **Limitation confirmée** : l'API publique Chariow est **lecture seule** sur les
produits (`GET /products`) — **aucune création/modification/suppression**.
Un produit généré par AfriLaunch n'est donc pas poussé automatiquement vers
Chariow : le marchand crée le produit (et son fichier) dans sa boutique Chariow,
puis l'**associe** à son projet SaaS (`POST /commerce/links`). La page d'achat
publique `/buy/:token` référence ce produit Chariow.

## Checkout (page d'achat publique)

- `GET /api/v1/commerce/public/{token}` : infos publiques (nom, prix, boutique).
- `POST /api/v1/commerce/public/{token}/checkout` : initie le checkout Chariow
  (email/prénom/nom/téléphone obligatoires), `redirect_url` vers `/buy/:token?thanks=1`,
  et renvoie `checkout_url`. `custom_metadata` = `saas_user_id/saas_project_id/saas_link_id`
  pour rattacher la vente au bon utilisateur/produit (multi-tenant).
- Le frontend n'embarque **aucune** clé Chariow.
- Statuts gérés : `completed`, `failed`, `abandoned`, `refunded` (via webhooks),
  `awaiting_payment`/`settled` (via sync).

## Webhooks (Pulse)

- Endpoint public : `POST /api/v1/commerce/webhook/chariow` (sans JWT).
- Config **manuelle** chez Chariow : Automations → Pulses → URL du webhook
  (affichée dans l'UI après connexion) + événements `successful.sale`,
  `failed.sale`, `abandoned.sale`. Le marchand fournit le secret `whsec_…`
  (chiffré en base).
- **Signature** : `x-chariow-signature: sha256=hmac_sha256(corps brut, secret)`
  vérifiée en temps constant sur le **corps brut** (jamais re-sérialisé).
- **Idempotence** : déduplication sur `x-pulse-delivery-id`
  (`commerce_webhook_events`). Retries Chariow : 5 tentatives, backoff exponentiel.
- Sans webhook configuré, les ventes restent récupérables via `POST /commerce/sync`
  (GET `/sales`, idempotent).

## Synchronisation

`POST /commerce/sync` : parcourt les ventes de la boutique et fait un
**upsert** idempotent. N'écrase jamais un `checkout_url`/metadata existant.

## n8n

Hors MVP. Le backend reste responsable du métier/sécurité ; n8n (branché plus
tard sur les événements internes/audit) ne doit pas devenir le cœur.

## Sécurité

- Clés API + secrets Pulse chiffrés AES-256-GCM au repos (`infra/crypto`).
- Isolation stricte : routes scopées `user_id` ; clés déchiffrées uniquement en
  mémoire le temps de l'appel ; jamais loggées/exposées.
- `401/403` Chariow → statut `reconnect_required` (l'historique est conservé).
- Déconnexion : efface les secrets, conserve ventes/mappings, interdit les appels.

## Erreurs

`CommerceRemoteError` (message sûr Chariow), sentinelles `domain.ErrCommerce*`
mappées en RFC 9457 côté handler. Aucune réponse brute Chariow renvoyée telle quelle.

## Limitations Chariow (documentées — ne rien inventer)

- **Pas de création/modification de produit** via l'API publique (association
  uniquement, cf. Produits).
- **Pas de création de Pulse/webhook** via l'API (configuration manuelle).
- Pas de checkout API pour produits `service`, `coaching`, `pay-what-you-want`.
- Pas d'OAuth tiers / Partner API confirmé (modèle clé API, architecture
  prête à migrer). Pas de création de compte/boutique via API.
- Pas de marketplace/split settlement. Répétition d'achat bloquée
  (`already_purchased`) si accès actif pour downloadable/course/bundle.
- Rate limit Chariow : **100 req/min/clé**.

## Ajouter un autre provider de commerce

1. Implémenter `port.CommerceProvider` (`infra/commerce/<provider>`).
2. Étendre `commerce_connections.provider` (check + valeurs) via une migration.
3. Enregistrer le provider dans `cmd/api/main.go` (registry) et ajuster l'UI.
