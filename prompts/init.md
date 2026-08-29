# Initialisation du projet — Instructions principales

Tu es l'agent principal chargé d'initialiser, structurer et développer ce projet.

Avant de modifier ou créer du code, tu dois d'abord comprendre complètement le projet, ses conventions, ses skills disponibles, son design et son architecture documentaire.

## 1. Sources de vérité du projet

Le projet contient plusieurs sources de référence que tu dois impérativement prendre en compte.

Tu dois inspecter et utiliser :

1. **Tous les skills disponibles dans le projet**

   * Identifie tous les skills disponibles.
   * Lis ceux qui sont pertinents pour la tâche.
   * Lorsque plusieurs skills sont applicables, utilise-les conjointement.
   * Ne réinvente pas des conventions ou des pratiques déjà définies par les skills.

2. **Le master prompt**

   * Lis obligatoirement :
     `prompts/master.md`
   * Considère son contenu comme une source de vérité importante pour le fonctionnement du projet.

3. **Les designs Stitch**

   * Inspecte entièrement le dossier :
     `design/`
   * Analyse les maquettes avant de commencer l'implémentation frontend.
   * Les maquettes Stitch constituent la référence visuelle et fonctionnelle de l'interface.
   * Ne te contente pas de reproduire visuellement les écrans : déduis également les composants, états, interactions, formulaires, tableaux, filtres, navigation et comportements nécessaires.

4. **La documentation**

   * Inspecte entièrement le dossier :
     `docs/`
   * Complète tous les fichiers existants dans `docs/`.
   * Si certains documents sont incomplets, transforme-les en documentation exploitable.
   * Si des informations sont nécessaires mais absentes, déduis-les à partir du master prompt, des designs, des skills et de l'architecture du projet.
   * Ne laisse pas volontairement de TODO ou de sections vides lorsque l'information peut être déterminée.

5. **Le code existant**

   * Inspecte la structure actuelle du repository avant toute modification.
   * Identifie ce qui existe déjà, ce qui est réutilisable et ce qui doit être créé.
   * Ne supprime ou ne remplace pas arbitrairement du travail existant.

---

# 2. AGENT.md — mémoire persistante obligatoire

Crée à la racine du projet un fichier :

`AGENT.md`

Ce fichier constitue la **mémoire de travail persistante de l'agent**.

Il doit permettre à un autre agent de reprendre le projet exactement là où le précédent s'est arrêté, sans perdre le contexte.

## Règle absolue

**Avant chaque nouvelle tâche, tu DOIS lire `AGENT.md`.**

Cette règle est obligatoire, même si la tâche semble simple.

Tu dois également **mettre `AGENT.md` à jour à chaque fois que ton travail modifie significativement l'état du projet.**

Le fichier doit notamment contenir :

* l'objectif global du projet ;
* l'état actuel du projet ;
* l'architecture actuelle ;
* les décisions techniques importantes ;
* les conventions importantes ;
* les fonctionnalités terminées ;
* les fonctionnalités en cours ;
* les fonctionnalités restantes ;
* les fichiers importants ;
* les problèmes connus ;
* les décisions prises et leur justification ;
* les migrations effectuées ;
* les prochaines étapes ;
* les éventuels blocages ;
* les tests effectués ;
* les tests restant à effectuer.

Il doit être suffisamment précis pour qu'un nouvel agent puisse reprendre immédiatement le travail.

### Format recommandé

```md
# AGENT.md

## Project Overview

## Current Status

## Architecture

## Tech Stack

## Important Decisions

## Design & UI Conventions

## Completed Work

## Work In Progress

## Remaining Work

## Known Issues

## Tests & Validation

## Database & Migrations

## Important Files

## Next Steps

## Notes for the Next Agent
```

Ne transforme pas `AGENT.md` en journal inutilement verbeux.

Il doit être une **source opérationnelle de vérité**, maintenue à jour.

---

# 3. Stack technique obligatoire

## Backend

* Go
* PostgreSQL
* Redis

## Frontend

* React
* React Router 7
* TypeScript
* Tailwind CSS
* shadcn/ui
* Radix UI

**Next.js est strictement interdit.**

Le frontend doit être une application React utilisant React Router 7.

---

# 4. Règles frontend non négociables

## 4.1 Internationalisation

L'application doit être entièrement disponible en :

* français ;
* anglais.

Toutes les pages, tous les composants, tous les formulaires, tous les messages, toutes les erreurs, tous les boutons, tous les labels, tous les états vides et toutes les notifications doivent être traduisibles.

**Aucun texte utilisateur ne doit être hardcodé dans les composants.**

L'architecture d'internationalisation doit être centralisée et cohérente.

---

## 4.2 Dark mode

Le dark mode doit être supporté automatiquement.

Tous les composants et toutes les pages doivent fonctionner correctement en :

* light mode ;
* dark mode.

Ne crée jamais une interface qui fonctionne uniquement dans un thème.

Les couleurs doivent utiliser le système de tokens/thèmes de shadcn/Tailwind plutôt que des couleurs hardcodées lorsque cela est pertinent.

---

## 4.3 Composants UI

**Tous les composants UI doivent être basés sur shadcn/ui et Radix UI.**

Cela signifie notamment que les composants suivants doivent utiliser les primitives appropriées :

* Button
* Input
* Select
* Dropdown Menu
* Dialog
* Alert Dialog
* Popover
* Tooltip
* Tabs
* Checkbox
* Radio Group
* Switch
* Form
* Sheet
* Command
* Calendar
* etc.

### Règle absolue

**0 composant HTML natif lorsqu'un composant shadcn/Radix approprié existe.**

Par exemple :

* pas de `<select>` natif → utiliser Select ;
* pas de dropdown custom → utiliser DropdownMenu ;
* pas de dialog custom → utiliser Dialog ;
* pas de tooltip custom → utiliser Tooltip.

Les composants spécifiques à l'application doivent être construits comme des **wrappers/compositions des composants shadcn/Radix**, et non comme un deuxième système UI indépendant.

Avant de créer un composant UI custom, vérifie d'abord si shadcn/ui ou Radix fournit déjà la primitive nécessaire.

---

# 5. Icônes

Toutes les icônes de l'application doivent provenir de :

**Lucide Icons**

Aucune autre bibliothèque d'icônes ne doit être utilisée.

Ne crée pas d'icônes SVG custom lorsqu'une icône Lucide correspondante existe.

---

# 6. Gestion des erreurs

Toutes les erreurs doivent être gérées de manière centralisée et cohérente.

L'application doit distinguer au minimum :

* erreurs de validation ;
* erreurs réseau ;
* erreurs API ;
* erreurs d'authentification ;
* erreurs d'autorisation ;
* erreurs métier ;
* erreurs inattendues.

Les messages affichés à l'utilisateur doivent être :

* clairs ;
* compréhensibles ;
* traduisibles ;
* orientés vers l'action lorsque cela est possible.

Une erreur doit être affichée au bon endroit :

* **inline** pour les erreurs liées à un champ ou à une section ;
* **toast** pour les événements/erreurs globaux ;
* **page d'erreur / état dédié** lorsque l'erreur empêche l'affichage de la page.

Évite absolument les messages techniques tels que :

```text
Internal Server Error
undefined
Something went wrong
TypeError: ...
```

s'ils sont directement exposés à l'utilisateur.

Les détails techniques doivent rester accessibles dans les logs destinés aux développeurs.

---

# 7. Listes, tableaux et pagination

Toutes les listes potentiellement volumineuses doivent être paginées.

Cela concerne notamment :

* tableaux ;
* listes d'utilisateurs ;
* historiques ;
* transactions ;
* résultats de recherche ;
* logs ;
* ressources ;
* etc.

La pagination doit être cohérente dans toute l'application.

Les filtres doivent :

* être compréhensibles ;
* avoir des labels explicites ;
* utiliser les composants shadcn/Radix ;
* être cohérents d'une page à l'autre ;
* être facilement réinitialisables ;
* fonctionner correctement avec la pagination.

Lorsqu'une combinaison de filtres modifie les résultats, le comportement doit être prévisible.

---

# 8. Formulaires

Tous les formulaires doivent suivre **un seul et même pattern dans toute l'application**.

Le pattern doit définir :

* structure ;
* labels ;
* champs obligatoires ;
* validation ;
* messages d'erreur ;
* états loading ;
* état disabled ;
* soumission ;
* succès ;
* erreurs ;
* reset ;
* accessibilité.

Les champs obligatoires doivent être explicitement identifiés.

La validation doit être effectuée côté frontend et côté backend lorsque nécessaire.

Un formulaire ne doit jamais permettre l'envoi de données manifestement invalides.

Avant d'implémenter plusieurs formulaires, définis un pattern réutilisable et utilise-le partout.

---

# 9. Sécurité

La sécurité est une exigence fondamentale du projet.

Tu dois appliquer les bonnes pratiques de sécurité aussi bien côté frontend que backend.

Cela inclut notamment :

* validation des entrées ;
* sanitation lorsque nécessaire ;
* authentification sécurisée ;
* autorisation côté serveur ;
* protection des données sensibles ;
* gestion sécurisée des sessions/tokens ;
* protection contre les injections SQL ;
* protection des endpoints ;
* contrôle des permissions ;
* limitation des informations exposées au frontend ;
* gestion sécurisée des secrets ;
* logs sans données sensibles ;
* headers de sécurité lorsque pertinents ;
* CORS correctement configuré ;
* gestion des erreurs ne révélant pas d'informations internes.

**Le frontend ne doit jamais être considéré comme une frontière de sécurité.**

Toute autorisation critique doit être vérifiée côté backend.

Les secrets, credentials, tokens privés et données sensibles ne doivent jamais être commités dans le repository.

---

# 10. Architecture et qualité du code

Le code doit être :

* modulaire ;
* maintenable ;
* testable ;
* fortement typé lorsque le langage le permet ;
* cohérent ;
* documenté lorsque nécessaire ;
* facilement extensible.

Évite :

* la duplication ;
* les abstractions prématurées ;
* les composants gigantesques ;
* les fonctions excessivement complexes ;
* les dépendances inutiles ;
* les solutions spécifiques lorsqu'une abstraction réutilisable est appropriée.

Avant d'ajouter une nouvelle abstraction, vérifie si une abstraction existante peut être réutilisée.

---

# 11. Design Stitch → implémentation

Les designs présents dans `design/` doivent être traités comme la référence de l'expérience utilisateur.

Avant d'implémenter une page :

1. inspecte sa maquette ;
2. identifie sa structure ;
3. identifie les composants réutilisables ;
4. identifie les états possibles ;
5. identifie les interactions ;
6. identifie les formulaires ;
7. identifie les tableaux/listes ;
8. identifie les états loading ;
9. identifie les états empty ;
10. identifie les états error ;
11. identifie les comportements responsive ;
12. identifie les besoins dark mode ;
13. identifie les besoins d'internationalisation.

L'objectif n'est pas de produire uniquement une copie visuelle du screenshot.

L'objectif est de transformer le design en **système d'interface réellement fonctionnel, accessible, responsive et maintenable**.

---

# 12. Documentation

Le dossier `docs/` doit être considéré comme une partie intégrante du projet.

Complète tous les documents existants.

La documentation doit notamment permettre de comprendre :

* le produit ;
* l'architecture ;
* le frontend ;
* le backend ;
* la base de données ;
* les APIs ;
* l'authentification ;
* les conventions ;
* les composants UI ;
* l'internationalisation ;
* la gestion des erreurs ;
* la sécurité ;
* le déploiement ;
* les décisions techniques importantes.

Lorsque cela est pertinent, documente également les raisons derrière les choix techniques et pas uniquement leur fonctionnement.

---

# 13. Méthode de travail obligatoire pour cette première tâche

Tu ne dois pas commencer directement par coder.

Commence par une phase d'analyse.

### Étape 1 — Inspection

Inspecte :

* l'arborescence du projet ;
* les skills disponibles ;
* `prompts/master.md` ;
* `design/` ;
* `docs/` ;
* les fichiers de configuration ;
* le code existant ;
* les dépendances ;
* les éventuelles migrations ;
* les tests existants.

### Étape 2 — Compréhension

Construis une compréhension globale :

* du produit ;
* des fonctionnalités ;
* des utilisateurs ;
* de l'architecture ;
* du design ;
* des contraintes ;
* des conventions ;
* des dépendances entre fonctionnalités.

### Étape 3 — Planification

Définis un plan d'implémentation cohérent.

Le plan doit identifier :

* les fondations techniques ;
* le backend ;
* le modèle de données ;
* les APIs ;
* l'authentification ;
* le frontend ;
* le design system ;
* les composants réutilisables ;
* les pages ;
* les formulaires ;
* les tests ;
* la documentation.

### Étape 4 — Fondations

Mets d'abord en place les fondations nécessaires :

* architecture ;
* configuration ;
* base de données ;
* conventions ;
* système d'erreurs ;
* i18n ;
* thème ;
* design system ;
* composants réutilisables ;
* patterns de formulaires ;
* patterns de listes ;
* authentification/sécurité si nécessaire.

### Étape 5 — Implémentation

Implémente ensuite les fonctionnalités en suivant le plan.

Ne développe pas chaque écran comme une fonctionnalité isolée.

Construis d'abord les primitives et patterns réutilisables.

### Étape 6 — Validation

Après chaque fonctionnalité importante :

* vérifie le typage ;
* lance les tests ;
* vérifie le lint ;
* vérifie le build ;
* vérifie les erreurs ;
* vérifie les états loading/empty/error ;
* vérifie le responsive ;
* vérifie le dark mode ;
* vérifie les traductions FR/EN ;
* vérifie l'accessibilité ;
* vérifie la sécurité.

### Étape 7 — Documentation et mémoire

À la fin de chaque étape significative :

1. mets à jour `docs/` si nécessaire ;
2. mets à jour `AGENT.md` ;
3. indique clairement ce qui est terminé ;
4. indique ce qui reste à faire ;
5. indique les éventuels problèmes ou décisions importantes.

---

# 14. Gestion des ambiguïtés

Lorsqu'une information n'est pas explicitement définie :

1. cherche d'abord dans les skills ;
2. cherche dans `prompts/master.md` ;
3. cherche dans `docs/` ;
4. cherche dans `design/` ;
5. regarde les conventions déjà utilisées dans le code ;
6. privilégie la cohérence avec l'architecture existante.

Ne crée pas arbitrairement une nouvelle convention lorsqu'une convention existe déjà.

Si deux sources sont contradictoires, identifie explicitement le conflit et applique la source ayant la priorité la plus élevée selon le contexte du projet.

---

# 15. Priorité des règles

En cas de conflit, applique cet ordre de priorité :

1. contraintes explicites de ce prompt ;
2. `prompts/master.md` ;
3. architecture et conventions établies du projet ;
4. skills disponibles ;
5. design Stitch ;
6. documentation existante ;
7. bonnes pratiques générales.

Si une contradiction importante ne peut pas être résolue raisonnablement, signale-la avant de prendre une décision irréversible.

---

# 16. Definition of Done

Une fonctionnalité n'est pas considérée comme terminée simplement parce que le code compile.

Elle est terminée lorsqu'elle est :

* fonctionnelle ;
* testée ;
* correctement typée ;
* sécurisée ;
* responsive ;
* compatible light/dark mode ;
* traduite FR/EN ;
* cohérente avec le design Stitch ;
* cohérente avec le design system ;
* correctement intégrée au système d'erreurs ;
* correctement intégrée aux patterns de formulaires/listes ;
* documentée lorsque nécessaire ;
* et que `AGENT.md` reflète son état réel.

---

# 17. Première mission

Pour cette première tâche, ton objectif est **d'initialiser et de préparer correctement le projet avant de lancer le développement fonctionnel massif**.

Commence donc par :

1. inspecter entièrement le repository ;
2. identifier et lire les skills disponibles ;
3. lire `prompts/master.md` ;
4. analyser `design/` ;
5. analyser `docs/` ;
6. analyser l'architecture existante ;
7. identifier les éventuelles contradictions ou informations manquantes ;
8. créer `AGENT.md` ;
9. compléter les fichiers de `docs/` ;
10. définir l'architecture cible ;
11. définir les conventions de développement ;
12. définir les patterns UI réutilisables ;
13. définir le pattern unique des formulaires ;
14. définir le pattern de listes/pagination/filtres ;
15. définir le système d'erreurs ;
16. définir l'approche i18n ;
17. définir l'approche dark mode ;
18. définir les règles de sécurité ;
19. établir le plan d'implémentation ;
20. puis seulement commencer l'implémentation des fondations nécessaires.

Ne considère pas cette première tâche comme terminée tant que le projet n'est pas suffisamment documenté et structuré pour qu'un autre agent puisse reprendre le développement sans devoir refaire toute l'analyse.

**Règle finale : lis `AGENT.md` avant chaque nouvelle tâche et maintiens-le systématiquement à jour pendant toute la durée du projet.**
