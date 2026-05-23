# Story #003 : Configuration des outils agentic

## Description
En tant que développeur, je veux configurer automatiquement les outils agentic (OpenCode, Cursor) via la CLI afin d'avoir un environnement de développement agentic prêt à l'emploi.

## Critères d'acceptation
- [ ] `ade init` détecte OpenCode et Cursor via les chemins par défaut Windows + PATH
- [ ] Les chemins des outils peuvent être surchargés dans la configuration YAML
- [ ] `ade init` installe ou référence les skills disponibles dans `.opencode/skills/`
- [ ] `ade init` configure les serveurs MCP définis dans la configuration du conteneur
- [ ] Un message clair indique les outils manquants avec instructions d'installation

## Tests automatisés
- Test unitaire : Détection simulée d'outils installés/non installés (chemins par défaut + PATH)
- Test d'intégration : Exécution de `ade init` dans un répertoire temporaire et vérification skills/MCP
- Test E2E : Exécution sur un système avec OpenCode installé et vérification de la configuration

## Documentation
- `docs/agentic/setup.md` — Guide de configuration des outils agentic
- `docs/agentic/skills.md` — Liste et description des skills installés
- `docs/configuration/yaml.md` — Configuration des chemins d'outils dans le YAML

## Valeur utilisateur
Permet de passer de zéro à un environnement agentic complet en une seule commande, avec détection automatique et surcharge possible via configuration.

## Dépendances
- Story #001
- Story #002
