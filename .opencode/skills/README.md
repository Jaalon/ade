# Agent Skills (Agnostiques)

Ce répertoire contient des guides de workflow (skills) pour aider les agents de code (Claude Code, GitHub Copilot, Junie, etc.) à réaliser des tâches complexes de manière structurée et itérative.

## Pourquoi ces skills ?

Ces fichiers sont conçus pour être lus par n'importe quel agent de code. Ils décrivent des workflows précis, des structures de fichiers attendues et des mécanismes de clarification avec l'utilisateur (via des fichiers `questions.md`).

## Comment utiliser une skill

Pour activer un workflow avec un agent, il suffit de lui demander de lire le fichier correspondant et de suivre ses instructions.

**Exemple :**
> "Lis `ia/skills/specification-fr/SKILL.md` et guide-moi pour créer la spécification du projet."

## Liste des skills disponibles

- **[specification-fr](./specification-fr/SKILL.md)** : Création itérative de spécifications (Français).
- **[specification-en](./specification-en/SKILL.md)** : Iterative creation of specifications (English).
- **[story-fr](./story-fr/SKILL.md)** : Découpage de spécification en User Stories (Français).
- **[story-en](./story-en/SKILL.md)** : Breaking down specifications into User Stories (English).
- **[tasks](tasks-fr/SKILL.md)** : Découpage de Story en tâches de code exploitables (Prompts).
- **[feedback-fr](./feedback-fr/SKILL.md)** : Génération de stories à partir de feedbacks utilisateur (Français).

## Conventions partagées

La plupart de ces skills utilisent un mécanisme de questions/réponses via un fichier `questions.md` dans le répertoire de travail concerné (`docs/specification/`, `docs/stories/`, etc.). Cela permet :
1. De ne pas polluer le chat avec de trop nombreuses questions.
2. De garder une trace écrite des décisions prises.
3. De permettre à l'agent de reprendre le travail là où il s'était arrêté ou à un autre agent de prendre le relais.
