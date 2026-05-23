# Story #003 : Configuration des outils agentic et IDE

## Description
En tant que développeur, je veux configurer automatiquement les outils agentic (OpenCode, Cursor) et les IDE JetBrains (IntelliJ IDEA, GoLand) via la CLI afin d'avoir un environnement de développement agentic prêt à l'emploi.

## Critères d'acceptation
- [ ] `ade init` détecte OpenCode et Cursor via les chemins par défaut Windows + PATH
- [ ] Les chemins des outils peuvent être surchargés dans la configuration YAML
- [ ] `ade init` installe ou référence les skills disponibles dans `.opencode/skills/`
- [ ] `ade init` génère la configuration IDE (fichiers `.idea/`, etc.)
- [ ] `ade init` configure les serveurs MCP définis dans la configuration du conteneur
- [ ] Un message clair indique les outils manquants avec instructions d'installation

## Tests automatisés
- Test unitaire : Détection simulée d'outils installés/non installés (chemins par défaut + PATH)
- Test d'intégration : Génération de la config IDE dans un répertoire temporaire
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
