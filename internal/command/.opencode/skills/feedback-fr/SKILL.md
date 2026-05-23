---
name: feedback-fr
description: Génération de stories à partir de feedbacks utilisateurs, en filtrant ceux déjà traités
---

# Skill: Feedback → Stories (FR)

Ce guide décrit comment transformer des feedbacks utilisateurs en user stories exploitables, en ne traitant que les retours les plus récents qui n'ont pas encore été pris en compte dans les stories existantes.

## Qu'est-ce que je fais

Je prends le fichier `docs/specification/feedbacks.md`, j'identifie les feedbacks qui n'ont pas encore été traités, et je génère une ou plusieurs stories (au même format que story-fr) avec le processus itératif de questions/réponses.

## Quand m'utiliser

Utilisez cette skill lorsque :
- Vous avez collecté des retours utilisateurs dans `docs/specification/feedbacks.md`
- Vous voulez transformer ces retours en stories structurées et testables
- Vous voulez éviter de retraiter des feedbacks déjà pris en compte

## Prérequis

Le fichier `docs/specification/feedbacks.md` doit être structuré comme suit :

```markdown
## 2025-05-14

* Feedback 1
* Feedback 2
...

## [DATE]

* Feedback N
...
```

Chaque feedback traité reçoit un marqueur `[PROCESSED: story-XXX]` :

```markdown
* Feedback 1 [PROCESSED: story-018]
```

## Workflow

### Étape 1 - Lecture du fichier de feedbacks

- Lire `docs/specification/feedbacks.md`
- Repérer les feedbacks NON marqués `[PROCESSED: story-XXX]` (ce sont ceux à traiter)
- Si TOUS les feedbacks sont marqués : annoncer qu'il n'y a rien de nouveau à traiter → fin du workflow

### Étape 2 - Analyse des stories existantes

- Lire les stories dans `docs/stories/done/` ou `docs/stories/` 
- Vérifier si certains feedbacks non marqués sont déjà couverts par des stories existantes
  - Exemple : "Avoir un écran avec la liste des zones" peut déjà exister dans Story #009
  - Si c'est le cas, ajouter le marqueur `[PROCESSED: story-XXX]` (donc déjà couvert) et ne pas recréer de story
- Les feedbacks restants (ni marqués PROCESSED, ni couverts) sont à transformer en nouvelles stories

### Étape 3 - Regroupement des feedbacks en stories

- Regrouper les feedbacks par thème commun (ex : "éditeur", "révision", "UI/UX")
- Un feedback seul peut devenir une story s'il est suffisamment riche
- Plusieurs feedbacks liés peuvent être regroupés dans une même story
- Un feedback complexe peut être découpé en plusieurs stories

### Étape 4 - Génération des stories

- Lire les stories dans `docs/stories/done/` pour déterminer le prochain numéro de story
- Créer les nouvelles stories dans `docs/stories/` en suivant le format story-fr :

```markdown
# Story #XXX : [Titre]

## Description
En tant que [rôle], je veux [action] afin de [bénéfice].

## Origine
Feedback du [DATE] : "[texte du feedback]"

## Critères d'acceptation
- [ ] Critère 1
- [ ] Critère 2

## Tests automatisés
- Test unitaire : [Description]
- Test d'intégration : [Description]
- Test E2E : [Description]

## Documentation
- [Nouvelle documentation à créer]

## Valeur utilisateur
[Description de la valeur probante pour l'utilisateur final]

## Dépendances
- [Story #YYY si applicable]
```

- Ajouter une section `## Origine` pour tracer le feedback source
- Marquer chaque feedback traité dans `docs/specification/feedbacks.md` avec `[PROCESSED: story-XXX]`
- Exemple : `* Ne pas afficher d'accueil [PROCESSED: story-018]`

### Étape 5 - Création du fichier de questions

- Créer `docs/stories/questions.md` (ou s'il existe déjà, le compléter)
- Contient les questions de clarification pour les nouvelles stories générées
- Même format que story-fr (contexte, question, propositions)

### Étape 6 - Analyse et clarification

- Analyser les nouvelles stories
- Vérifier la cohérence avec la spécification `docs/specification/specification.md`
- Vérifier la cohérence avec les stories existantes dans `docs/stories/done/`
- Ajouter des questions dans `questions.md` si nécessaire

### Étape 7 - Mise à jour

- L'utilisateur répond aux questions dans `questions.md`
- L'utilisateur tape : `J'ai répondu aux questions, mets à jour les stories`
- Mettre à jour les stories selon les réponses
- Retourner à l'Étape 6

### Étape 8 - Fin du workflow

- Quand il n'y a plus rien à clarifier
- Supprimer le fichier `questions.md`
- Déplacer les nouvelles stories dans `docs/stories/done/` (ou les laisser dans `docs/stories/` selon la convention)
- Mettre à jour `AGENTS.md` et `workflow.md` si nécessaire
- Annoncer la fin du workflow

## Format du fichier feedbacks.md

```markdown
## [DATE]

* [Texte du feedback] [PROCESSED: story-XXX]
* [Texte du feedback]

## [AUTRE DATE]

* [Texte du feedback] [PROCESSED: story-XXX]
```

Les feedbacks marqués `[PROCESSED: story-XXX]` sont ignorés.
Les feedbacks sans marqueur sont à traiter.

## Exemple

### Fichier feedbacks.md initial
```markdown
## 2025-05-14

* Ne pas afficher d'accueil : on tombe direct sur "Mes équipes"
* Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock [PROCESSED: story-010]
* Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés
```

### Stories générées
- `docs/stories/story-018.md` : "Suppression de la page d'accueil" (basé sur feedback #1)
- `docs/stories/story-019.md` : "Popup d'aide dans l'éditeur" (basé sur feedback #3)

### Fichier feedbacks.md après traitement
```markdown
## 2025-05-14

* Ne pas afficher d'accueil : on tombe direct sur "Mes équipes" [PROCESSED: story-018]
* Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock [PROCESSED: story-010]
* Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés [PROCESSED: story-019]
```

## Comment passer d'une étape à l'autre

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 → 2 | Donner le chemin du fichier de feedbacks en argument | Lit les feedbacks, analyse les stories existantes |
| Étape 2 → 3 | Automatique | Regroupe les feedbacks non traités par thème |
| Étape 3 → 4 | Automatique | Génère les nouvelles stories dans docs/stories/ |
| Étape 4 → 5 | Automatique après génération | Crée questions.md avec les points à clarifier |
| Étape 5 → 6 | Automatique | Analyse les stories et ajoute les questions |
| Étape 6 → 7 | Taper "J'ai répondu aux questions, mets à jour les stories" | Met à jour les stories selon les réponses |
| Étape 7 → 6 | Automatique après mise à jour | Re-vérifie la cohérence |
| Étape 7 → 8 | Automatique (plus de questions) | Supprime questions.md, marque les feedbacks traités, annonce la fin |

**Important :** L'agent ne surveille pas les modifications de fichiers en temps réel. L'utilisateur doit toujours donner une instruction explicite pour que l'agent passe à l'étape suivante.

## Exemple d'utilisation

```
/ feedback feedbacks.md
```

L'agent lit `docs/specification/feedbacks.md`, identifie les feedbacks non marqués, vérifie les stories existantes, et génère les nouvelles stories avec le processus itératif de clarification.

### Gestion des doublons avec les stories existantes

Si un feedback non marqué est déjà couvert par une story existante :

1. L'agent l'identifie lors de l'Étape 2
2. Il ajoute le marqueur `[PROCESSED: story-XXX]` dans le fichier feedbacks.md
3. Il n'en génère pas de nouvelle story
4. Il peut notifier l'utilisateur du doublon dans questions.md
