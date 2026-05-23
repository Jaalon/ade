---
name: specification-fr
description: Création itérative de spécifications de projet (Français)
---

# Skill: Spécification (FR)

Ce guide décrit comment générer une spécification de projet itérative avec questions et réponses jusqu'à résolution complète.

## Qu'est-ce que je fais

Je guide la création d'une spécification de projet complète à travers un processus itératif de questions, réponses et clarification jusqu'à ce que toutes les inconsistances soient résolues.

## Quand m'utiliser

Utilisez cette skill lorsque vous voulez créer une spécification de projet structurée à partir d'une idée sommaire, en itérant jusqu'à ce que tout soit clair et cohérent.

## Étapes du processus

### Étape 1 : Initiation
- Créer un répertoire de travail documentaire (ex: `docs/specification/`) s’il n’existe pas déjà
- Générer le fichier `questions.md` s’il n’existe pas. 
- Si un fichier de spécifications existe déjà, utiliser son contenu comme base de spécification.  
- Sinon, générer ce fichier specification.md vide et demander à l'utilisateur de décrire sommairement ce que fera le projet dans questions.md.
- Poser dans questions.md des questions ouvertes pour mieux comprendre le projet

### Étape 2 : Création de spécification
- Lire les réponses de l'utilisateur dans `questions.md`
- Mettre à jour `specification.md` avec :
  - La description du projet structurée
  - Des placeholders clairs (ex: `[À_CLARIFIER: ...]`) pour identifier les points flous
  - Des marqueurs d'inconsistances (ex: `[INCONSISTANCE: ...]`)
- Générer un nouveau fichier `questions.md` contenant :
  - Les questions nécessaires pour résoudre les placeholders
  - Le contexte pertinent pour chaque question
  - Les inconsistances détectées à résoudre

### Étape 3 : Mise à jour
- Lire les réponses dans `questions.md`
- Mettre à jour `specification.md` en remplaçant les placeholders par les réponses
- Résoudre les inconsistances identifiées

### Étape 4 : Itération
- Relancer l'Étape 2 avec un nouvel agent
- Vérifier s'il reste des placeholders ou inconsistances
- Si oui, continuer l'itération
- Si non, supprimer le fichier questions.md et considérer la spécification comme complète

## Format des fichiers

### questions.md
```markdown
# Questions de spécification

## Contexte
[Brief résumé du projet actuel]

## Questions
1. [Question claire avec contexte]
2. [Question claire avec contexte]
...

## Inconsistances à résoudre
- [Description de l'inconsistance]
...
```

### specification.md
```markdown
# Spécification - [Nom du projet]

## Vue d'ensemble
[Description générale]

## Fonctionnalités
[Liste des fonctionnalités]

## [Section avec placeholder]
[À_CLARIFIER: Précisions nécessaires sur...]

## [Section avec inconsistances]
[INCONSISTANCE: Conflit identifié entre...]
```

## Exemple d'utilisation

**Lancement de la skill :**
```
/ specification
```

**Étape 1 - L'agent crée `docs/specification/questions.md` :**
```markdown
# Questions de spécification

## Contexte
Expliquer ce que le projet devra faire

## Questions
1. Décrivez sommairement ce que fera votre projet (objectif, fonctionnalité principale)
2. Quel est le public cible ?
3. Y a-t-il des contraintes techniques particulières ?
```

**L'utilisateur répond dans `questions.md` :**
```markdown
1. Une application web pour gérer des tâches collaboratives avec tableaux Kanban
2. Équipes de développement agile
3. Doit être compatible mobile, pas de contrainte de tech stack
```

**Étape 2 - L'agent génère `specification.md` :**
```markdown
# Spécification - Application Kanban Collaborative

## Vue d'ensemble
Application web permettant aux équipes de développement agile de gérer leurs tâches via des tableaux Kanban collaboratifs.

## Fonctionnalités
- Création de tableaux et colonnes personnalisables
- Gestion des tâches avec drag-and-drop
- [À_CLARIFIER: Système d'authentification (SSO, email, OAuth ?)]
- [À_CLARIFIER: Gestion des rôles et permissions]

## Contraintes techniques
- Compatible mobile (responsive design)
- [INCONSISTANCE: Pas de tech stack imposée mais besoin de temps réel pour la collaboration]
```

**Et génère un nouveau `questions.md` :**
```markdown
# Questions de spécification

## Contexte
Application Kanban pour équipes agiles, compatible mobile, besoin de temps réel pour la collaboration

## Questions
1. [À_CLARIFIER: Système d'authentification] Quel système d'authentification souhaitez-vous ? (SSO, email/password, OAuth avec Google/GitHub ?)
2. [À_CLARIFIER: Rôles] Quels rôles prévoyez-vous ? (Admin, membre, invité ?)
3. [INCONSISTANCE: Temps réel vs Tech stack] Vous n'avez pas de préférence de tech stack mais avez besoin de temps réel. Préférez-vous une solution WebSocket, Server-Sent Events, ou une approche avec une BaaS comme Supabase/Firebase ?
```

**Transition Étape 1 → Étape 2 :**
L'utilisateur doit informer l'agent que les questions sont répondues en tapant simplement :
```
J'ai répondu aux questions, continue
```
L'agent ne détecte pas automatiquement les modifications de fichier. L'utilisateur doit explicitement donner l'instruction de continuer.

**Étape 3 - L'utilisateur répond dans `questions.md`, puis tape :**
```
Questions répondues, mets à jour la spécification
```
L'agent met à jour `specification.md` en remplaçant les placeholders par les réponses.

**Transition Étape 3 → Étape 4 :**
L'agent vérifie automatiquement s'il reste des placeholders. S'il en reste, il relance l'Étape 2 avec un nouveau cycle de questions.

**Étape 4 - Itération :**
L'agent relance l'Étape 2, vérifie qu'il n'y a plus de placeholders, la spécification est complète !

## Comment passer d'une étape à l'autre

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 → 2 | Taper `continue` ou `j'ai répondu` | Lit `questions.md` et génère `specification.md` + nouveau `questions.md` |
| Étape 2 → 3 | Taper `continue` ou `questions répondues` | Lit les réponses et met à jour `specification.md` |
| Étape 3 → 4 | Automatique | Vérifie s'il reste des placeholders, relance l'Étape 2 si nécessaire |
| Fin | Automatique | Annonce que la spécification est complète |

**Important :** L'agent ne surveille pas les modifications de fichiers en temps réel. L'utilisateur doit toujours donner une instruction explicite pour que l'agent passe à l'étape suivante, sauf pour la transition automatique Étape 3 → 4.

## Critère de fin

La spécification est terminée quand `specification.md` ne contient plus aucun placeholder `[À_CLARIFIER: ...]` ou `[INCONSISTANCE: ...]`.
