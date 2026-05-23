---
name: tasks-fr
description: Découpage de User Stories en tâches de code exploitables (Prompts)
---

# Skill: Tasks

Ce guide décrit comment découper une story en tâches "prompt" exécutables par un agent de code avec tests, documentation et exemples d'API.

## Qu'est-ce que je fais

Je prends une story existante, je la découpe en tâches élémentaires. Chaque tâche contient un prompt le plus détaillé possible, prêt à être exécuté par un agent de code. Je vérifie la cohérence des tâches générées et je clarifie avec l'utilisateur via un fichier `questions.md` jusqu'à résolution complète.

## Quand m'utiliser

Utilisez cette skill après avoir créé vos user stories avec la skill `stories`, quand vous voulez passer à l'implémentation.

## Workflow

### Étape 1 - Lecture de la story
- Prendre le numéro de story fourni en argument (ex: `/ tasks-fr 001`)
- Lire le fichier `docs/stories/story-XXX.md`
- Extraire la description, les critères d'acceptation, les tests et la documentation attendue
- Lire la spécification et les autres stories pour comprendre le contexte global

### Étape 2 - Découpage et génération des tâches avec prompts détaillés
- Analyser la story pour identifier les unités de travail atomiques
- Pour chaque unité, générer un prompt extrêmement détaillé incluant :
  - **Objectif précis** de la tâche
  - **Description fonctionnelle** : comportement attendu, cas nominaux et cas limites
  - **Fichiers concernés** : chemins exacts, nature de la modification (création, modification, suppression)
  - **Contraintes techniques** : framework, patterns, conventions du projet, gestion d'erreurs, performance, sécurité
  - **Signature des fonctions/classes** : noms, paramètres, types, valeurs de retour
  - **Structure des données** : schémas, modèles, DTO, interfaces
  - **Tests à implémenter** : scénarios unitaires, intégration, E2E avec cas de test concrets
  - **Documentation** : quoi documenter et où
  - **Exemples d'utilisation** : extraits de code, requêtes API HTTP
- Créer un répertoire `docs/tasks/story-XXX/`
- Générer les fichiers `docs/tasks/story-XXX/task-YYY.md`

### Étape 3 - Vérification de cohérence
- Analyser les tâches générées entre elles :
  - Vérifier qu'il n'y a pas de dépendances circulaires ou d'ordre d'exécution problématique
  - Vérifier que chaque tâche a des entrées et sorties claires
  - Vérifier que la somme des tâches couvre bien 100% des critères d'acceptation de la story
  - Identifier les doublons ou chevauchements entre tâches
  - Vérifier la cohérence des nommages, types, signatures entre tâches
- Analyser la cohérence avec le projet existant :
  - Vérifier la compatibilité avec les conventions et l'architecture existante
  - Vérifier que les fichiers référencés existent ou sont bien à créer

### Étape 4 - Création de questions.md
- Créer un fichier `docs/tasks/story-XXX/questions.md`
- Ajouter des questions de clarification pour :
  - Les choix d'implémentation non tranchés
  - Les incohérences entre tâches
  - Les ambiguïtés de la story
  - Les dépendances et l'ordre d'exécution
  - Les propositions techniques alternatives
- Chaque question doit inclure :
  - Le contexte (tâche concernée, section de la story)
  - Des propositions de réponse pour guider l'utilisateur
  - Les implications de chaque choix possible

### Étape 5 - Mise à jour
- L'utilisateur répond aux questions dans `questions.md`
- L'utilisateur tape : `J'ai répondu, mets à jour les tâches`
- Mettre à jour les tâches en fonction des réponses
- Retourner à l'Étape 3

### Étape 6 - Fin du workflow
- Quand il n'y a plus rien à clarifier
- Supprimer le fichier `questions.md`
- Présenter le plan final à l'utilisateur

## Format des prompts détaillés

### Structure du prompt (task-YYY.md)

```markdown
# Tâche #YYY - Story #XXX : [Titre de la tâche]

## Objectif
[Description concise de ce que cette tâche doit livrer]

## Contexte
- Story #XXX : [Lien vers la story]
- Dépend de : Tâche #ZZZ (si applicable)
- Nécessaire pour : Tâche #WWW (si applicable)

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle
[Description détaillée du comportement attendu]

**Cas nominaux :**
- [Cas 1]
- [Cas 2]

**Cas limites :**
- [Cas limite 1]
- [Cas limite 2]

**Gestion d'erreurs :**
- [Erreur 1] → [Comportement attendu]
- [Erreur 2] → [Comportement attendu]

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `src/feature/exemple.ts` | Créer | Nouveau service |
| `src/feature/types.ts` | Modifier | Ajouter le type X |
| `src/feature/__tests__/exemple.test.ts` | Créer | Tests unitaires |

### Signatures

```typescript
// Fonction à implémenter
function exemple(
  param1: string,
  param2: number
): Promise<Resultat>
```

### Contraintes techniques
- **Framework** : Utiliser [framework] (version X)
- **Pattern** : Suivre le pattern [XXX] déjà utilisé dans [fichier de référence]
- **Style** : Respecter la config ESLint et Prettier du projet
- **Performance** : [contrainte si applicable]
- **Sécurité** : [contrainte si applicable]
- **Tests** : Coverage minimum de 80% pour cette tâche

### Tests à implémenter

#### Tests unitaires
- **Fichier** : `src/feature/__tests__/exemple.test.ts`
- Scénario 1 : [Description]
  - Données : [Input]
  - Résultat attendu : [Output]
- Scénario 2 : [Description]
  - Données : [Input]
  - Résultat attendu : [Output]

#### Tests d'intégration
- **Fichier** : `src/feature/__tests__/integration.test.ts`
- Scénario : [Description]

### Documentation

#### Documentation à créer
- `docs/feature/README.md` : Documentation de la nouvelle fonctionnalité
- `docs/feature/exemples.md` : Exemples d'utilisation

#### Documentation à mettre à jour
- `docs/api/README.md` : Ajouter la section sur le nouvel endpoint

### Exemples d'API (si applicable)
- Fichier : `docs/tasks/story-XXX/task-YYY-examples.http`

Voir les exemples de requêtes dans le fichier .http associé.
```

### Format du fichier questions.md

```markdown
# Questions de clarification - Story #XXX

## Contexte
Story #XXX : [Titre de la story]
Tâches générées : [liste des tâches]

## Questions

### 1. [Tâche #YYY] - [Sujet de la question]
**Contexte :** [Description du contexte précis]
**Question :** [Question à choix ou ouverte]
**Propositions :**
- a) [Proposition A] → [Implication]
- b) [Proposition B] → [Implication]
- c) Autre : [à préciser]

### 2. [Incohérence entre tâches #YYY et #ZZZ] - [Sujet]
**Contexte :** La tâche #YYY définit X alors que la tâche #ZZZ attend Y.
**Question :** Quelle est la bonne définition ?
**Propositions :**
- a) Aligner sur la tâche #YYY : [conséquence]
- b) Aligner sur la tâche #ZZZ : [conséquence]
- c) Définir un nouveau standard : [proposition]

### 3. [Ambiguïté de la story] - [Sujet]
**Contexte :** La story mentionne [point ambigu] sans précision.
**Question :** [Question de clarification]
**Propositions :**
- a) [Proposition A]
- b) [Proposition B]
```

## Comment passer d'une étape à l'autre

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 | Donner le numéro de story (ex: `/ tasks 001`) | Lit la story, le contexte projet |
| Étape 1 → 2 | Automatique | Découpe et génère les prompts détaillés |
| Étape 2 → 3 | Automatique | Vérifie la cohérence des tâches |
| Étape 3 → 4 | Automatique | Crée `questions.md` avec les points à clarifier |
| Étape 4 → 5 | Répondre dans `questions.md` puis taper `J'ai répondu, mets à jour les tâches` | Met à jour les tâches selon les réponses |
| Étape 5 → 3 | Automatique | Re-vérifie la cohérence (itération) |
| Étape 5 → 6 | Automatique (plus de questions) | Supprime `questions.md`, présente le plan final |

**Important :** L'agent ne surveille pas les modifications de fichiers en temps réel. L'utilisateur doit donner une instruction explicite pour passer à l'étape suivante.

## Exemple d'utilisation

**Lancement :**
```
/ tasks 001
```

**Lecture de `docs/stories/story-001.md` :**
Story décrivant la création d'une API REST de gestion de tâches avec création, lecture, mise à jour et suppression.

**Découpage en 3 tâches avec prompts détaillés :**
- `docs/tasks/story-001/task-001.md` : Modèle de données, base et validations
- `docs/tasks/story-001/task-002.md` : Endpoints CRUD complets avec signatures et gestion d'erreurs
- `docs/tasks/story-001/task-003.md` : Tests automatisés, documentation et fichiers `.http`

**Vérification de cohérence :**
- Tâche #002 référence un type défini dans tâche #001 → OK
- Tâche #003 couvre tous les cas de test des critères d'acceptation → OK
- Gestion d'erreurs : tâche #002 et spécification alignées → OK

**Création de `docs/tasks/story-001/questions.md` :**
- Question sur le format de retour des erreurs (JSON API vs standard)
- Question sur l'ordre d'exécution des tâches (tâche #001 doit précéder #002)

**Mise à jour :**
L'utilisateur répond, l'agent met à jour les tâches, itération jusqu'à résolution complète.
