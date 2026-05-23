---
name: story-fr
description: Découpage de spécifications en User Stories (Français)
---

# Skill: Stories (FR)

Ce guide décrit comment générer des user stories à partir d'une spécification en considérant la documentation et l'automatisation des tests.

## Qu'est-ce que je fais

Je découpe une spécification en user stories exploitables, en tenant compte de la documentation du logiciel et de l'automatisation des tests. Chaque story doit apporter une valeur probante à l'utilisateur final.

## Quand m'utiliser

Utilisez cette skill lorsque vous avez une spécification de projet et que vous voulez générer des user stories structurées, testables et documentées.

## Workflow

### Cas 1 : Aucune story n'existe (début depuis zéro)

**Étape 1 - Lecture de la spécification :**
- L'utilisateur donne le chemin de la spécification en argument (ex: specification.md ou docs/specification/specification.md)
- Lire la spécification fournie
- Générer des user stories qui répondent aux critères :
  - Considèrent la documentation du logiciel existante
  - Intègrent l'automatisation des tests
  - Apportent une valeur probante à l'utilisateur final
- Créer un répertoire de travail (ex: docs/stories/) si inexistant
- Générer les fichiers de stories (ex: docs/stories/story-001.md, story-002.md, etc.)

**Étape 2 - Création du fichier de questions :**
- Créer un fichier temporaire docs/stories/questions.md
- Ce fichier contiendra les questions auxquelles l'utilisateur devra répondre

**Étape 3 - Analyse et clarification :**
- Analyser les stories générées
- Vérifier la cohérence vis-à-vis de la spécification
- Identifier ce qui n'est pas clair ou ce qu'il manque
- Ajouter des questions de clarification dans questions.md avec :
  - Le contexte de la question
  - Des propositions éventuelles pour que l'utilisateur puisse répondre directement
  - Des références à la documentation si applicable

**Étape 4 - Mise à jour :**
- L'utilisateur répond aux questions dans questions.md
- L'utilisateur tape : J'ai répondu aux questions, mets à jour les stories
- Mettre à jour les stories en fonction des réponses données
- Retourner à l'Étape 3

**Étape 5 - Génération du plan d'implémentation et fin du workflow :**
- Quand il n'y a plus rien à clarifier (plus de questions dans questions.md)
- Supprimer le fichier questions.md
- Analyser les dépendances entre les stories (section ## Dépendances de chaque story)
- Générer le fichier docs/stories/plan-implementation.md contenant :
  - Ordre d'implémentation recommandé (basé sur les dépendances, la valeur utilisateur, la complexité)
  - Stories pouvant être développées en parallèle (sans dépendances croisées)
  - Justifications de l'ordre proposé
  - Diagramme de dépendances au format Mermaid
- Annoncer la fin du workflow

### Cas 2 : Des stories existent déjà

- Démarrer directement à l'Étape 3 du Cas 1
- Lire les stories existantes dans le répertoire de travail
- Analyser et itérer jusqu'à résolution complète

## Format des fichiers

### Story (story-XXX.md)
``markdown
# Story #XXX : [Titre de la story]

## Description
En tant que [rôle], je veux [action] afin de [bénéfice].

## Critères d'acceptation
- [ ] Critère 1
- [ ] Critère 2
- [ ] Critère 3

## Tests automatisés
- Test unitaire : [Description]
- Test d'intégration : [Description]
- Test E2E : [Description]

## Documentation
- [Lien vers la documentation impactée]
- [Nouvelle documentation à créer]

## Valeur utilisateur
[Description de la valeur probante pour l'utilisateur final]

## Dépendances
- [Story #YYY]
- [Documentation existante]
``

### questions.md
``markdown
# Questions de clarification - Stories

## Contexte
[Spécification utilisée : spec.md]
[Stories générées : 5 stories]

## Questions
1. **[Story #002 - Critère d'acceptation]** 
   - Contexte : La story #002 mentionne une authentification mais les critères ne précisent pas la gestion des rôles.
   - Question : Quels rôles doivent être gérés (Admin, User, Guest) ?
   - Propositions : 
     - a) Admin (tout accès), User (accès limité), Guest (lecture seule)
     - b) Admin et User uniquement
     - c) Autre : [préciser]

2. **[Story #004 - Test automatisé]**
   - Contexte : La story #004 nécessite des tests E2E pour le drag-and-drop des tâches.
   - Question : Utilisez-vous déjà un framework E2E (Cypress, Playwright, Selenium) ?
   - Propositions :
     - a) Cypress (déjà configuré dans le projet)
     - b) Playwright (recommandé pour les apps modernes)
     - c) Autre : [préciser]

## Inconsistances détectées
- [Description de l'inconsistance entre stories ou avec la spécification]
``

### plan-implementation.md
``markdown
# Plan d'implémentation des stories

## Ordre d'implémentation recommandé
1. **Story #001** - [Titre] (Priorité : Haute)
   - Justification : [Dépendance nulle, valeur utilisateur élevée]
2. **Story #002** - [Titre] (Priorité : Haute)
   - Justification : [Dépend de Story #001, nécessaire pour les fonctionnalités suivantes]
3. **Story #003** - [Titre] (Priorité : Moyenne)
   - Justification : [Peut être développée en parallèle de #004]

## Stories développables en parallèle
- **Groupe 1** : Story #003, Story #004 (aucune dépendance croisée)
- **Groupe 2** : Story #005, Story #006 (dépendent uniquement des stories déjà implémentées)

## Diagramme de dépendances (format Mermaid)
``mermaid
flowchart TD
    %% Stories ordonnées
    S001[Story #001 : Titre] --> S002[Story #002 : Titre]
    S002 --> S003[Story #003 : Titre]
    S002 --> S004[Story #004 : Titre]
    
    %% Groupes parallèles
    subgraph Parallel1
        S003
        S004
    end
    
    S003 --> S005[Story #005 : Titre]
    S004 --> S005
    S005 --> S006[Story #006 : Titre]
    
    subgraph Parallel2
        S005
        S006
    end
``

## Notes
- Les stories doivent être implémentées dans l'ordre indiqué pour respecter les dépendances
- Les groupes en parallèle peuvent être développés simultanément par différentes équipes
``

## Comment passer d'une étape à l'autre

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 → 2 | Donner le chemin de la spécification en argument | Génère les stories initiales et crée questions.md |
| Étape 2 → 3 | Automatique après création de questions.md | Analyse les stories et ajoute les questions |
| Étape 3 → 4 | Taper J'ai répondu aux questions, mets à jour les stories | Met à jour les stories selon les réponses |
| Étape 4 → 3 | Automatique après mise à jour | Re-analyse les stories mises à jour |
| Étape 4 → 5 | Automatique (plus de questions) | Supprime questions.md, génère plan-implementation.md, annonce la fin |

**Important :** L'agent ne surveille pas les modifications de fichiers en temps réel. L'utilisateur doit toujours donner une instruction explicite pour que l'agent passe à l'étape suivante.

## Exemple d'utilisation

**Cas 1 - Début depuis zéro :**
``
/ stories specification.md
``

**Cas 2 - Stories existantes :**
``
/ stories
``
(L'agent détecte automatiquement les stories existantes dans docs/stories/ et commence l'analyse)
