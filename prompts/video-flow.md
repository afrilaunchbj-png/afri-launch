# Mission : construire un pipeline SaaS de génération automatique de vidéos publicitaires avec IA

Tu es un ingénieur logiciel senior spécialisé en SaaS, architecture backend, IA générative et systèmes de traitement asynchrones.

Je veux construire un module de mon SaaS permettant à un utilisateur de transformer la description d'un **produit digital** en une ou plusieurs **vidéos publicitaires prêtes pour TikTok, Instagram Reels, Facebook et WhatsApp**.

L'objectif est de construire une architecture propre, extensible et provider-agnostic.

---

# 1. Concept produit

L'utilisateur fournit :

* nom du produit ;
* description ;
* bénéfices ;
* audience cible ;
* prix ;
* CTA ;
* image/mockup du produit ;
* éventuellement quelques informations supplémentaires.

Le système doit ensuite pouvoir générer automatiquement une publicité vidéo.

Exemple :

```text
Produit :
"Guide pour lancer son business en ligne au Bénin"

Audience :
Jeunes entrepreneurs africains

Prix :
5 000 FCFA

CTA :
"Achète le guide maintenant"

Image :
mockup du PDF
```

Le système transforme cela en :

```text
Produit
   ↓
Analyse marketing par LLM
   ↓
Hooks
   ↓
Script publicitaire
   ↓
Storyboard
   ↓
Découpage en scènes
   ↓
Génération des assets
   ├── Avatar parlant → HeyGen
   ├── B-roll → Veo / Kling / autre provider
   ├── Images → image generation provider
   └── Voice-over → TTS si nécessaire
   ↓
Montage
   ↓
Sous-titres
   ↓
Vidéo finale
   ↓
Export 9:16 / 1:1 / 16:9
```

---

# 2. Principe architectural important

NE PAS construire le système directement autour d'un provider unique.

Je veux une architecture avec des abstractions :

```text
VideoProvider
ImageProvider
AvatarProvider
VoiceProvider
LLMProvider
StorageProvider
```

Par exemple :

```typescript
interface VideoProvider {
  generateVideo(input: VideoGenerationInput): Promise<VideoGenerationJob>;
  getJobStatus(jobId: string): Promise<VideoGenerationResult>;
  cancelJob?(jobId: string): Promise<void>;
}
```

Puis :

```text
providers/
  video/
    kling/
    veo/
    runway/

  avatar/
    heygen/

  image/
    ...

  voice/
    ...

  llm/
    ...
```

Le code métier ne doit jamais dépendre directement de HeyGen, Kling ou Veo.

---

# 3. Pipeline de génération

Créer un pipeline asynchrone.

Une génération complète doit être représentée par un job :

```text
GenerationJob
```

avec un état du type :

```text
PENDING
ANALYZING
SCRIPT_GENERATED
STORYBOARD_GENERATED
ASSETS_GENERATING
ASSETS_READY
RENDERING
COMPLETED
FAILED
CANCELLED
```

Le pipeline doit pouvoir reprendre après une erreur.

Il ne faut surtout pas avoir une requête HTTP qui reste ouverte pendant plusieurs minutes en attendant la génération vidéo.

---

# 4. Étape 1 — analyse marketing

Utiliser un LLM pour analyser le produit.

Entrée :

```json
{
  "productName": "...",
  "description": "...",
  "targetAudience": "...",
  "benefits": [],
  "price": "...",
  "cta": "..."
}
```

Le LLM doit produire une structure strictement typée :

```json
{
  "targetAudience": "...",
  "painPoints": [],
  "desiredOutcome": "...",
  "valueProposition": "...",
  "marketingAngles": [],
  "hooks": []
}
```

Générer plusieurs angles marketing.

Exemples :

```text
Pain point
Transformation
Curiosity
Social proof
Urgency
Educational
Problem/Solution
```

---

# 5. Étape 2 — génération du script

Pour chaque angle, générer un script court.

Le script doit être adapté aux formats short-form.

Par exemple :

```text
0-3s
HOOK

3-8s
PROBLEM

8-18s
SOLUTION

18-25s
BENEFITS

25-30s
CTA
```

Le système doit pouvoir générer :

* 15 secondes ;
* 30 secondes ;
* 45 secondes ;
* éventuellement 60 secondes.

La durée doit être configurable.

---

# 6. Étape 3 — storyboard

Le LLM transforme le script en storyboard structuré.

Exemple :

```json
{
  "duration": 30,
  "aspectRatio": "9:16",
  "scenes": [
    {
      "id": "scene_1",
      "duration": 3,
      "type": "avatar",
      "script": "...",
      "visualDescription": "...",
      "textOverlay": "...",
      "emotion": "energetic"
    },
    {
      "id": "scene_2",
      "duration": 5,
      "type": "broll",
      "visualDescription": "...",
      "textOverlay": "..."
    }
  ]
}
```

Les types de scènes doivent être extensibles :

```text
avatar
broll
product
image
screen_recording
text
transition
```

---

# 7. HeyGen

HeyGen doit être utilisé pour les scènes où un présentateur/avatar parle face caméra.

Exemple :

```text
scene.type === "avatar"
```

Le provider HeyGen doit être encapsulé derrière :

```typescript
interface AvatarProvider {
  createVideo(input: AvatarVideoInput): Promise<AvatarGenerationJob>;
  getJobStatus(jobId: string): Promise<AvatarGenerationResult>;
}
```

Le système doit permettre de configurer :

```text
avatar
voice
language
accent
tone
background
aspectRatio
resolution
```

Ne hard-code aucun avatar.

Les avatars et voix doivent être récupérables depuis le provider et/ou configurables dans notre base de données.

---

# 8. Génération B-roll

Pour les scènes `broll`, utiliser un provider vidéo externe.

Prévoir une abstraction :

```typescript
interface VideoProvider {
  generateVideo(...)
  getJobStatus(...)
}
```

Le premier provider peut être Kling, Veo ou un autre provider disponible.

IMPORTANT :

Ne jamais mettre les clés API directement dans le code.

Utiliser :

```text
environment variables
secret manager
configuration service
```

---

# 9. Génération d'images

Certaines scènes pourront utiliser :

* image du produit ;
* mockup ;
* image générée ;
* composition graphique.

Créer une abstraction :

```typescript
interface ImageProvider {
  generateImage(...)
}
```

---

# 10. Voice-over

Prévoir également :

```typescript
interface VoiceProvider {
  generateSpeech(...)
  getJobStatus(...)
}
```

Même si HeyGen fournit déjà une voix pour les avatars.

Cela permettra d'avoir des publicités sans avatar.

---

# 11. Montage vidéo

Une fois tous les assets disponibles :

```text
scene 1
scene 2
scene 3
...
```

les assembler dans un renderer.

Privilégier :

```text
FFmpeg
```

ou éventuellement :

```text
Remotion
```

Le renderer doit gérer :

* concaténation ;
* transitions ;
* images ;
* vidéos ;
* texte ;
* sous-titres ;
* musique ;
* voice-over ;
* logo ;
* CTA ;
* branding ;
* aspect ratio ;
* résolution.

Prévoir plusieurs presets :

```text
TikTok/Reels:
1080x1920

Square:
1080x1080

Landscape:
1920x1080
```

---

# 12. Sous-titres

Les vidéos doivent pouvoir générer automatiquement des sous-titres.

Format interne :

```text
SRT
ou
WebVTT
```

Le renderer doit pouvoir incruster les sous-titres directement dans la vidéo.

Prévoir plusieurs styles :

```text
classic
bold
minimal
highlight
karaoke
```

---

# 13. Musique

Prévoir un système de background music.

Le produit doit pouvoir sélectionner :

```text
energetic
calm
corporate
emotional
afro
modern
```

Attention aux droits d'utilisation.

Ne jamais utiliser automatiquement des musiques commerciales sans licence.

Prévoir une bibliothèque d'assets sous licence ou libres d'utilisation.

---

# 14. Architecture backend

Le système doit être asynchrone.

Exemple :

```text
API
 │
 ▼
Create Generation Job
 │
 ▼
PostgreSQL
 │
 ▼
BullMQ / Redis
 │
 ├── marketing worker
 ├── script worker
 ├── storyboard worker
 ├── avatar worker
 ├── video worker
 ├── image worker
 ├── voice worker
 └── rendering worker
 │
 ▼
Object Storage
 │
 ▼
CDN
```

Chaque étape doit être idempotente.

Si un worker tombe après la génération d'un asset, il ne doit pas forcément régénérer l'asset.

---

# 15. Modèle de données

Proposer un modèle de données propre.

Prévoir au minimum :

```text
User
Product
AdCampaign
GenerationJob
GenerationScene
GeneratedAsset
ProviderJob
VideoRender
Voice
Avatar
Template
```

Relations :

```text
Product
   │
   └── AdCampaign
           │
           └── GenerationJob
                   │
                   ├── GenerationScene
                   │       └── GeneratedAsset
                   │
                   └── VideoRender
```

Ajouter les timestamps et métadonnées nécessaires.

---

# 16. Templates publicitaires

Le système doit supporter des templates.

Exemples :

```text
UGC
Avatar Expert
Problem → Solution
Product Showcase
Testimonial
Educational
Storytelling
Offer
Before → After
```

Un template doit pouvoir définir :

```text
duration
scene structure
prompt strategy
avatar usage
B-roll usage
text overlays
CTA
music
```

Exemple :

```json
{
  "name": "UGC Product",
  "duration": 30,
  "scenes": [
    "hook",
    "problem",
    "solution",
    "benefits",
    "cta"
  ]
}
```

---

# 17. Génération multi-variantes

Une fonctionnalité importante est de pouvoir demander :

```text
Generate 5 ads
```

Le système doit alors produire différentes variantes :

```text
Ad #1 — Curiosity
Ad #2 — Problem/Solution
Ad #3 — Transformation
Ad #4 — Social proof
Ad #5 — Urgency
```

Les vidéos ne doivent pas être de simples copies.

Le LLM doit varier :

* hook ;
* script ;
* angle ;
* scènes ;
* CTA ;
* présentation.

---

# 18. Scoring automatique

Après génération, prévoir la possibilité de faire analyser les publicités par un LLM.

Score :

```text
Hook: 8/10
Clarity: 9/10
Persuasion: 7/10
CTA: 9/10
Visual quality: 8/10
Target relevance: 9/10
```

Puis :

```text
overallScore
```

Cela permettra plus tard de sélectionner automatiquement les meilleures variantes.

---

# 19. Gestion des erreurs

Chaque provider externe peut :

* timeout ;
* retourner une erreur ;
* être temporairement indisponible ;
* avoir un rate limit ;
* produire un résultat invalide.

Prévoir :

```text
retry
exponential backoff
dead-letter queue
provider fallback
timeout
circuit breaker si nécessaire
```

Ne jamais perdre un job.

Chaque appel externe doit être traçable.

---

# 20. Observabilité

Ajouter des logs structurés.

Pour chaque génération :

```text
generationId
userId
provider
providerJobId
sceneId
duration
status
error
cost
createdAt
completedAt
```

Prévoir également le calcul du coût estimé :

```text
LLM cost
video generation cost
avatar cost
voice cost
storage cost
rendering cost
```

Le système doit permettre de connaître :

```text
cost per video
cost per campaign
cost per user
```

C'est extrêmement important pour le business model du SaaS.

---

# 21. API

Concevoir des endpoints propres.

Par exemple :

```http
POST /api/v1/products
POST /api/v1/ad-campaigns
POST /api/v1/ad-campaigns/:id/generate

GET /api/v1/generations/:id
GET /api/v1/generations/:id/scenes
GET /api/v1/generations/:id/assets

POST /api/v1/generations/:id/cancel

GET /api/v1/videos/:id
```

La création d'une génération doit répondre rapidement :

```json
{
  "id": "generation_123",
  "status": "PENDING"
}
```

Le frontend peut ensuite utiliser :

```text
polling
SSE
WebSocket
```

pour suivre la progression.

---

# 22. Sécurité

Appliquer les bonnes pratiques :

* validation stricte des inputs ;
* rate limiting ;
* authentication ;
* authorization ;
* isolation des fichiers utilisateurs ;
* signed URLs ;
* secrets uniquement côté serveur ;
* validation des fichiers uploadés ;
* limites de taille ;
* protection SSRF ;
* sanitation des prompts ;
* quotas par utilisateur ;
* contrôle des coûts.

Un utilisateur ne doit jamais pouvoir accéder aux assets d'un autre utilisateur.

---

# 23. Stockage

Utiliser un object storage compatible S3 ou GCS.

Organisation recommandée :

```text
/users/{userId}/products/{productId}/
/users/{userId}/campaigns/{campaignId}/
/users/{userId}/generations/{generationId}/
/users/{userId}/generations/{generationId}/scenes/
```

Ne pas stocker les vidéos directement dans PostgreSQL.

---

# 24. Architecture frontend

Le frontend doit permettre :

```text
Create Product
      ↓
Choose Ad Template
      ↓
Configure Audience
      ↓
Choose Avatar
      ↓
Choose Voice
      ↓
Choose Style
      ↓
Generate
      ↓
Generation Progress
      ↓
Preview
      ↓
Edit
      ↓
Regenerate Scene
      ↓
Export
```

L'utilisateur doit pouvoir régénérer **une seule scène** sans devoir régénérer toute la publicité.

Exemple :

```text
Scene 3
[Regenerate]

Scene 4
[Change prompt]

Scene 5
[Change avatar]
```

---

# 25. Important : architecture provider-agnostic

Je veux pouvoir remplacer :

```text
HeyGen
```

par :

```text
Provider B
```

sans modifier toute la logique métier.

Même chose pour :

```text
Kling
Veo
Runway
```

Le système doit donc avoir un mécanisme de configuration :

```text
VIDEO_PROVIDER=kling
AVATAR_PROVIDER=heygen
IMAGE_PROVIDER=...
VOICE_PROVIDER=...
```

ou mieux, une configuration dynamique par projet/organisation.

---

# 26. UX importante

La génération vidéo peut prendre plusieurs minutes.

Le frontend doit afficher une progression :

```text
✓ Analyse du produit
✓ Génération du script
✓ Création du storyboard
✓ Génération de l'avatar
⏳ Génération du B-roll
○ Montage
○ Finalisation
```

Ne jamais afficher simplement :

```text
Loading...
```

---

# 27. MVP

NE PAS implémenter tout le système immédiatement.

Commencer par un MVP fonctionnel :

```text
Product
   ↓
LLM
   ↓
Script
   ↓
Storyboard
   ↓
HeyGen
   ↓
FFmpeg
   ↓
Final video
```

Le MVP doit permettre :

1. créer un produit ;
2. générer un script ;
3. générer un storyboard ;
4. générer une vidéo avec HeyGen ;
5. récupérer la vidéo ;
6. ajouter éventuellement le mockup du produit ;
7. ajouter les sous-titres ;
8. produire une vidéo 9:16 ;
9. stocker le résultat ;
10. afficher le résultat dans le frontend.

Ensuite seulement ajouter :

```text
Veo/Kling
B-roll
multiple variants
music
scoring
A/B testing
provider fallback
```

---

# 28. Méthode de travail demandée

Avant d'écrire du code :

1. inspecte entièrement le repository ;
2. comprends l'architecture existante ;
3. identifie les conventions déjà présentes ;
4. identifie les skills disponibles ;
5. identifie les composants réutilisables ;
6. ne réinvente pas ce qui existe déjà ;
7. propose l'architecture avant l'implémentation ;
8. identifie les risques techniques ;
9. définis les interfaces ;
10. puis implémente le MVP.

Si des conventions existent déjà dans le projet, elles sont prioritaires sur mes exemples ci-dessus.

---

# 29. Qualité du code

Je veux :

* TypeScript strict ;
* code modulaire ;
* SOLID lorsque pertinent ;
* tests unitaires ;
* tests d'intégration ;
* validation des inputs ;
* gestion d'erreurs explicite ;
* logs structurés ;
* documentation des interfaces ;
* migrations propres ;
* aucune clé API hard-codée ;
* aucune logique provider dans le domaine métier.

Éviter :

```text
god classes
god services
huge controllers
magic strings
duplicated provider logic
```

---

# 30. Critères d'acceptation MVP

À la fin du MVP, je dois pouvoir faire :

```text
1. Créer un produit digital
2. Choisir "UGC / Avatar Ad"
3. Cliquer Generate
4. Le backend crée un GenerationJob
5. Le worker génère le script
6. Le worker génère le storyboard
7. HeyGen génère l'avatar
8. Le système récupère automatiquement la vidéo
9. FFmpeg ajoute le produit / texte / sous-titres
10. La vidéo finale est uploadée dans object storage
11. Le frontend reçoit le statut COMPLETED
12. L'utilisateur peut regarder la vidéo
13. L'utilisateur peut télécharger/exporter la vidéo
14. L'utilisateur peut régénérer une scène
```

---

# 31. Première tâche

NE COMMENCE PAS immédiatement à coder.

Commence par :

### Phase 1 — Audit

Analyse le repository et explique :

* architecture actuelle ;
* stack ;
* structure des dossiers ;
* système de queue existant ;
* système de stockage existant ;
* système d'auth existant ;
* conventions ;
* composants réutilisables.

### Phase 2 — Architecture

Propose :

* architecture du module ;
* modèle de données ;
* interfaces providers ;
* jobs BullMQ ;
* workflow de génération ;
* stratégie de stockage ;
* gestion des erreurs ;
* stratégie de coûts.

### Phase 3 — Plan d'implémentation

Découpe le travail en petites étapes indépendantes et testables.

### Phase 4 — Implémentation

Implémente progressivement le MVP.

Après chaque étape :

* lance les tests ;
* vérifie les types ;
* vérifie le lint ;
* vérifie les migrations ;
* corrige les erreurs ;
* documente les décisions importantes.

Ne fais aucune grosse réécriture inutile du projet existant.

L'objectif est d'intégrer ce pipeline proprement dans l'architecture actuelle du projet, pas de créer un système parallèle.
