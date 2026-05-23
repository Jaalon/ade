# Story #002 : Génération des fichiers de configuration

## Description
En tant que développeur, je veux exécuter `ade init specs` afin de générer les fichiers de configuration locaux (`.gitignore`, config IDE, skills, serveurs MCP, workflow de développement) depuis des templates fournis par les plugins.

## Critères d'acceptation
- [ ] `ade init specs` interroge les plugins disponibles pour récupérer les templates
- [ ] `ade init specs` génère un fichier `.gitignore` adapté à un projet Go
- [ ] `ade init specs` génère la configuration IDE pour IntelliJ IDEA et GoLand
- [ ] `ade init specs` génère la structure de skills et serveurs MCP
- [ ] `ade init specs` génère un fichier de workflow de développement décrivant l'ordre des skills
- [ ] Les templates sont fournis via l'API des plugins Docker (REST/gRPC)
- [ ] La commande n'écrase pas les fichiers existants sans confirmation

## Tests automatisés
- Test unitaire : Génération de chaque fichier de config individuellement
- Test d'intégration : Exécution de `ade init specs` dans un répertoire temporaire avec un plugin mock fournissant des templates
- Test E2E : Exécution complète avec plugin Docker réel, modification des fichiers, puis ré-exécution (vérification non-écrasement)

## Documentation
- `docs/commands/init-specs.md` — Documentation de la sous-commande
- `docs/plugins/templates.md` — Comment les plugins fournissent des templates

## Valeur utilisateur
Évite la configuration manuelle répétitive et garantit des fichiers de configuration cohérents, avec des templates maintenus par les plugins de l'écosystème.

## Dépendances
- Story #001
- Story #004 (pour le conteneur de configuration et les plugins Docker)
