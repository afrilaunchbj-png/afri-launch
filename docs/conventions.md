# Conventions — AfriLaunch

> Conventions obligatoires. Toute nouvelle feature doit s'y conformer.

## 1. Internationalisation (i18n)

- **react-i18next**, langues `fr` (défaut) et `en`.
- Locales centralisées dans `frontend/src/i18n/locales/{fr,en}/*.json`, découpées par domaine (`common`, `auth`, `credits`, `opportunities`, …).
- **Aucun texte utilisateur hardcodé** dans les composants. Toujours `t('namespace.key')`.
- Les erreurs, labels, placeholders, états vides, notifications et aria-labels sont aussi traduits.
- Ajouter une clé : l'ajouter dans **les deux** fichiers `fr` et `en` (pas de clé manquante).

## 2. Dark mode

- Stratégie Tailwind `class` (classe `.dark` sur `<html>`).
- Toggle via `ThemeProvider` ; persistance `localStorage` ; respect de `prefers-color-scheme` au premier chargement.
- Couleurs via **tokens CSS** (design system « Emerald & Amber Ledger ») uniquement — pas de couleurs hex hardcodées dans les composants.
- Chaque composant définit ses variantes `dark:` (background, texte, bordures). Contraste WCAG AA (4.5:1 texte, 3:1 UI) dans les deux modes.

## 3. Composants UI (shadcn/ui + Radix)

- **Zéro composant HTML natif** quand un équivalent shadcn/Radix existe (pas de `<select>`, `<input type="checkbox">` custom, dropdown custom, dialog custom, tooltip custom…).
- Composants de base dans `frontend/src/components/ui/` (générés par shadcn/ui).
- Composants applicatifs = **wrappers/compositions** de shadcn, jamais un 2e système UI.
- Variants définis avec **CVA** + `cn()` (clsx + tailwind-merge) pour fusionner `className`.

## 4. Icônes

- **lucide-react uniquement**. Aucune autre lib d'icônes, pas de SVG custom si un équivalent Lucide existe.

Mapping des Material Symbols (utilisées dans les maquettes) vers Lucide :

| Material Symbol | Lucide |
|---|---|
| `rocket_launch` | `Rocket` |
| `search_insights` / `insights` | `LineChart` / `TrendingUp` |
| `lightbulb` | `Lightbulb` |
| `edit_document` | `FilePen` |
| `campaign` | `Megaphone` |
| `storefront` | `Store` |
| `payments` | `CreditCard` |
| `monetization_on` / `toll` | `Coins` |
| `account_balance_wallet` | `Wallet` |
| `notifications` | `Bell` |
| `settings` | `Settings` |
| `help` | `HelpCircle` |
| `add` / `add_circle` | `Plus` / `PlusCircle` |
| `visibility` / `visibility_off` | `Eye` / `EyeOff` |
| `mail` | `Mail` |
| `lock` | `Lock` |
| `person` | `User` |
| `search` | `Search` |
| `tune` | `SlidersHorizontal` |
| `verified` | `BadgeCheck` |
| `location_on` | `MapPin` |
| `article` | `FileText` |
| `trending_up` | `TrendingUp` |
| `trending_down` | `TrendingDown` |
| `moving` | `Activity` |
| `bookmark` / `bookmark_border` | `Bookmark` |
| `auto_awesome` | `Sparkles` |
| `phone_iphone` | `Smartphone` |
| `verified_user` | `ShieldCheck` |
| `flash_on` | `Zap` |
| `money_off` | `BadgeDollarSign` |
| `check` | `Check` |
| `chevron_right` | `ChevronRight` |
| `menu` | `Menu` |
| `close` | `X` |
| `database` | `Database` |
| `workspace_premium` | `Crown` |
| `download` | `Download` |
| `calendar_month` | `Calendar` |
| `menu_book` | `BookOpen` |
| `credit_card` | `CreditCard` |
| `format_bold/italic/h1/list_bulleted` | `Bold`/`Italic`/`Heading1`/`List` |
| `spellcheck` | `SpellCheck` |
| `check_circle` | `CheckCircle2` |
| `sync` | `RefreshCw` |
| `arrow_forward` | `ArrowRight` |
| `public` | `Globe` |
| `handyman` | `Wrench` |
| `account_circle` | `CircleUser` |

## 5. Gestion des erreurs

### Backend (Go)

- Type d'erreur centralisé `internal/server/apierror` conforme **RFC 9457** :
  ```go
  type APIError struct { Type, Title string; Status int; Detail string; Errors []FieldError }
  ```
- Distinguer : `validation` (422), `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` (409), `business` (422/409), `internal` (500).
- Middleware unique de conversion erreur → réponse JSON ; **jamais** de stacktrace/message SQL exposé.
- Logs techniques (détail) côté serveur ; message utilisateur orienté action, traduisible.

### Frontend (React)

- Un **api client** (`src/lib/api`) qui convertit les réponses Problem Details en une `AppError` typée.
- Affichage selon le contexte :
  - **inline** (champ/section) → composant `FieldError` ;
  - **toast** (événement/erreur global) → `sonner` (shadcn) ;
  - **page/état dédié** (`ErrorState`) quand la page ne peut pas s'afficher.
- Messages orientés action, jamais `undefined`/`Something went wrong`/`Internal Server Error`.

## 6. Pattern unique de formulaire

- Bibliothèques : **react-hook-form** + **zod** (`zodResolver`), composant shadcn `<Form>` (FormField, FormItem, FormLabel, FormControl, FormMessage).
- Structure imposée :
  1. `zod` schema (co-localisé) = source de vérité de validation.
  2. `useForm<z.infer<Schema>>` + `zodResolver`.
  3. Champs obligatoires marqués `*` (via `required` dans le label).
  4. Chaque champ = `FormField` → `FormItem` → `FormLabel` + `FormControl` + `FormMessage` (erreur inline).
  5. Bouton submit : `disabled={form.formState.isSubmitting}` + spinner pendant l'envoi.
  6. `onSubmit` → `useMutation` (TanStack Query) ; succès = toast + invalidation + reset/navigation ; échec = erreur inline/toast.
- Validation **frontend ET backend** (jamais d'envoi de données manifestement invalides).

## 7. Pattern listes / pagination / filtres

- **Listes volumineuses** (historiques, transactions, opportunités, utilisateurs…) : pagination **serveur** (TanStack Query + `useQuery`).
- Composant `DataTable` basé sur **@tanstack/react-table** + shadcn `Table` :
  - pagination (`page`, `pageSize`, total) ;
  - tri (`sort`) ;
  - filtres explicites (Select/Checkbox/Command de shadcn) ;
  - état de filtres réinitialisable (« Réinitialiser les filtres ») ;
  - états **loading** (squelettes), **empty** (guidé + CTA), **error** (message + Retry).
- Filtres + pagination doivent fonctionner ensemble et être **prévisibles** ; état conservé dans l'URL (`?page=&status=…`).
- Pagination API : cursor-based pour les grosses listes (curseur opaque), offset pour les petites.

## 8. Conventions Go

- Layout standard (cf. `architecture.md`) ; module unique `backend/`.
- `gofmt` + `go vet` ; noms idiomatiques (camelCase, acronymes en majuscules : `HTTP`, `ID`).
- Erreurs : enveloppées avec `fmt.Errorf("...: %w", err)` ; `errors.Is`/`errors.As` pour le typage.
- Tests : table-driven, `testify` (assert/require), `*_test.go` co-localisés.
- Aucune dépendance framework dans `domain/`.

## 9. Conventions React/TypeScript

- TypeScript **strict** ; `@/` alias → `src/`.
- Props typées via `interface` ; jamais `any` (utiliser `unknown` + narrowing).
- Organisation **feature-based** : `src/features/<domaine>/{components,hooks,types,api}.ts` + barrel `index.ts`.
- Hooks custom préfixés `use` ; `useQuery`/`useMutation` enveloppés dans des hooks dédiés (jamais appelés directement dans les composants).
- Composant par fichier ; tests co-localisés (`*.test.tsx`).
- Nommage : PascalCase composants, camelCase hooks/fonctions, kebab-case dossiers/fichiers.

## 10. Sécurité

- **Frontend jamais une frontière de sécurité** : toute autorisation est revérifiée côté serveur.
- JWT en cookies httpOnly/Secure/SameSite ; jamais de token en localStorage.
- Validation des entrées des deux côtés ; requêtes paramétrées (sqlc) ; anti XSS (encodage React + CSP) ; CSRF (SameSite + Origin check).
- Rate limiting (login strict, API modéré) ; uploads validés (type/size/magic bytes, noms générés serveur).
- Secrets via env vars ; `.env` ignoré ; `.env.example` fourni ; aucun secret commité.
- Isolation tenant : chaque requête filtre par `user_id`/`organization_id` (RLS PostgreSQL en filet de sécurité).

## 11. Accessibilité & responsive

- Mobile-first ; touch targets ≥ 48px ; breakpoints Tailwind (`sm/md/lg/xl`).
- `focus-visible:` (pas `focus:`) ; `sr-only` pour les boutons icône ; `aria-label`/`aria-labelledby` sur les inputs.
- `prefers-reduced-motion` respecté.
