# Story #008 : Conteneur de configuration (orchestrateur) avec interface web

## Description
En tant que développeur, je veux disposer d'un conteneur Docker orchestrateur qui coordonne les plugins indépendants, expose une API (REST + gRPC) et une interface web (frontend React/Vue) pour gérer les projets, plugins et workflows.

## Critères d'acceptation
- [ ] L'orchestrateur est un conteneur Docker léger qui découvre et coordonne les plugins
- [ ] L'orchestrateur expose une API (REST et gRPC) pour le CLI et les plugins
- [ ] Le CLI détecte si l'orchestrateur est en cours d'exécution et le démarre si nécessaire
- [ ] Le CLI récupère la configuration depuis l'orchestrateur
- [ ] L'interface web est un frontend séparé (React ou Vue) dans le même conteneur
- [ ] L'interface web permet de gérer les projets (création, modification, suppression)
- [ ] L'interface web permet de gérer les plugins (découverte, installation, activation)
- [ ] L'interface web permet de visualiser les workflows et leur historique d'exécution
- [ ] L'interface web permet de voir les rapports de validation et de CI
- [ ] Le conteneur est défini dans le `docker-compose.yml` généré par `ade init ci`
- [ ] L'orchestrateur reste fonctionnel même sans plugins

## Tests automatisés
- Test unitaire : API de l'orchestrateur (REST + gRPC)
- Test d'intégration : Communication CLI ↔ orchestrateur, orchestrateur ↔ plugin
- Test E2E : Démarrage de l'orchestrateur, configuration via web UI, découverte de plugin, récupération par le CLI

## Documentation
- `docs/orchestrator/architecture.md` — Architecture de l'orchestrateur
- `docs/orchestrator/api.md` — API REST et gRPC
- `docs/orchestrator/web-ui.md` — Guide d'utilisation de l'interface web
- `docs/orchestrator/discovery.md` — Découverte et enregistrement des plugins

## Valeur utilisateur
Centralise la configuration et la coordination via un orchestrateur léger, avec une interface web pour une gestion visuelle des projets, plugins et workflows, sans édition manuelle de fichiers.

## Dépendances
- Story #004
