# Spécification - Automated Dev Environment

## Vue d'ensemble
Outil en ligne de commande (CLI) pour initialiser un environnement de développement agentic robuste sur Windows, 
avec dépendance Docker ou Podman.
L'outil configure l'outillage agentic, l'IDE, les skills/sous-agents, l'intégration continue locale,
le déploiement d'une préproduction locale via docker-compose, et un système de validation de l'environnement.

## Public cible
- Principalement l'utilisateur, adapté à son workflow
- Développeurs augmentés par IA
- Vibecodeurs adoptant des pratiques de développement robustes

## Fonctionnalités

### Initialisation de l'environnement agentic
- Configuration des outils agentic : OpenCode et Cursor
- Installation et configuration des skills/sous-agents

### Configuration IDE
- Configuration des IDE JetBrains : IntelliJ IDEA et GoLand

### Génération de fichiers de configuration locaux
- `.gitignore`
- Configuration IDE
- Skills
- `docker-compose.yml`
- Serveurs MCP
- Workflow de développement (ordre des skills, etc.)
- D'autres fichiers pourront être ajoutés au fur et à mesure

### Intégration continue locale
- Pipeline CI locale à déterminer selon analyse de l'existant
- Étapes attendues :
  1. Build du projet
  2. Exécution des tests unitaires et d'intégration
  3. Déploiement d'un environnement de test
  4. Lancement des tests e2e
  5. Déploiement d'un environnement de préproduction

### Déploiement de préproduction locale
- Déploiement via docker-compose avec tous les services nécessaires à l'environnement de développement
- Création de conteneurs à la volée selon les besoins du workflow

### Validation de l'environnement
- Système modulaire : la validation dépend de l'application développée
- Modules prévus pour environnements supportés : Go, Quarkus, etc.

### Connecteurs pour services tiers
- Architecture modulaire : chaque service tiers est un module/plugin dédié
- Conteneur Docker dédié exposant une API (REST ou gRPC)
- Le CLI interagit avec ces conteneurs pour offrir les services

## Contraintes techniques
- **Windows uniquement** (V1), avec ouverture possible vers d'autres OS à terme
- **Local first** — aucune dépendance cloud obligatoire
- **Dépendance conteneurs** — Docker ou Podman fonctionnel requis
- **Configuration** — via fichier de configuration local au format **YAML**

## Architecture

### CLI
- Binaire unique Go pour Windows
- Le CLI intègre à la fois la génération de fichiers de configuration et l'orchestration de conteneurs Docker
- Des conteneurs Docker "plugins" peuvent ajouter des fonctionnalités supplémentaires
- Le CLI pourrait n'être qu'une coquille vide qui se connecte à des conteneurs Docker locaux pour les 
  fonctionnalités et commandes

### Sous-commandes
- `init`
  - `specs` — initialisation des spécifications
  - `ci` — initialisation de l'intégration continue
- D'autres sous-commandes à définir (`validate`, `deploy`, `update`, etc.)
